package tags

type valuer interface {
	ValueString() string
}

type TagMap[V valuer] map[string]V

func (tm TagMap[V]) Difference(in TagMap[V]) TagMap[V] {
	result := make(map[string]V)

	for _, v := range tm {
		v.ValueString()
	}
	return result
}
