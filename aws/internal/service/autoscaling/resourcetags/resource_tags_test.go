package resourcetags

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestResourceTagsNew(t *testing.T) {
	testCases := []struct {
		name string
		data interface{}
		fail bool
		want []string
	}{
		{
			name: "empty_list",
			data: []interface{}{},
			want: []string{},
		},
		{
			name: "empty_set",
			data: schema.NewSet(testResourceTagsHashSet, []interface{}{}),
			want: []string{},
		},
		{
			name: "invalid_type",
			data: 42,
			fail: true,
		},
		{
			name: "list_with_empty_map",
			data: []interface{}{map[string]interface{}{}},
			fail: true,
		},
		{
			name: "set_with_empty_map",
			data: schema.NewSet(testResourceTagsHashSet, []interface{}{map[string]interface{}{}}),
			fail: true,
		},
		{
			name: "single_list",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": true,
				},
			},
			want: []string{"key1"},
		},
		{
			name: "single_set",
			data: schema.NewSet(testResourceTagsHashSet, []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": "true",
				},
			}),
			want: []string{"key1"},
		},
		{
			name: "multi_list",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": true,
				},
				map[string]interface{}{
					"key":                 "key2",
					"value":               "value2",
					"propagate_at_launch": false,
				},
				map[string]interface{}{
					"key":                 "key3",
					"value":               "value3",
					"propagate_at_launch": true,
				},
			},
			want: []string{"key1", "key2", "key3"},
		},
		{
			name: "multi_set",
			data: schema.NewSet(testResourceTagsHashSet, []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": "true",
				},
				map[string]interface{}{
					"key":                 "key2",
					"value":               "value2",
					"propagate_at_launch": "false",
				},
				map[string]interface{}{
					"key":                 "key3",
					"value":               "value3",
					"propagate_at_launch": "true",
				},
			}),
			want: []string{"key1", "key2", "key3"},
		},
		{
			name: "missing_key",
			data: []interface{}{
				map[string]interface{}{
					"value":               "value1",
					"propagate_at_launch": true,
				},
			},
			fail: true,
		},
		{
			name: "missing_value",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"propagate_at_launch": true,
				},
			},
			fail: true,
		},
		{
			name: "missing_propagate_at_launch",
			data: []interface{}{
				map[string]interface{}{
					"key":   "key1",
					"value": "value1",
				},
			},
			fail: true,
		},
		{
			name: "empty_key",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "",
					"value":               "value1",
					"propagate_at_launch": true,
				},
			},
			fail: true,
		},
		{
			name: "empty_value",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "",
					"propagate_at_launch": true,
				},
			},
			want: []string{"key1"},
		},
		{
			name: "invalid_key_type",
			data: []interface{}{
				map[string]interface{}{
					"key":                 42,
					"value":               "value1",
					"propagate_at_launch": true,
				},
			},
			fail: true,
		},
		{
			name: "invalid_value_type",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               42,
					"propagate_at_launch": true,
				},
			},
			fail: true,
		},
		{
			name: "invalid_propagate_at_launch_type",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": 42,
				},
			},
			fail: true,
		},
		{
			name: "invalid_propagate_at_launch_boolean",
			data: []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": "nein",
				},
			},
			fail: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := New(testCase.data)
			if err != nil {
				if !testCase.fail {
					t.Errorf("unexpected failure: %s", err)
				}
				return
			}
			if testCase.fail {
				t.Errorf("unexpected success")
			}

			testResourceTagsVerifyKeys(t, got.Keys(), testCase.want)
		})
	}
}

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
		{
			name: "all",
			tags: testResourceTagsNew(t, []interface{}{
				map[string]interface{}{
					"key":                 "aws:cloudformation:key1",
					"value":               "value1",
					"propagate_at_launch": true,
				},
				map[string]interface{}{
					"key":                 "aws:cloudformation:key2",
					"value":               "value2",
					"propagate_at_launch": false,
				},
				map[string]interface{}{
					"key":                 "aws:cloudformation:key3",
					"value":               "value3",
					"propagate_at_launch": true,
				},
			}),
			want: []string{},
		},
		{
			name: "mixed",
			tags: testResourceTagsNew(t, []interface{}{
				map[string]interface{}{
					"key":                 "aws:cloudformation:key1",
					"value":               "value1",
					"propagate_at_launch": true,
				},
				map[string]interface{}{
					"key":                 "key2",
					"value":               "value2",
					"propagate_at_launch": false,
				},
				map[string]interface{}{
					"key":                 "key3",
					"value":               "value3",
					"propagate_at_launch": true,
				},
			}),
			want: []string{"key2", "key3"},
		},
		{
			name: "all",
			tags: testResourceTagsNew(t, []interface{}{
				map[string]interface{}{
					"key":                 "key1",
					"value":               "value1",
					"propagate_at_launch": true,
				},
				map[string]interface{}{
					"key":                 "key2",
					"value":               "value2",
					"propagate_at_launch": false,
				},
				map[string]interface{}{
					"key":                 "key3",
					"value":               "value3",
					"propagate_at_launch": true,
				},
			}),
			want: []string{"key1", "key2", "key3"},
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

func testResourceTagsHashSet(v interface{}) int {
	m := v.(map[string]interface{})
	key, ok := m["key"].(string)
	if !ok {
		return 0
	}
	return hashcode.String(key)
}
