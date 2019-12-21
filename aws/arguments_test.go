package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestDiffArguments(t *testing.T) {
	tests := []struct {
		name             string
		oldArgs, newArgs map[string]interface{}
		expectedCreate   map[string]string
		expectedRemove   []string
	}{
		{
			name:           "empty old and new arguments",
			oldArgs:        map[string]interface{}{},
			newArgs:        map[string]interface{}{},
			expectedCreate: map[string]string{},
			expectedRemove: []string{},
		},
		{
			name: "old and new arguments are the same",
			oldArgs: map[string]interface{}{
				"foo": "foo-val",
				"bar": "bar-val",
			},
			newArgs: map[string]interface{}{
				"foo": "foo-val",
				"bar": "bar-val",
			},
			expectedCreate: map[string]string{},
			expectedRemove: []string{},
		},
		{
			name:    "same argument with new value created",
			oldArgs: map[string]interface{}{},
			newArgs: map[string]interface{}{
				"foo": "foo-val",
			},
			expectedCreate: map[string]string{
				"foo": "foo-val",
			},
			expectedRemove: []string{},
		},
		{
			name: "old argument deleted",
			oldArgs: map[string]interface{}{
				"foo": "foo-val",
			},
			newArgs:        map[string]interface{}{},
			expectedCreate: map[string]string{},
			expectedRemove: []string{"foo"},
		},
		{
			name: "some old and new arguments overlap",
			oldArgs: map[string]interface{}{
				"foo": "foo-val",
				"bar": "bar-val",
			},
			newArgs: map[string]interface{}{
				"foo": "foo-val",
				"bar": "baz-val",
			},
			expectedCreate: map[string]string{
				"bar": "baz-val",
			},
			expectedRemove: []string{"bar"},
		},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			create, remove := diffArguments(tc.oldArgs, tc.newArgs)
			cr := aws.StringValueMap(create)
			rm := aws.StringValueSlice(remove)

			if !reflect.DeepEqual(cr, tc.expectedCreate) {
				t.Fatalf("%d: bad create: %#v", i, cr)
			}
			if !reflect.DeepEqual(rm, tc.expectedRemove) {
				t.Fatalf("%d: bad remove: %#v", i, rm)
			}

		})
	}
}
