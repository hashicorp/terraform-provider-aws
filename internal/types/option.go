package types

type Option[T any] []T

const (
	value = iota
)

// Some returns an Option containing a value.
func Some[T any](v T) Option[T] {
	return Option[T]{
		value: v,
	}
}

// None returns an Option with no value.
func None[T any]() Option[T] {
	return nil
}

// IsNone returns whether the Option has no value.
func (o Option[T]) IsNone() bool {
	return o == nil
}

// IsSome returns whether the Option has a value.
func (o Option[T]) IsSome() bool {
	return o != nil
}

// UnwrapOr returns the contained value or the specified default.
func (o Option[T]) UnwrapOr(v T) T {
	if o.IsNone() {
		return v
	}
	return o[value]
}

// UnwrapOrDefault returns the contained value or the default value for T.
func (o Option[T]) UnwrapOrDefault() T {
	if o.IsNone() {
		var v T
		return v
	}
	return o[value]
}

// UnwrapOrElse returns the contained value or computes a value from f.
func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.IsNone() {
		return f()
	}
	return o[value]
}
