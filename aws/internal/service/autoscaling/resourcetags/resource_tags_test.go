package resourcetags

import (
	"testing"
)

func TestResourceTagsIgnoreAws(t *testing.T) {
	testCases := []struct {
		name string
		tags ResourceTags
		want []string
	}{
		{
			name: "empty",
			tags: testResourceTagsNew(t, []interface{}{}),
			want: []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.IgnoreAws()

			testResourceTagsVerifyKeys(t, got.Keys(), testCase.want)
		})
	}
}

func testResourceTagsNew(t *testing.T, i interface{}) ResourceTags {
	tags, err := New(i)
	if err != nil {
		t.Errorf("%w", err)
	}

	return tags
}

func testResourceTagsVerifyKeys(t *testing.T, got []string, want []string) {
	for _, g := range got {
		found := false

		for _, w := range want {
			if w == g {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("got extra key: %s", g)
		}
	}

	for _, w := range want {
		found := false

		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("want missing key: %s", w)
		}
	}
}
