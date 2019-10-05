package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestDiffQLDBTags(t *testing.T) {
	cases := []struct {
		Old, New map[string]interface{}
		Create   map[string]*string
		Remove   []*string
	}{
		// Basic add/remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"bar": "baz",
			},
			Create: map[string]*string{
				"bar": aws.String("baz"),
			},
			Remove: []*string{
				aws.String("foo"),
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]*string{
				"foo": aws.String("baz"),
			},
			Remove: []*string{
				aws.String("foo"),
			},
		},
	}

	for i, tc := range cases {
		cm := tagsFromMapQLDBCreate(tc.New)
		rm := tagsFromMapQLDBRemove(tc.Old)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}
