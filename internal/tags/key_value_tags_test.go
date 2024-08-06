// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestKeyValueTagsDefaultConfigGetTags(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name          string
		defaultConfig *DefaultConfig
		want          KeyValueTags
	}{
		{
			name:          "empty config",
			defaultConfig: &DefaultConfig{},
			want:          KeyValueTags{},
		},
		{
			name:          "nil config",
			defaultConfig: nil,
			want:          nil,
		},
		{
			name: "with Tags config",
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
			},
			want: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.defaultConfig.GetTags()
			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want.Map())
		})
	}
}

func TestKeyValueTagsDefaultConfigMergeTags(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name          string
		tags          KeyValueTags
		defaultConfig *DefaultConfig
		want          map[string]string
	}{
		{
			name: "empty config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: nil,
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no tags",
			tags: New(ctx, map[string]string{}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys all matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys some matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys some overridden",
			tags: New(ctx, map[string]string{
				"key1": "value2",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
				}),
			},
			want: map[string]string{
				"key1": "value2",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys none matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key4": "value4",
					"key5": "value5",
					"key6": "value6",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.defaultConfig.MergeTags(testCase.tags)
			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsDefaultConfigTagsEqual(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name          string
		tags          KeyValueTags
		defaultConfig *DefaultConfig
		want          bool
	}{
		{
			name: "empty config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{},
			want:          false,
		},
		{
			name: "no config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: nil,
			want:          false,
		},
		{
			name: "empty tags",
			tags: New(ctx, map[string]string{}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: false,
		},
		{
			name: "no tags",
			tags: nil,
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: false,
		},
		{
			name:          "empty config and no tags",
			tags:          nil,
			defaultConfig: &DefaultConfig{},
			want:          true,
		},
		{
			name:          "no config and tags",
			tags:          nil,
			defaultConfig: nil,
			want:          true,
		},
		{
			name: "keys and values all matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: true,
		},
		{
			name: "only keys matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value0",
					"key2": "value1",
					"key3": "value2",
				}),
			},
			want: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.defaultConfig.TagsEqual(testCase.tags)

			if got != testCase.want {
				t.Errorf("got %t; want %t", got, testCase.want)
			}
		})
	}
}

