package portal

import (
	"core/invoke"
	"core/variant"
	"reflect"
	"strings"
)

type PortalActor[T any] struct {
	Base Portal
	Id   []string
}

type ActorPortal[T any] interface {
	Obtain(id []string) PortalActor[T]
}

type Portal interface {
	Invoker() *invoke.DynamicInvoker
}

type SimpleActorPortal[T any] struct {
	base Portal
}

type SimplePortal struct {
	invoker *invoke.DynamicInvoker
}

func (portal SimplePortal) Invoker() *invoke.DynamicInvoker {
	return portal.invoker
}

func New(invoker *invoke.DynamicInvoker) Portal {
	return SimplePortal{
		invoker: invoker,
	}
}

func ForType[T any](base Portal) ActorPortal[T] {
	p := SimpleActorPortal[T]{
		base: base,
	}
	return p
}

func (portal SimpleActorPortal[T]) Obtain(id []string) PortalActor[T] {
	return PortalActor[T]{
		Base: portal.base,
		Id:   id,
	}
}

func (rec PortalActor[T]) Invoke(input any) error {
	var a T
	var tp = reflect.TypeOf(a)
	var inputtp = reflect.ValueOf(input).Elem().Type()
	var inputtps = strings.Split(inputtp.Name(), "_")
	var action = inputtps[len(inputtps)-1]

	req := invoke.InvokeRequest{
		ActorType:  strings.ToLower(tp.Name()),
		ActorId:    rec.Id,
		ActionName: strings.ToLower(action),
	}
	a0 := reflect.ValueOf(input).Elem().FieldByName("A0")
	a0value := a0.Interface()
	req.Parameters = []any{a0value}
	resp, err := rec.Base.Invoker().Invoke(&req)
	if err != nil {
		return err
	}
	res := reflect.ValueOf(input)
	res = res.Elem().FieldByName("Result")
	return variant.Assign(resp, &res)
}

func (rec PortalActor[T]) Call() T {
	var t T
	return t
}
