package variant

type Variant interface {
	Get(template any) (any, error)
}

type baseVariant struct {
	value any
}

func (variant *baseVariant) Get(value any) (any, error) {
	return variant.value, nil
}

func New(value any) Variant {
	if value == nil {
		return nil
	}

	variant := baseVariant{
		value: value,
	}
	return &variant
}
