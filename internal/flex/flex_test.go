package flex

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func TestExpandStringList(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"abc", "xyz123"}
	got := ExpandStringList(configured)
	want := []*string{
		aws.String("abc"),
		aws.String("xyz123"),
	}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandStringListEmptyItems(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"foo", "bar", "", "baz"}
	got := ExpandStringList(configured)
	want := []*string{
		aws.String("foo"),
		aws.String("bar"),
		aws.String("baz"),
	}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandStringValueList(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"abc", "xyz123"}
	got := ExpandStringValueList(configured)
	want := []string{"abc", "xyz123"}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandStringValueListEmptyItems(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"foo", "bar", "", "baz"}
	got := ExpandStringValueList(configured)
	want := []string{"foo", "bar", "baz"}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandResourceId(t *testing.T) {
	t.Parallel()

	id := "foo,bar,baz"
	got, _ := ExpandResourceId(id, 3)
	want := []string{
		"foo",
		"bar",
		"baz",
	}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
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
	got, _ := FlattenResourceId(idParts, 3)
	want := "foo,bar,baz"

	if !cmp.Equal(got, want) {
		t.Errorf("flattened = %v, want = %v", got, want)
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
