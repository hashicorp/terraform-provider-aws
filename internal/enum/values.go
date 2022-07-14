package enum

type valueser[T ~string] interface {
	~string
	Values() []T
}

func Values[T valueser[T]]() []string {
	l := T("").Values()

	result := make([]string, len(l))
	for i, v := range l {
		result[i] = string(v)
	}

	return result
}
