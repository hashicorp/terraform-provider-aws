package flex

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

func TestExpandResourceRegion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		region         interface{}
		defaultRegion  string
		allowedRegions []string
		expectedRegion string
		expectedError  error
	}{
		{
			name:           "resourceRegionPositive",
			region:         "us-east-1",
			defaultRegion:  "us-west-2",
			allowedRegions: []string{"us-west-2", "us-east-1"},
			expectedRegion: "us-east-1",
			expectedError:  nil,
		},
		{
			name:           "resourceRegionNilPositive",
			region:         nil,
			defaultRegion:  "us-west-2",
			allowedRegions: []string{"us-west-2", "us-east-1"},
			expectedRegion: "us-west-2",
			expectedError:  nil,
		},
		{
			name:           "resourceRegionNilNegative",
			region:         nil,
			defaultRegion:  "us-west-2",
			allowedRegions: []string{"us-east-1"},
			expectedRegion: "",
			expectedError:  fmt.Errorf("provided resource region is not an allowed region in the provider configuration. To deploy to this region, add it to the 'allowed_regions' provider setting, or remove the list of 'allowed_regions' from your provider configuration. Provided region us-west-2, provider allowed regions: [us-east-1]"),
		},
		{
			name:           "resourceRegionNegative",
			region:         "us-east-1",
			defaultRegion:  "us-west-2",
			allowedRegions: []string{"us-west-2", "ap-south-1"},
			expectedRegion: "",
			expectedError:  fmt.Errorf("provided resource region is not an allowed region in the provider configuration. To deploy to this region, add it to the 'allowed_regions' provider setting, or remove the list of 'allowed_regions' from your provider configuration. Provided region us-east-1, provider allowed regions: [us-west-2 ap-south-1]"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := ExpandResourceRegion(tc.region, tc.allowedRegions, tc.defaultRegion)

			if tc.expectedError != nil {
				if err != nil && !strings.Contains(err.Error(), tc.expectedError.Error()) {
					t.Fatalf("expected = %s, want = %s", tc.expectedError.Error(), err.Error())
				} else if err != nil && tc.expectedError == nil {
					t.Errorf("unexpected error returned: %s", err)
				}
			}

			if !cmp.Equal(out, tc.expectedRegion) {
				t.Errorf("expanded = %s, want = %s", out, tc.expectedRegion)
			}
		})
	}
}
