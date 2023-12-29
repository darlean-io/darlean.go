/*
Package portal exposes the types to create and use portals that can be used to retrieve and invoke remote actors.

# Defining actor and action signatures

Before you can invoke a remote actor, you must define its actor signature. See [portal.ActorSignature] for more info.

As an example, let us consider a `FriendlyActor` actor that has two actions:
  - `Echo` receives an input string, and simply returns the exact same input string
  - `Greet` receives whom to greet (a string) and the number of times to greet (a number),
    and returns a string like `Friendly actor brings <times> greetings to <whom>`
  - `GetHistory` returns a struct with the distinct whom's and the distinct times values.

The action signatures for the `Echo` and `Greet` actions are:

	// Definition of `Echo` action of actor type `FriendlyActor`.
	type FriendlyActor_Echo struct {
		A0_InputString  string
		Result          string
	}

	// Definition of `Greet` action of actor type `FriendlyActor`.
	type FriendlyActor_Greet struct {
		A0_Whom  string
		A1_Times int
		Result   string
	}

	// Helper struct
	type History struct {
		Whoms []string
		Times []int
	}

	// Definition of `GetHistory` action of actor type `FriendlyActor`.
	type FriendlyActor_GetHistory struct {
		Result History
	}

The actor signature for the FriendlyActor becomes:

	type FriendlyActor struct {
		Echo  FriendlyActor_Echo
		Greet FriendlyActor_Greet
	}

# Obtaining a portal

To invoke an actor, we need a portal. The simplest way to obtain a portal is by wrapping one
around an [invoker.Invoker] instance:

	p := portal.New(invoker)

This portal is generic. That is, it can be used to invoke any type of actor. In our example, we
want to invoke the FriendlyActor. Therefore, we have to obtain a typed portal that is specific
to our FriendlyActor:

	friendlyActorPortal := typedportal.ForSignature[FriendlyActor](p)

# Invoking the actor

Now that we have the typed portal, we can invoke the actor.

	actorId := []string{"friendlyActor1"}
	actor := friendlyActorPortal.Obtain(actorId)

	// Obtain and fill in a new call structure (of type `FriendlyActor_Greet`)
	call := actor.NewCall().Greet
	call.A0_Whom = "World"
	call.A1_Times = 42

	// Invoke the action
	err = actor.Invoke(&call)
	if err != nil {
		panic(err)
	}

	// The result is now present as string in call.Result
	fmt.Printf("Received via Portal: %v\n", call.Result)
	// Prints: Received via Portal: Friendly actor brings 42 greetings to World

When we want to make another call to the same actor, we can reuse the `actor` variable. We just have
to create and invoke a new call:

	call2 := actor.NewCall().Greet
	call2.A0_Whom = "Moon"
	call2.A1_Times = 12
	err = actor.Invoke(&call2)

Portals also support structured data. For example, the GetHistory action returns a struct with the
distinct whom names and the distinct times numbers:

	call3 := actor.NewCall().GetHistory
	err = actor.Invoke(&call3)
	// Omit error checking for now

	fmt.Printf("Distinct whoms: %v", call3.Result.History.Whoms)
	// Prints: Distinct whoms: World Moon

	fmt.Printf("Distinct times: %v", call3.Result.History.Times)
	// Prints: Distinct times: 42 12
*/
package portal
