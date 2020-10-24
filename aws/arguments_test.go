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
				"arg1": "value1",
				"arg2": "value2",
			},
			newArgs: map[string]interface{}{
				"arg1": "value1",
				"arg2": "value2",
			},
			expectedCreate: map[string]string{},
			expectedRemove: []string{},
		},
		{
			name: "argument updated",
			oldArgs: map[string]interface{}{
				"arg1": "value1",
				"arg2": "value2",
			},
			newArgs: map[string]interface{}{
				"arg1": "value1updated",
				"arg2": "value2",
			},
			expectedCreate: map[string]string{
				"arg1": "value1updated",
			},
			expectedRemove: []string{"arg1"},
		},
		{
			name: "argument added",
			oldArgs: map[string]interface{}{
				"arg1": "value1",
			},
			newArgs: map[string]interface{}{
				"arg1": "value1",
				"arg2": "value2",
			},
			expectedCreate: map[string]string{
				"arg2": "value2",
			},
			expectedRemove: []string{},
		},
		{
			name: "argument deleted",
			oldArgs: map[string]interface{}{
				"arg1": "value1",
				"arg2": "value2",
			},
			newArgs: map[string]interface{}{
				"arg2": "value2",
			},
			expectedCreate: map[string]string{},
			expectedRemove: []string{"arg1"},
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