func TestKeyValueTagsIgnoreAWS(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"aws:cloudformation:key2": "value2",
				"aws:cloudformation:key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"key3":                    "value3",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.IgnoreAWS()

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnoreConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name         string
		tags         KeyValueTags
		ignoreConfig *IgnoreConfig
		want         map[string]string
	}{
		{
			name: "empty config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: nil,
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no tags",
			tags: New(ctx, map[string]string{}),
			ignoreConfig: &IgnoreConfig{
				KeyPrefixes: New(ctx, []string{
					"key1",
					"key2",
					"key3",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "keys all matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				Keys: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "keys some matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				Keys: New(ctx, map[string]string{
					"key1": "value1",
				}),
			},
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys none matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				Keys: New(ctx, map[string]string{
					"key4": "value4",
					"key5": "value5",
					"key6": "value6",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys and key prefixes",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				Keys: New(ctx, []string{
					"key1",
				}),
				KeyPrefixes: New(ctx, []string{
					"key2",
				}),
			},
			want: map[string]string{
				"key3": "value3",
			},
		},
		{
			name: "key prefixes all exact",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				KeyPrefixes: New(ctx, []string{
					"key1",
					"key2",
					"key3",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "key prefixes all prefixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				KeyPrefixes: New(ctx, []string{
					"key",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "key prefixes some prefixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				KeyPrefixes: New(ctx, []string{
					"key1",
				}),
			},
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "key prefixes none prefixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreConfig: &IgnoreConfig{
				KeyPrefixes: New(ctx, []string{
					"key4",
					"key5",
					"key6",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.IgnoreConfig(testCase.ignoreConfig)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnoreElasticbeanstalk(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"elasticbeanstalk:key2":   "value2",
				"Name":                    "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"elasticbeanstalk:key3":   "value3",
				"key4":                    "value4",
				"Name":                    "value5",
			}),
			want: map[string]string{
				"key2": "value2",
				"key4": "value4",
			},
		},
		{
			name: "none",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.IgnoreElasticbeanstalk()

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnorePrefixes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name              string
		tags              KeyValueTags
		ignoreTagPrefixes KeyValueTags
		want              map[string]string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			ignoreTagPrefixes: New(ctx, []string{
				"key1",
				"key2",
				"key3",
			}),
			want: map[string]string{},
		},
		{
			name: "all_exact",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New(ctx, []string{
				"key1",
				"key2",
				"key3",
			}),
			want: map[string]string{},
		},
		{
			name: "all_prefix",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New(ctx, []string{
				"key",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New(ctx, []string{
				"key1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New(ctx, []string{
				"key4",
				"key5",
				"key6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.IgnorePrefixes(testCase.ignoreTagPrefixes)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnoreSystem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name        string
		serviceName string
		tags        KeyValueTags
		want        map[string]string
	}{
		{
			name:        "empty",
			serviceName: names.EC2,
			tags:        New(ctx, map[string]string{}),
			want:        map[string]string{},
		},
		{
			name:        "all",
			serviceName: names.ElasticBeanstalk,
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"elasticbeanstalk:key2":   "value2",
			}),
			want: map[string]string{},
		},
		{
			name:        "mixed",
			serviceName: names.S3,
			tags: New(ctx, map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"elasticbeanstalk:key3":   "value3",
				"key4":                    "value4",
			}),
			want: map[string]string{
				"key2":                  "value2",
				"elasticbeanstalk:key3": "value3",
				"key4":                  "value4",
			},
		},
		{
			name:        "none",
			serviceName: names.RDS,
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.IgnoreSystem(testCase.serviceName)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name       string
		tags       KeyValueTags
		ignoreTags KeyValueTags
		want       map[string]string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			ignoreTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(ctx, map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Ignore(testCase.ignoreTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsKeyAdditionalBoolValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name  string
		tags  KeyValueTags
		key   string
		field string
		want  *bool
	}{
		{
			name:  "empty",
			tags:  New(ctx, map[string]*string{}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
		{
			name:  "non-existent key",
			tags:  New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:   "key2",
			field: "field2",
			want:  nil,
		},
		{
			name:  "non-existent TagData",
			tags:  New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
		{
			name: "non-existent field",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
					Value:                testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field2",
			want:  nil,
		},
		{
			name: "matching value",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
					Value:                testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field1",
			want:  testBoolPtr(true),
		},
		{
			name: "matching nil",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalBoolFields: map[string]*bool{"field1": nil},
					Value:                testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.KeyAdditionalBoolValue(testCase.key, testCase.field)

			if testCase.want == nil && got != nil {
				t.Fatalf("expected: nil, got: %#v", got)
			}

			if testCase.want != nil && got == nil {
				t.Fatalf("expected: %#v, got: nil", testCase.want)
			}

			if testCase.want != nil && got != nil && *testCase.want != *got {
				t.Fatalf("expected: %#v, got: %#v", testCase.want, got)
			}
		})
	}
}

func TestKeyValueTagsKeyAdditionalStringValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name  string
		tags  KeyValueTags
		key   string
		field string
		want  *string
	}{
		{
			name:  "empty",
			tags:  New(ctx, map[string]*string{}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
		{
			name:  "non-existent key",
			tags:  New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:   "key2",
			field: "field2",
			want:  nil,
		},
		{
			name:  "non-existent TagData",
			tags:  New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
		{
			name: "non-existent field",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
					Value:                  testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field2",
			want:  nil,
		},
		{
			name: "matching value",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
					Value:                  testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field1",
			want:  testStringPtr("field1value"),
		},
		{
			name: "matching nil",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalStringFields: map[string]*string{"field1": nil},
					Value:                  testStringPtr("value1"),
				},
			}),
			key:   "key1",
			field: "field1",
			want:  nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.KeyAdditionalStringValue(testCase.key, testCase.field)

			if testCase.want == nil && got != nil {
				t.Fatalf("expected: nil, got: %#v", got)
			}

			if testCase.want != nil && got == nil {
				t.Fatalf("expected: %#v, got: nil", testCase.want)
			}

			if testCase.want != nil && got != nil && *testCase.want != *got {
				t.Fatalf("expected: %#v, got: %#v", testCase.want, got)
			}
		})
	}
}

func TestKeyValueTagsKeyExists(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		key  string
		want bool
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]*string{}),
			key:  "key1",
			want: false,
		},
		{
			name: "non-existent",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key2",
			want: false,
		},
		{
			name: "matching with string value",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key1",
			want: true,
		},
		{
			name: "matching with nil value",
			tags: New(ctx, map[string]*string{"key1": nil}),
			key:  "key1",
			want: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.KeyExists(testCase.key)

			if got != testCase.want {
				t.Fatalf("expected: %t, got: %t", testCase.want, got)
			}
		})
	}
}

