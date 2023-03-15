package flex

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestExpandStringList(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestExpandResourceId(t *testing.T) {
	t.Parallel()

	resourceId := "foo,bar,baz"
	expandedId, _ := ExpandResourceId(resourceId, 3)
	expected := []string{
		"foo",
		"bar",
		"baz",
	}

	if !reflect.DeepEqual(expandedId, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			expandedId,
			expected)
	}
}

func TestExpandResourceIdEmptyPart(t *testing.T) {
	t.Parallel()

	resourceId := "foo,,baz"
	_, err := ExpandResourceId(resourceId, 3)

	if !strings.Contains(err.Error(), "format for ID (foo,,baz), the following id parts indexes are blank ([1])") {
		t.Fatalf("Expected an error when parsing ResourceId with an empty part")
	}
}

func TestExpandResourceIdIncorrectPartCount(t *testing.T) {
	t.Parallel()

	resourceId := "foo,bar,baz"
	_, err := ExpandResourceId(resourceId, 2)

	if !strings.Contains(err.Error(), "unexpected format for ID (foo,bar,baz), expected (2) parts separated by (,)") {
		t.Fatalf("Expected an error when parsing ResourceId with incorrect part count")
	}
}

func TestExpandResourceIdSinglePart(t *testing.T) {
	t.Parallel()

	resourceId := "foo"
	_, err := ExpandResourceId(resourceId, 2)

	if !strings.Contains(err.Error(), "unexpected format for ID ([foo]), expected more than one part") {
		t.Fatalf("Expected an error when parsing ResourceId with single part count")
	}
}

func TestFlattenResourceId(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "bar", "baz"}
	flattenedId, _ := FlattenResourceId(idParts, 3)
	expected := "foo,bar,baz"

	if !reflect.DeepEqual(flattenedId, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			flattenedId,
			expected)
	}
}

func TestFlattenResourceIdEmptyPart(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "", "baz"}
	_, err := FlattenResourceId(idParts, 3)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo  baz]), the following id parts indexes are blank ([1])") {
		t.Fatalf("Expected an error when parsing ResourceId with an empty part")
	}
}

func TestFlattenResourceIdIncorrectPartCount(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "bar", "baz"}
	_, err := FlattenResourceId(idParts, 2)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo bar baz]), expected (2) parts") {
		t.Fatalf("Expected an error when parsing ResourceId with incorrect part count")
	}
}

func TestFlattenResourceIdSinglePart(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo"}
	_, err := FlattenResourceId(idParts, 2)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo]), expected more than one part") {
		t.Fatalf("Expected an error when parsing ResourceId with single part count")
	}
}
