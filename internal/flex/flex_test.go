package flex

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestExpandStringList(t *testing.T) {
	expanded := []interface{}{"us-east-1a", "us-east-1b"} //lintignore:AWSAT003
	stringList := ExpandStringList(expanded)
	expected := []*string{
		aws.String("us-east-1a"), //lintignore:AWSAT003
		aws.String("us-east-1b"), //lintignore:AWSAT003
	}

	if !reflect.DeepEqual(stringList, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			stringList,
			expected)
	}
}

func TestExpandStringListEmptyItems(t *testing.T) {
	expanded := []interface{}{"foo", "bar", "", "baz"}
	stringList := ExpandStringList(expanded)
	expected := []*string{
		aws.String("foo"),
		aws.String("bar"),
		aws.String("baz"),
	}

	if !reflect.DeepEqual(stringList, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			stringList,
			expected)
	}
}
