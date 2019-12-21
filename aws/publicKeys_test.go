package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestDiffPublicKeys(t *testing.T) {
	tests := []struct {
		name                           string
		oldKeys, newKeys               []*string
		expectedCreate, expectedRemove []string
	}{
		{
			name:           "empty old and new keys",
			oldKeys:        []*string{},
			newKeys:        []*string{},
			expectedCreate: []string{},
			expectedRemove: []string{},
		},
		{
			name:           "old and new keys are the same",
			oldKeys:        []*string{aws.String("foo"), aws.String("bar")},
			newKeys:        []*string{aws.String("foo"), aws.String("bar")},
			expectedCreate: []string{},
			expectedRemove: []string{},
		},
		{
			name:           "new key created",
			oldKeys:        []*string{},
			newKeys:        []*string{aws.String("foo")},
			expectedCreate: []string{"foo"},
			expectedRemove: []string{},
		},
		{
			name:           "old key deleted",
			oldKeys:        []*string{aws.String("foo")},
			newKeys:        []*string{},
			expectedCreate: []string{},
			expectedRemove: []string{"foo"},
		},
		{
			name:           "some old and new keys overlap",
			oldKeys:        []*string{aws.String("foo"), aws.String("bar")},
			newKeys:        []*string{aws.String("bar"), aws.String("baz")},
			expectedCreate: []string{"baz"},
			expectedRemove: []string{"foo"},
		},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			create, remove := diffPublicKeys(tc.oldKeys, tc.newKeys)
			cr := aws.StringValueSlice(create)
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