func TestKeyValueTagsKeyTagData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		key  string
		want *TagData
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]*string{}),
			key:  "key1",
			want: nil,
		},
		{
			name: "non-existent",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key2",
			want: nil,
		},
		{
			name: "matching with additional boolean fields",
			tags: New(ctx, map[string]*TagData{
				"key1": {
					AdditionalBoolFields: map[string]*bool{"boolfield": testBoolPtr(true)},
					Value:                testStringPtr("value1"),
				},
			}),
			key: "key1",
			want: &TagData{
				AdditionalBoolFields: map[string]*bool{"boolfield": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
		},
		{
			name: "matching with string value",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key1",
			want: &TagData{
				Value: testStringPtr("value1"),
			},
		},
		{
			name: "matching with nil value",
			tags: New(ctx, map[string]*string{"key1": nil}),
			key:  "key1",
			want: nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.KeyTagData(testCase.key)

			if testCase.want == nil && got != nil {
				t.Fatalf("expected: nil, got: %#v", *got)
			}

			if testCase.want != nil && got == nil {
				t.Fatalf("expected: %#v, got: nil", *testCase.want)
			}

			if testCase.want != nil && got != nil && !testCase.want.Equal(got) {
				t.Fatalf("expected: %#v, got: %#v", testCase.want, got)
			}
		})
	}
}

func TestKeyValueTagsKeyValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		key  string
		want *string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]*string{}),
			key:  "key1",
			want: nil,
		},
		{
			name: "non-existent",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key2",
			want: nil,
		},
		{
			name: "matching with string value",
			tags: New(ctx, map[string]*string{"key1": testStringPtr("value1")}),
			key:  "key1",
			want: testStringPtr("value1"),
		},
		{
			name: "matching with nil value",
			tags: New(ctx, map[string]*string{"key1": nil}),
			key:  "key1",
			want: nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.KeyValue(testCase.key)

			if testCase.want == nil && got != nil {
				t.Fatalf("expected: nil, got: %s", *got)
			}

			if testCase.want != nil && got == nil {
				t.Fatalf("expected: %s, got: nil", *testCase.want)
			}

			if testCase.want != nil && got != nil && *testCase.want != *got {
				t.Fatalf("expected: %s, got: %s", *testCase.want, *got)
			}
		})
	}
}

func TestKeyValueTagsKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want []string
	}{
		{
			name: "empty_map_string_interface",
			tags: New(ctx, map[string]interface{}{}),
			want: []string{},
		},
		{
			name: "empty_map_string_stringPointer",
			tags: New(ctx, map[string]*string{}),
			want: []string{},
		},
		{
			name: "empty_map_string_string",
			tags: New(ctx, map[string]string{}),
			want: []string{},
		},
		{
			name: "empty_slice_interface",
			tags: New(ctx, map[string]interface{}{}),
			want: []string{},
		},
		{
			name: "empty_slice_string",
			tags: New(ctx, map[string]string{}),
			want: []string{},
		},
		{
			name: "non_empty_map_string_interface",
			tags: New(ctx, map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_map_string_string",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_map_string_stringPointer",
			tags: New(ctx, map[string]*string{
				"key1": testStringPtr("value1"),
				"key2": testStringPtr("value2"),
				"key3": testStringPtr("value3"),
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_slice_interface",
			tags: New(ctx, []interface{}{
				"key1",
				"key2",
				"key3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_slice_string",
			tags: New(ctx, []string{
				"key1",
				"key2",
				"key3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Keys()

			testKeyValueTagsVerifyKeys(t, got, testCase.want)
		})
	}
}

func TestKeyValueTagsMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty_map_string_interface",
			tags: New(ctx, map[string]interface{}{}),
			want: map[string]string{},
		},
		{
			name: "empty_map_string_string",
			tags: New(ctx, map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "empty_map_string_stringPointer",
			tags: New(ctx, map[string]*string{}),
			want: map[string]string{},
		},
		{
			name: "non_empty_map_string_interface",
			tags: New(ctx, map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_string",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_stringPointer",
			tags: New(ctx, map[string]*string{
				"key1": testStringPtr("value1"),
				"key2": testStringPtr("value2"),
				"key3": testStringPtr("value3"),
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "empty_value",
			tags: New(ctx, map[string]*string{
				"key1": testStringPtr(""),
			}),
			want: map[string]string{
				"key1": "",
			},
		},
		{
			name: "nil_value",
			tags: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: map[string]string{
				"key1": "",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Map()

			testKeyValueTagsVerifyMap(t, got, testCase.want)
		})
	}
}

func TestKeyValueTagsMerge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name      string
		tags      KeyValueTags
		mergeTags KeyValueTags
		want      map[string]string
	}{
		{
			name:      "empty",
			tags:      New(ctx, map[string]string{}),
			mergeTags: New(ctx, map[string]string{}),
			want:      map[string]string{},
		},
		{
			name: "add_all",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(ctx, map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(ctx, map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
		},
		{
			name: "update_all",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(ctx, map[string]string{
				"key1": "value1updated",
				"key2": "value2updated",
				"key3": "value3updated",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key2": "value2updated",
				"key3": "value3updated",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Merge(testCase.mergeTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name     string
		tags     KeyValueTags
		onlyTags KeyValueTags
		want     map[string]string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			onlyTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			onlyTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "mixed",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			onlyTags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "none",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			onlyTags: New(ctx, map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Only(testCase.onlyTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name    string
		oldTags KeyValueTags
		newTags KeyValueTags
		want    map[string]string
	}{
		{
			name:    "empty",
			oldTags: New(ctx, map[string]string{}),
			newTags: New(ctx, map[string]string{}),
			want:    map[string]string{},
		},
		{
			name: "all_new",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "mixed",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no_changes",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.oldTags.Removed(testCase.newTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsUpdated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name    string
		oldTags KeyValueTags
		newTags KeyValueTags
		want    map[string]string
	}{
		{
			name:    "empty",
			oldTags: New(ctx, map[string]string{}),
			newTags: New(ctx, map[string]string{}),
			want:    map[string]string{},
		},
		{
			name: "all_new",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			},
		},
		{
			name: "mixed",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			},
		},
		{
			name: "no_changes",
			oldTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.oldTags.Updated(testCase.newTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsChunks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		size int
		want []int
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			size: 10,
			want: []int{},
		},
		{
			name: "chunk_1",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 1,
			want: []int{1, 1, 1, 1},
		},
		{
			name: "chunk_2",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 2,
			want: []int{2, 2},
		},
		{
			name: "chunk_3",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 3,
			want: []int{3, 1},
		},
		{
			name: "chunk_4",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 4,
			want: []int{4},
		},
		{
			name: "chunk_5",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 5,
			want: []int{4},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Chunks(testCase.size)

			if len(got) != len(testCase.want) {
				t.Errorf("unexpected number of chunks: %d", len(got))
			}

			for i, n := range testCase.want {
				if len(got[i]) != n {
					t.Errorf("chunk (%d) length %d; want length %d", i, len(got[i]), n)
				}
			}
		})
	}
}

func TestKeyValueTagsContainsAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name   string
		source KeyValueTags
		target KeyValueTags
		want   bool
	}{
		{
			name:   "empty",
			source: New(ctx, map[string]string{}),
			target: New(ctx, map[string]string{}),
			want:   true,
		},
		{
			name:   "source_empty",
			source: New(ctx, map[string]string{}),
			target: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: false,
		},
		{
			name: "target_empty",
			source: New(ctx, map[string]string{
				"key1": "value1",
			}),
			target: New(ctx, map[string]string{}),
			want:   true,
		},
		{
			name: "nil value matches",
			source: New(ctx, map[string]*string{
				"key1": nil,
			}),
			target: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: true,
		},
		{
			name: "exact_match",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			want: true,
		},
		{
			name: "source_contains_all",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key3": "value3",
			}),
			want: true,
		},
		{
			name: "source_does_not_contain_all",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key4": "value4",
			}),
			want: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.source.ContainsAll(testCase.target)

			if got != testCase.want {
				t.Errorf("unexpected ContainsAll: %t", got)
			}
		})
	}
}

func TestKeyValueTagsEqual(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name   string
		source KeyValueTags
		target KeyValueTags
		want   bool
	}{
		{
			name:   "nil",
			source: nil,
			target: nil,
			want:   true,
		},
		{
			name:   "empty",
			source: New(ctx, map[string]string{}),
			target: New(ctx, map[string]string{}),
			want:   true,
		},
		{
			name:   "source_nil",
			source: nil,
			target: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: false,
		},
		{
			name:   "source_empty",
			source: New(ctx, map[string]string{}),
			target: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: false,
		},
		{
			name: "target_nil",
			source: New(ctx, map[string]string{
				"key1": "value1",
			}),
			target: nil,
			want:   false,
		},
		{
			name: "target_empty",
			source: New(ctx, map[string]string{
				"key1": "value1",
			}),
			target: New(ctx, map[string]string{}),
			want:   false,
		},
		{
			name: "nil value matches",
			source: New(ctx, map[string]*string{
				"key1": nil,
			}),
			target: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: true,
		},
		{
			name: "exact_match",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			want: true,
		},
		{
			name: "source_contains_all",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key3": "value3",
			}),
			want: false,
		},
		{
			name: "source_does_not_contain_all",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key4": "value4",
			}),
			want: false,
		},
		{
			name: "target_value_neq",
			source: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value4",
			}),
			want: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.source.Equal(testCase.target)

			if got != testCase.want {
				t.Errorf("unexpected Equal: %t", got)
			}
		})
	}
}

func TestKeyValueTagsHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		zero bool
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			zero: true,
		},
		{
			name: "nil value",
			tags: New(ctx, map[string]*string{
				"key1": nil,
			}),
			zero: false,
		},
		{
			name: "not_empty",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			zero: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.Hash()

			if (got == 0 && !testCase.zero) || (got != 0 && testCase.zero) {
				t.Errorf("unexpected hash code: %d", got)
			}
		})
	}
}

func TestKeyValueTagsRemoveDefaultConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name          string
		tags          KeyValueTags
		defaultConfig *DefaultConfig
		want          map[string]string
	}{
		{
			name: "empty config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no config",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: nil,
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no tags",
			tags: New(ctx, map[string]string{}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "keys all matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
			},
			want: map[string]string{},
		},
		{
			name: "keys some matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
				}),
			},
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys some overridden",
			tags: New(ctx, map[string]string{
				"key1": "value2",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key1": "value1",
				}),
			},
			want: map[string]string{
				"key1": "value2",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "keys none matching",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			defaultConfig: &DefaultConfig{
				Tags: New(ctx, map[string]string{
					"key4": "value4",
					"key5": "value5",
					"key6": "value6",
				}),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.RemoveDefaultConfig(testCase.defaultConfig)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsURLEncode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			want: "",
		},
		{
			name: "nil value",
			tags: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: "",
		},
		{
			name: "single",
			tags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: "key1=value1",
		},
		{
			name: "multiple",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: "key1=value1&key2=value2&key3=value3",
		},
		{
			name: "multiple_with_encoded",
			tags: New(ctx, map[string]string{
				"key1":  "value 1",
				"key@2": "value+:2",
				"key3":  "value3",
			}),
			want: "key1=value+1&key3=value3&key%402=value%2B%3A2",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.URLEncode()

			if got != testCase.want {
				t.Errorf("unexpected URL encoded value: %q", got)
			}
		})
	}
}

