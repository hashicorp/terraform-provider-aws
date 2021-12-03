package verify

import (
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
)

func StringSlicesEqualIgnoreOrder(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	sort.Strings(s1)
	sort.Strings(s2)

	return reflect.DeepEqual(s1, s2)
}

func StringValueSlicesEqualIgnoreOrder(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	return StringSlicesEqualIgnoreOrder(v1, v2)
}

func StringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	return reflect.DeepEqual(s1, s2)
}

func StringValueSlicesEqual(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	return StringSlicesEqual(v1, v2)
}
