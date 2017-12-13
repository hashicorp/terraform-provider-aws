package set

// testRules is a rules implementation that is used for testing. It only
// accepts ints as values, and it has a Hash function that just returns the
// given value modulo 16 so that we can easily and dependably test the
// situation where two non-equivalent values have the same hash value.
type testRules struct{}

func (r testRules) Hash(val interface{}) int {
	return val.(int) % 16
}

func (r testRules) Equivalent(val1 interface{}, val2 interface{}) bool {
	return val1 == val2
}