func TestKeyValueTagsURLQueryString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			want: "",
		},
		{
			name: "nil value",
			tags: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: "",
		},
		{
			name: "single",
			tags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: "key1=value1",
		},
		{
			name: "multiple",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: "key1=value1&key2=value2&key3=value3",
		},
		{
			name: "multiple_with_encoded",
			tags: New(ctx, map[string]string{
				"key1":  "value 1",
				"key@2": "value+:2",
				"key3":  "value3",
			}),
			want: "key1=value 1&key3=value3&key@2=value+:2",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.URLQueryString()

			if got != testCase.want {
				t.Errorf("unexpected query string value: %q", got)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name   string
		source interface{}
		want   map[string]string
	}{
		{
			name:   "empty_KeyValueTags",
			source: KeyValueTags{},
			want:   map[string]string{},
		},
		{
			name:   "empty_map_string_TagDataPointer",
			source: map[string]*TagData{},
			want:   map[string]string{},
		},
		{
			name:   "empty_map_string_interface",
			source: map[string]interface{}{},
			want:   map[string]string{},
		},
		{
			name:   "empty_map_string_string",
			source: map[string]string{},
			want:   map[string]string{},
		},
		{
			name:   "empty_map_string_stringPointer",
			source: map[string]*string{},
			want:   map[string]string{},
		},
		{
			name:   "empty_slice_interface",
			source: []interface{}{},
			want:   map[string]string{},
		},
		{
			name:   "empty_typesMap",
			source: types.MapValueMust(types.StringType, map[string]attr.Value{}),
			want:   map[string]string{},
		},
		{
			name: "non_empty_KeyValueTags",
			source: KeyValueTags{
				"key1": &TagData{
					Value: nil,
				},
				"key2": &TagData{
					Value: testStringPtr(""),
				},
				"key3": &TagData{
					Value: testStringPtr("value3"),
				},
			},
			want: map[string]string{
				"key1": "",
				"key2": "",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_TagDataPointer",
			source: map[string]*TagData{
				"key1": {
					Value: nil,
				},
				"key2": {
					Value: testStringPtr(""),
				},
				"key3": {
					Value: testStringPtr("value3"),
				},
			},
			want: map[string]string{
				"key1": "",
				"key2": "",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_interface",
			source: map[string]interface{}{
				"key1": nil,
				"key2": "",
				"key3": "value3",
			},
			want: map[string]string{
				"key1": "",
				"key2": "",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_string",
			source: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			want: map[string]string{
				"key1": "",
				"key2": "value2",
			},
		},
		{
			name: "non_empty_map_string_stringPointer",
			source: map[string]*string{
				"key1": nil,
				"key2": testStringPtr(""),
				"key3": testStringPtr("value3"),
			},
			want: map[string]string{
				"key1": "",
				"key2": "",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_slice_interface",
			source: []interface{}{
				"key1",
				"key2",
			},
			want: map[string]string{
				"key1": "",
				"key2": "",
			},
		},
		{
			name: "non_empty_typesMap",
			source: types.MapValueMust(types.StringType, map[string]attr.Value{
				"key1": types.StringValue(""),
				"key2": types.StringValue("value2"),
			}),
			want: map[string]string{
				"key1": "",
				"key2": "value2",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := New(ctx, testCase.source)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)

			// Verify that any source KeyTagValues types are copied
			// Unfortunately must be done for each separate type
			switch src := testCase.source.(type) {
			case KeyValueTags:
				src.Merge(New(ctx, map[string]string{"mergekey": "mergevalue"}))

				_, ok := got.Map()["mergekey"]

				if ok {
					t.Fatal("expected source to be copied, got source modification")
				}
			case map[string]*TagData:
				src["mergekey"] = &TagData{Value: testStringPtr("mergevalue")}

				_, ok := got.Map()["mergekey"]

				if ok {
					t.Fatal("expected source to be copied, got source modification")
				}
			}
		})
	}
}

func TestTagDataEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tagData1 *TagData
		tagData2 *TagData
		want     bool
	}{
		{
			name:     "both nil",
			tagData1: nil,
			tagData2: nil,
			want:     true,
		},
		{
			name:     "first nil",
			tagData1: nil,
			tagData2: &TagData{
				Value: testStringPtr("value1"),
			},
			want: false,
		},
		{
			name: "second nil",
			tagData1: &TagData{
				Value: testStringPtr("value1"),
			},
			tagData2: nil,
			want:     false,
		},
		{
			name: "differing value",
			tagData1: &TagData{
				Value: testStringPtr("value1"),
			},
			tagData2: &TagData{
				Value: testStringPtr("value2"),
			},
			want: false,
		},
		{
			name: "differing additional bool fields",
			tagData1: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalBoolFields: map[string]*bool{"field2": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			want: false,
		},
		{
			name: "differing additional bool field values",
			tagData1: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(false)},
				Value:                testStringPtr("value1"),
			},
			want: false,
		},
		{
			name: "differing additional string fields",
			tagData1: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalStringFields: map[string]*string{"field2": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			want: false,
		},
		{
			name: "differing additional string field values",
			tagData1: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field2value")},
				Value:                  testStringPtr("value1"),
			},
			want: false,
		},
		{
			name: "same value",
			tagData1: &TagData{
				Value: testStringPtr("value1"),
			},
			tagData2: &TagData{
				Value: testStringPtr("value1"),
			},
			want: true,
		},
		{
			name: "same additional bool fields",
			tagData1: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			want: true,
		},
		{
			name: "same additional string fields",
			tagData1: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			tagData2: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			want: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tagData1.Equal(testCase.tagData2)

			if testCase.want != got {
				t.Fatalf("expected: %t, got: %t", testCase.want, got)
			}
		})
	}
}

