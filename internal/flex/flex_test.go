// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func TestExpandStringList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		configured []interface{}
		want       []*string
	}{
		{
			configured: []interface{}{"abc", "xyz123"},
			want:       []*string{aws.String("abc"), aws.String("xyz123")},
		},
		{
			configured: []interface{}{"abc", 123, "xyz123"},
			want:       []*string{aws.String("abc"), aws.String("xyz123")},
		},
		{
			configured: []interface{}{"foo", "bar", "", "baz"},
			want:       []*string{aws.String("foo"), aws.String("bar"), aws.String("baz")},
		},
	}
	for _, testCase := range testCases {
		if got, want := ExpandStringList(testCase.configured), testCase.want; !cmp.Equal(got, want) {
			t.Errorf("ExpandStringList(%v) = %v, want %v", testCase.configured, got, want)
		}
	}
}

func TestExpandStringListEmpty(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		configured []interface{}
		want       []*string
	}{
		{
			configured: []interface{}{"abc", "xyz123"},
			want:       []*string{aws.String("abc"), aws.String("xyz123")},
		},
		{
			configured: []interface{}{"abc", 123, "xyz123"},
			want:       []*string{aws.String("abc"), aws.String(""), aws.String("xyz123")},
		},
		{
			configured: []interface{}{"foo", "bar", "", "baz"},
			want:       []*string{aws.String("foo"), aws.String("bar"), aws.String(""), aws.String("baz")},
		},
		{
			configured: []interface{}{"foo", "bar", nil, "baz"},
			want:       []*string{aws.String("foo"), aws.String("bar"), aws.String(""), aws.String("baz")},
		},
	}
	for _, testCase := range testCases {
		if got, want := ExpandStringListEmpty(testCase.configured), testCase.want; !cmp.Equal(got, want) {
			t.Errorf("ExpandStringListEmpty(%v) = %v, want %v", testCase.configured, got, want)
		}
	}
}

func TestExpandStringTimeList(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"2006-01-02T15:04:05+07:00", "2023-04-13T10:25:05+01:00"}
	got := ExpandStringTimeList(configured, time.RFC3339)
	want := []*time.Time{
		aws.Time(time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("UTC-7", 7*60*60))),
		aws.Time(time.Date(2023, 4, 13, 10, 25, 5, 0, time.FixedZone("UTC-1", 60*60))),
	}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandStringTimeListEmptyItems(t *testing.T) {
	t.Parallel()

	configured := []interface{}{"2006-01-02T15:04:05+07:00", "", "2023-04-13T10:25:05+01:00"}
	got := ExpandStringTimeList(configured, time.RFC3339)
	want := []*time.Time{
		aws.Time(time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("UTC+7", 7*60*60))),
		aws.Time(time.Date(2023, 4, 13, 10, 25, 5, 0, time.FixedZone("UTC+1", 60*60))),
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
	got, _ := ExpandResourceId(id, 3, false)
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
	_, err := ExpandResourceId(resourceId, 3, false)

	if !strings.Contains(err.Error(), "format for ID (foo,,baz), the following id parts indexes are blank ([1])") {
		t.Fatalf("Expected an error when parsing ResourceId with an empty part")
	}
}

func TestExpandResourceIdAllowEmptyPart(t *testing.T) {
	t.Parallel()

	resourceId := "foo,,baz"
	got, _ := ExpandResourceId(resourceId, 3, true)

	want := []string{
		"foo",
		"",
		"baz",
	}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestExpandResourceIdIncorrectPartCount(t *testing.T) {
	t.Parallel()

	resourceId := "foo,bar,baz"
	_, err := ExpandResourceId(resourceId, 2, false)

	if !strings.Contains(err.Error(), "unexpected format for ID (foo,bar,baz), expected (2) parts separated by (,)") {
		t.Fatalf("Expected an error when parsing ResourceId with incorrect part count")
	}
}

func TestExpandResourceIdSinglePart(t *testing.T) {
	t.Parallel()

	resourceId := "foo"
	_, err := ExpandResourceId(resourceId, 2, false)

	if !strings.Contains(err.Error(), "unexpected format for ID ([foo]), expected more than one part") {
		t.Fatalf("Expected an error when parsing ResourceId with single part count")
	}
}

func TestFlattenResourceId(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "bar", "baz"}
	got, _ := FlattenResourceId(idParts, 3, false)
	want := "foo,bar,baz"

	if !cmp.Equal(got, want) {
		t.Errorf("flattened = %v, want = %v", got, want)
	}
}

func TestFlattenResourceIdEmptyPart(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "", "baz"}
	_, err := FlattenResourceId(idParts, 3, false)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo  baz]), the following id parts indexes are blank ([1])") {
		t.Fatalf("Expected an error when parsing ResourceId with an empty part")
	}
}

func TestFlattenResourceIdIncorrectPartCount(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo", "bar", "baz"}
	_, err := FlattenResourceId(idParts, 2, false)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo bar baz]), expected (2) parts") {
		t.Fatalf("Expected an error when parsing ResourceId with incorrect part count")
	}
}

func TestFlattenResourceIdSinglePart(t *testing.T) {
	t.Parallel()

	idParts := []string{"foo"}
	_, err := FlattenResourceId(idParts, 2, false)

	if !strings.Contains(err.Error(), "unexpected format for ID parts ([foo]), expected more than one part") {
		t.Fatalf("Expected an error when parsing ResourceId with single part count")
	}
}

