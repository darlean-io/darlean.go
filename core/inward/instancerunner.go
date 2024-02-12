package inward

import (
	"fmt"
	"sync"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/core/internal/frameworkerror"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/wire"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type InstanceWrapper interface {
	Create() *actionerror.Error
	Activate() *actionerror.Error
	Deactivate() *actionerror.Error
	Release() *actionerror.Error
	Perform(actionName normalized.ActionName, args []variant.Assignable) (result any, err *actionerror.Error)
}

type ActionLockKind int

const ACTION_LOCK_EXCLUSIVE = ActionLockKind(0)
const ACTION_LOCK_SHARED = ActionLockKind(1)
const ACTION_LOCK_NONE = ActionLockKind(2)

type actionKind int

const action_kind_action = actionKind(0)
const action_kind_activate = actionKind(1)
const action_kind_deactivate = actionKind(2)

type ActionDef struct {
	Locking ActionLockKind
}

type DefaultInstanceRunner struct {
	actorType      normalized.ActorType
	actorId        []string
	wrapper        InstanceWrapper
	requiresLock   bool
	actionDefs     map[normalized.ActionName]ActionDef
	sharedCalls    callQueue
	exclusiveCalls callQueue
	noneCalls      callQueue
	onceLoop       sync.Once
	finishedCalls  chan *callFinishedRec
	queueLock      sync.RWMutex
	running        bool
	onDeactivated  func()
}

const state_created = 0
const state_activating = 1
const state_active = 2
const state_deactivation_wanted = 3
const state_deactivating = 4
const state_deactivated = 5

type FinishedHandler func(result any, err *actionerror.Error)

type callRec struct {
	call       *wire.ActorCallRequestIn
	kind       actionKind
	onFinished FinishedHandler
	def        *ActionDef
}

// Queue to which callRec's can be pushed. The queue maintains the number of in-progress items.
type callQueue struct {
	queue      chan callRec
	inProgress int
}

func (queue *callQueue) push(item callRec) {
	queue.queue <- item
}

func (queue *callQueue) do() {
	queue.inProgress++
}

func (queue *callQueue) done() {
	queue.inProgress--
	if queue.inProgress < 0 {
		panic("instancerunner: callqueue in progress can not be negative")
	}
}

func (queue *callQueue) doing() bool {
	return queue.inProgress > 0
}

type callFinishedRec struct {
	finishedActionKind actionKind
	finishedQueue      *callQueue
	finishedHandler    FinishedHandler
	result             any
	err                *actionerror.Error
}

func newCallQueue() callQueue {
	return callQueue{
		queue: make(chan callRec),
	}

}

func (runner *DefaultInstanceRunner) acquireActorLock() error {
	// TODO
	return nil
}

func (runner *DefaultInstanceRunner) releaseActorLock() error {
	// TODO
	return nil
}

const ERROR_DEACTIVATED = "DEACTIVATED"
const ERROR_UNKNOWN_ACTION = "UNKNOWN_ACTION"

// Invokes a `call`. May block until the call is actually being processed.
func (runner *DefaultInstanceRunner) Invoke(call *wire.ActorCallRequestIn, onFinished FinishedHandler) {
	actionDef, has := runner.actionDefs[normalized.NormalizeActionName(call.ActionName)]
	if !has {
		onFinished(nil, frameworkerror.New(actionerror.Options{
			Code:     ERROR_UNKNOWN_ACTION,
			Template: "Unknown action [Action] on actor [ActorType]",
			Parameters: map[string]any{
				"Action":    call.ActionName,
				"ActorType": call.ActorType,
			},
		}))
		return
	}

	runner.onceLoop.Do(func() {
		runner.running = true
		go runner.loop(onFinished, runner.finishedCalls)
	})

	runner.queueLock.RLock()
	defer runner.queueLock.RUnlock()

	if !runner.running {
		onFinished(nil, frameworkerror.New(actionerror.Options{
			Code:     ERROR_DEACTIVATED,
			Template: "Actor type [ActorType] is deactivated",
			Parameters: map[string]any{
				"ActorType": call.ActorType,
			}}))
		return
	}

	switch actionDef.Locking {
	case ACTION_LOCK_EXCLUSIVE:
		runner.exclusiveCalls.push(callRec{call: call, def: &actionDef, onFinished: onFinished})
	case ACTION_LOCK_SHARED:
		runner.sharedCalls.push(callRec{call: call, def: &actionDef, onFinished: onFinished})
	case ACTION_LOCK_NONE:
		runner.noneCalls.push(callRec{call: call, def: &actionDef, onFinished: onFinished})
	}
}

func (runner *DefaultInstanceRunner) TriggerDeactivate() {
	runner.finishedCalls <- nil
}

func (runner *DefaultInstanceRunner) loop(activationErrorHandler FinishedHandler, finishedCalls chan *callFinishedRec) {
	if runner.onDeactivated != nil {
		defer runner.onDeactivated()
	}

	// Loop internal state
	state := state_created

	// Accessors that return a channel when the state is such that we can process
	// events from that channel
	getExclusiveChannel := func() chan callRec {
		if runner.exclusiveCalls.doing() {
			return nil
		}
		if state == state_active {
			return runner.exclusiveCalls.queue
		}
		return nil
	}

	getNoneChannel := func() chan callRec {
		if state == state_activating || state == state_active || state == state_deactivation_wanted || state == state_deactivating {
			return runner.noneCalls.queue
		}
		return nil
	}

	getSharedChannel := func() chan callRec {
		if runner.exclusiveCalls.doing() {
			return nil
		}
		if len(runner.exclusiveCalls.queue) > 0 {
			return nil
		}
		if state == state_active {
			return runner.sharedCalls.queue
		}
		return nil
	}

	drainQueues := func() {
		for _, queue := range []callQueue{runner.exclusiveCalls, runner.sharedCalls, runner.noneCalls} {
		inner:
			for {
				select {
				case call := <-queue.queue:
					err := frameworkerror.New(actionerror.Options{
						Code:     ERROR_DEACTIVATED,
						Template: "Actor type [call.ActorType] is deactivated",
						Parameters: map[string]any{
							"ActorType": call.call.ActorType,
						}})

					call.onFinished(nil, err)
				default:
					break inner
				}
			}
		}
	}

	// Invoke one specific call and update the administration for the queue accordingly
	invoke := func(call callRec, queue *callQueue) {
		queue.do()
		go func() {
			// Note: This code runs parallel to our loop in a goroutine. It should not modify the state
			// of the runner to avoid race conditions/corruption. The only allowed communication with
			// the loop is by pushing to the finishedCalls channel.
			var result any
			var err *actionerror.Error

			defer func() {
				if r := recover(); r != nil {
					err = actionerror.New(actionerror.Options{
						Code:     "UNEXPECTED_APPLICATION_ERROR",
						Template: "Unexpected application error: [Message]",
						Parameters: map[string]any{
							"Message": fmt.Sprintf("instancerunner: invoke: panic: %v", r),
						},
					})
				}

				finishedCalls <- &callFinishedRec{
					finishedActionKind: call.kind,
					finishedQueue:      queue,
					finishedHandler:    call.onFinished,
					result:             result,
					err:                err,
				}
			}()
			switch call.kind {
			case action_kind_activate:
				err = runner.wrapper.Activate()
			case action_kind_deactivate:
				err = runner.wrapper.Deactivate()
			default:
				result, err = runner.wrapper.Perform(normalized.NormalizeActionName(call.call.ActionName), call.call.Arguments)
			}
		}()
	}

	// Acquire the actor lock
	err := runner.acquireActorLock()
	if err != nil {
		activationErrorHandler(nil, frameworkerror.New(actionerror.Options{
			Code:     "ACTOR_LOCK_FAILED",
			Template: "Unable to obtain actor lock for an instance of [ActorType]: [Reason]",
			Parameters: map[string]any{
				"ActorType": runner.actorType,
				"Reason":    err.Error(),
			}}))
	}

	defer runner.releaseActorLock()

	defer func() {
		// Warning: a read-only queuelock may be held by Invoke. The Invoke waits for us
		// to drain the channel, but we wait to acquire the write lock first. Deadlock.
		// Therefore we do a try lock, and continue processing of the queue channels
		for !runner.queueLock.TryLock() {
			drainQueues()
		}
		runner.running = false
		runner.queueLock.Unlock()
	}()

	// Invoke the "activation" action (under water, it runs in a separate goroutine). It must be able to run in parallel
	// with "none" locking actions. That is why it runs in a goroutine.
	state = state_activating
	invoke(callRec{call: nil, kind: action_kind_activate, onFinished: activationErrorHandler}, &runner.exclusiveCalls)

	for {
		switch state {
		case state_deactivated:
			// Await any pending calls. This must be none calls (other calls should already be done)
			if !runner.noneCalls.doing() {
				return
			}

		case state_deactivation_wanted:
			if !runner.exclusiveCalls.doing() && !runner.sharedCalls.doing() {
				state = state_deactivating
				invoke(callRec{kind: action_kind_deactivate, onFinished: activationErrorHandler}, &runner.exclusiveCalls)
				continue
			}

		}
		select {
		case finished := <-finishedCalls:
			// When nil, it means that deactivation is requested.
			// When assigned, it means that a previous call is finished.

			if finished == nil {
				if state < state_deactivation_wanted {
					state = state_deactivation_wanted
				}
				continue
			}

			finished.finishedQueue.done()
			if finished.finishedActionKind == action_kind_activate {
				if finished.err == nil {
					state = state_active
				} else {
					state = state_deactivation_wanted
					finished.finishedHandler(nil, finished.err)
				}
				// Do not invoke finishedHandler; already done for error, but should not
				// be done on success because activation must be transparent to the caller.
				// We will invoke the finishedHandler on the actual action call that triggered
				// the activation.
				continue
			}
			if finished.finishedActionKind == action_kind_deactivate {
				state = state_deactivated
				// Do not invoke the finished handler, even not in the situation of an error. Deactivating
				// is invisible from the caller.
				continue
			}
			finished.finishedHandler(finished.result, finished.err)
		case call := <-getExclusiveChannel():
			invoke(call, &runner.exclusiveCalls)
		case call := <-getNoneChannel():
			invoke(call, &runner.noneCalls)
		case call := <-getSharedChannel():
			invoke(call, &runner.sharedCalls)
		}
	}

	// When we get here, the instance is deactivated.
	// Just have to clean up some administration, which happens by the above defined "defer" blocks:
	// * The actor lock is released by the "defer" as defined before.
	// * The list of pending calls are informed by the other "defer".
}

func NewInstanceRunner(wrapper InstanceWrapper, actorType normalized.ActorType, actorId []string, requiresLock bool, actionDefs map[normalized.ActionName]ActionDef, onDeactivated func()) *DefaultInstanceRunner {
	runner := DefaultInstanceRunner{
		actorType:      actorType,
		actorId:        actorId,
		wrapper:        wrapper,
		requiresLock:   requiresLock,
		actionDefs:     actionDefs,
		sharedCalls:    newCallQueue(),
		exclusiveCalls: newCallQueue(),
		noneCalls:      newCallQueue(),
		finishedCalls:  make(chan *callFinishedRec),
		onDeactivated:  onDeactivated,
	}
	return &runner
}