func TestTagDataString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		tagData *TagData
		want    string
	}{
		{
			name:    "nil",
			tagData: nil,
			want:    "",
		},
		{
			name: "value",
			tagData: &TagData{
				Value: testStringPtr("value1"),
			},
			want: "TagData{Value: value1}",
		},
		{
			name: "additional bool fields",
			tagData: &TagData{
				AdditionalBoolFields: map[string]*bool{"field1": testBoolPtr(true)},
				Value:                testStringPtr("value1"),
			},
			want: "TagData{AdditionalBoolFields: map[field1:true], Value: value1}",
		},
		{
			name: "additional string fields",
			tagData: &TagData{
				AdditionalStringFields: map[string]*string{"field1": testStringPtr("field1value")},
				Value:                  testStringPtr("value1"),
			},
			want: "TagData{AdditionalStringFields: map[field1:field1value], Value: value1}",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tagData.String()

			if testCase.want != got {
				t.Fatalf("expected: %s, got: %s", testCase.want, got)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Input    string
		Expected string
	}{
		{
			Input:    "ARN",
			Expected: "arn",
		},
		{
			Input:    "PropagateAtLaunch",
			Expected: "propagate_at_launch",
		},
		{
			Input:    "ResourceId",
			Expected: "resource_id",
		},
		{
			Input:    "ResourceArn",
			Expected: "resource_arn",
		},
		{
			Input:    "ResourceARN",
			Expected: "resource_arn",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()

			got := ToSnakeCase(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestKeyValueTagsString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name string
		tags KeyValueTags
		want string
	}{
		{
			name: "empty",
			tags: New(ctx, map[string]string{}),
			want: "map[]",
		},
		{
			name: "no value",
			tags: New(ctx, map[string]*string{
				"key1": nil,
			}),
			want: "map[key1:]",
		},
		{
			name: "single",
			tags: New(ctx, map[string]string{
				"key1": "value1",
			}),
			want: "map[key1:TagData{Value: value1}]",
		},
		{
			name: "multiple",
			tags: New(ctx, map[string]string{
				"key1": "value1",
				"key3": "value3",
				"key2": "value2",
				"key5": "value5",
				"key4": "value4",
			}),
			want: "map[key1:TagData{Value: value1} key2:TagData{Value: value2} key3:TagData{Value: value3} key4:TagData{Value: value4} key5:TagData{Value: value5}]",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tags.String()

			if got != testCase.want {
				t.Errorf("unexpected string value: %q", got)
			}
		})
	}
}

func testKeyValueTagsVerifyKeys(t *testing.T, got []string, want []string) {
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

func testKeyValueTagsVerifyMap(t *testing.T, got map[string]string, want map[string]string) {
	for k, wantV := range want {
		gotV, ok := got[k]

		if !ok {
			t.Errorf("want missing key: %s", k)
			continue
		}

		if gotV != wantV {
			t.Errorf("got key (%s) value %s; want value %s", k, gotV, wantV)
		}
	}

	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("got extra key: %s", k)
		}
	}
}

func testBoolPtr(b bool) *bool {
	return &b
}

func testStringPtr(str string) *string {
	return &str
}
