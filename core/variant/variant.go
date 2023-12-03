package variant

type Assignable interface {
	AssignTo(target any) error
}

func Assign(source any, target any) error {
	assignable, supported := source.(Assignable)
	if supported {
		return assignable.AssignTo(target)
	}
	return Decode(source, target)
}
