package anny

type Anny interface {
	Get(template any) (any, error)
}

type baseAnny struct {
	value any
}

func (anny *baseAnny) Get(value any) (any, error) {
	return anny.value, nil
}

func New(value any) Anny {
	if value == nil {
		return nil
	}

	anny := baseAnny{
		value: value,
	}
	return &anny
}
