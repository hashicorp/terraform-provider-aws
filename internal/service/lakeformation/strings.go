package lakeformation

import (
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
)

func StringSlicesEqualIgnoreOrder(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	sort.Strings(v1)
	sort.Strings(v2)

	return reflect.DeepEqual(v1, v2)
}

func StringSlicesEqual(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	return reflect.DeepEqual(v1, v2)
}
