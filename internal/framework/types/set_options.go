package types

type SetNestedObjectOfOption[T any] func(*SetNestedObjectOfOptions[T])

type SetNestedObjectOfOptions[T any] struct {
	SemanticEqualityFunc setSemanticEqualityFunc[T]
}

func SetWithSemanticEqualityFunc[T any](f setSemanticEqualityFunc[T]) SetNestedObjectOfOption[T] {
	return func(o *SetNestedObjectOfOptions[T]) {
		o.SemanticEqualityFunc = f
	}
}

func newSetNestedObjectOfOptions[T any](options ...SetNestedObjectOfOption[T]) *SetNestedObjectOfOptions[T] {
	opts := &SetNestedObjectOfOptions[T]{}

	for _, opt := range options {
		opt(opts)
	}

	return opts
}