func TestResourceIdPartCount(t *testing.T) {
	t.Parallel()

	id := "foo,bar,baz"
	partCount := ResourceIdPartCount(id)
	expectedCount := 3
	if partCount != expectedCount {
		t.Fatalf("Expected part count of %d.", expectedCount)
	}
}

func TestResourceIdPartCountLegacySeparator(t *testing.T) {
	t.Parallel()

	id := "foo_bar_baz"
	partCount := ResourceIdPartCount(id)
	expectedCount := 1
	if partCount != expectedCount {
		t.Fatalf("Expected part count of %d.", expectedCount)
	}
}

func TestFlattenTimeStringList(t *testing.T) {
	t.Parallel()

	configured := []*time.Time{
		aws.Time(time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("UTC-7", 7*60*60))),
		aws.Time(time.Date(2023, 4, 13, 10, 25, 5, 0, time.FixedZone("UTC-1", 60*60))),
	}
	got := FlattenTimeStringList(configured, time.RFC3339)
	want := []interface{}{"2006-01-02T15:04:05+07:00", "2023-04-13T10:25:05+01:00"}

	if !cmp.Equal(got, want) {
		t.Errorf("expanded = %v, want = %v", got, want)
	}
}

func TestDiffStringMaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Old, New                  map[string]interface{}
		Create, Remove, Unchanged map[string]interface{}
	}{
		// Add
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Create: map[string]interface{}{
				"bar": "baz",
			},
			Remove: map[string]interface{}{},
			Unchanged: map[string]interface{}{
				"foo": "bar",
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
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{},
		},

		// Overlap
		{
			Old: map[string]interface{}{
				"foo":   "bar",
				"hello": "world",
			},
			New: map[string]interface{}{
				"foo":   "baz",
				"hello": "world",
			},
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{
				"hello": "world",
			},
		},

		// Remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			New: map[string]interface{}{
				"foo": "bar",
			},
			Create: map[string]interface{}{},
			Remove: map[string]interface{}{
				"bar": "baz",
			},
			Unchanged: map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	for _, tc := range cases {
		c, r, u := DiffStringMaps(tc.Old, tc.New)
		cm := FlattenStringMap(c)
		rm := FlattenStringMap(r)
		um := FlattenStringMap(u)
		if diff := cmp.Diff(cm, tc.Create); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(rm, tc.Remove); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(um, tc.Unchanged); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
	}
}

func TestDiffStringValueMaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Old, New                  map[string]interface{}
		Create, Remove, Unchanged map[string]string
	}{
		// Add
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{},
			Unchanged: map[string]string{
				"foo": "bar",
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
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
			Unchanged: map[string]string{},
		},

		// Overlap
		{
			Old: map[string]interface{}{
				"foo":   "bar",
				"hello": "world",
			},
			New: map[string]interface{}{
				"foo":   "baz",
				"hello": "world",
			},
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
			Unchanged: map[string]string{
				"hello": "world",
			},
		},

		// Remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			New: map[string]interface{}{
				"foo": "bar",
			},
			Create: map[string]string{},
			Remove: map[string]string{
				"bar": "baz",
			},
			Unchanged: map[string]string{
				"foo": "bar",
			},
		},
	}

	for _, tc := range cases {
		c, r, u := DiffStringValueMaps(tc.Old, tc.New)
		if diff := cmp.Diff(c, tc.Create); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(r, tc.Remove); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(u, tc.Unchanged); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
	}
}

func TestDiffSlices(t *testing.T) {
	t.Parallel()

	type x struct {
		A string
		B int
	}

	cases := []struct {
		Old, New                  []x
		Create, Remove, Unchanged []x
	}{
		// Add
		{
			Old:       []x{{A: "foo", B: 1}},
			New:       []x{{A: "foo", B: 1}, {A: "bar", B: 2}},
			Create:    []x{{A: "bar", B: 2}},
			Remove:    []x{},
			Unchanged: []x{{A: "foo", B: 1}},
		},
		// Modify
		{
			Old:       []x{{A: "foo", B: 1}},
			New:       []x{{A: "foo", B: 2}},
			Create:    []x{{A: "foo", B: 2}},
			Remove:    []x{{A: "foo", B: 1}},
			Unchanged: []x{},
		},
		// Overlap
		{
			Old:       []x{{A: "foo", B: 1}, {A: "bar", B: 2}},
			New:       []x{{A: "foo", B: 3}, {A: "bar", B: 2}},
			Create:    []x{{A: "foo", B: 3}},
			Remove:    []x{{A: "foo", B: 1}},
			Unchanged: []x{{A: "bar", B: 2}},
		},
		// Remove
		{
			Old:       []x{{A: "foo", B: 1}, {A: "bar", B: 2}},
			New:       []x{{A: "foo", B: 1}},
			Create:    []x{},
			Remove:    []x{{A: "bar", B: 2}},
			Unchanged: []x{{A: "foo", B: 1}},
		},
	}
	for _, tc := range cases {
		c, r, u := DiffSlices(tc.Old, tc.New, func(x1, x2 x) bool { return x1.A == x2.A && x1.B == x2.B })
		if diff := cmp.Diff(c, tc.Create); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(r, tc.Remove); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(u, tc.Unchanged); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
	}
}
