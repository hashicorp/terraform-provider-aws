// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func TestExpandTableItemAttributes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected map[string]*dynamodb.AttributeValue
	}{
		"B": {
			input: fmt.Sprintf(`{"attr":{"B":"%s"}}`, base64.StdEncoding.EncodeToString([]byte("blob"))),
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					B: []byte("blob"),
				},
			},
		},
		"BOOL": {
			input: `{"true":{"BOOL":true},"false":{"BOOL":false}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"true": {
					BOOL: aws.Bool(true),
				},
				"false": {
					BOOL: aws.Bool(false),
				},
			},
		},
		"BS": {
			input: fmt.Sprintf(`{"attr":{"BS":["%[1]s","%[2]s"]}}`,
				base64.StdEncoding.EncodeToString([]byte("blob1")),
				base64.StdEncoding.EncodeToString([]byte("blob2")),
			),
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					BS: [][]byte{
						[]byte("blob1"),
						[]byte("blob2"),
					},
				},
			},
		},
		"L": {
			input: `{"attr":{"L":[{"S":"one"},{"N":"2"}]}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					L: []*dynamodb.AttributeValue{
						{S: aws.String("one")},
						{N: aws.String("2")},
					},
				},
			},
		},
		"M": {
			input: `{"attr":{"M":{"one":{"S":"one"},"two":{"N":"2"}}}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					M: map[string]*dynamodb.AttributeValue{
						"one": {S: aws.String("one")},
						"two": {N: aws.String("2")},
					},
				},
			},
		},
		"N": {
			input: `{"attr":{"N":"123"}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					N: aws.String("123"),
				},
			},
		},
		"NS": {
			input: `{"attr":{"NS":["42.2","-19"]}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					NS: aws.StringSlice([]string{"42.2", "-19"}),
				},
			},
		},
		"NULL": {
			input: `{"attr":{"NULL":true}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					NULL: aws.Bool(true),
				},
			},
		},
		"S": {
			input: `{"attr":{"S":"value"}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					S: aws.String("value"),
				},
			},
		},
		"SS": {
			input: `{"attr":{"SS":["one","two"]}}`,
			expected: map[string]*dynamodb.AttributeValue{
				"attr": {
					SS: aws.StringSlice([]string{"one", "two"}),
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := ExpandTableItemAttributes(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !maps.EqualFunc(actual, tc.expected, attributeValuesEqual) {
				t.Fatalf("expected\n%s\ngot\n%s", tc.expected, actual)
			}
		})
	}
}

func attributeValuesEqual(a, b *dynamodb.AttributeValue) bool {
	if a.B != nil {
		return bytes.Equal(a.B, b.B)
	}
	if a.BOOL != nil {
		return b.BOOL != nil && aws.BoolValue(a.BOOL) == aws.BoolValue(b.BOOL)
	}
	if a.BS != nil {
		return slices.EqualFunc(a.BS, b.BS, func(x, y []byte) bool {
			return bytes.Equal(x, y)
		})
	}
	if a.L != nil {
		return slices.EqualFunc(a.L, b.L, attributeValuesEqual)
	}
	if a.M != nil {
		return maps.EqualFunc(a.M, b.M, attributeValuesEqual)
	}
	if a.N != nil {
		return b.N != nil && aws.StringValue(a.N) == aws.StringValue(b.N)
	}
	if a.NS != nil {
		return slices.EqualFunc(a.NS, b.NS, func(x, y *string) bool {
			return aws.StringValue(x) == aws.StringValue(y)
		})
	}
	if a.NULL != nil {
		return b.NULL != nil && aws.BoolValue(a.NULL) == aws.BoolValue(b.NULL)
	}
	if a.S != nil {
		return b.S != nil && aws.StringValue(a.S) == aws.StringValue(b.S)
	}
	if a.SS != nil {
		return slices.EqualFunc(a.SS, b.SS, func(x, y *string) bool {
			return aws.StringValue(x) == aws.StringValue(y)
		})
	}
	return false
}

func TestFlattenTableItemAttributes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		attrs    map[string]*dynamodb.AttributeValue
		expected string
	}{
		"B": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					B: []byte("blob"),
				},
			},
			expected: fmt.Sprintf(`{"attr":{"B":"%s"}}`, base64.StdEncoding.EncodeToString([]byte("blob"))),
		},
		"BOOL": {
			attrs: map[string]*dynamodb.AttributeValue{
				"true": {
					BOOL: aws.Bool(true),
				},
				"false": {
					BOOL: aws.Bool(false),
				},
			},
			expected: `{"true":{"BOOL":true},"false":{"BOOL":false}}`,
		},
		"BS": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					BS: [][]byte{
						[]byte("blob1"),
						[]byte("blob2"),
					},
				},
			},
			expected: fmt.Sprintf(`{"attr":{"BS":["%[1]s","%[2]s"]}}`,
				base64.StdEncoding.EncodeToString([]byte("blob1")),
				base64.StdEncoding.EncodeToString([]byte("blob2")),
			),
		},
		"L": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					L: []*dynamodb.AttributeValue{
						{S: aws.String("one")},
						{N: aws.String("2")},
					},
				},
			},
			expected: `{"attr":{"L":[{"S":"one"},{"N":"2"}]}}`,
		},
		"M": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					M: map[string]*dynamodb.AttributeValue{
						"one": {S: aws.String("one")},
						"two": {N: aws.String("2")},
					},
				},
			},
			expected: `{"attr":{"M":{"one":{"S":"one"},"two":{"N":"2"}}}}`,
		},
		"N": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					N: aws.String("123"),
				},
			},
			expected: `{"attr":{"N":"123"}}`,
		},
		"NS": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					NS: aws.StringSlice([]string{"42.2", "-19"}),
				},
			},
			expected: `{"attr":{"NS":["42.2","-19"]}}`,
		},
		"NULL": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					NULL: aws.Bool(true),
				},
			},
			expected: `{"attr":{"NULL":true}}`,
		},
		"S": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					S: aws.String("value"),
				},
			},
			expected: `{"attr":{"S":"value"}}`,
		},
		"SS": {
			attrs: map[string]*dynamodb.AttributeValue{
				"attr": {
					SS: aws.StringSlice([]string{"one", "two"}),
				},
			},
			expected: `{"attr":{"SS":["one","two"]}}`,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := flattenTableItemAttributes(tc.attrs)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			e, err := structure.NormalizeJsonString(tc.expected)
			if err != nil {
				t.Fatalf("normalizing expected JSON: %s", err)
			}

			a, err := structure.NormalizeJsonString(actual)
			if err != nil {
				t.Fatalf("normalizing returned JSON: %s", err)
			}

			if a != e {
				t.Fatalf("expected\n%s\ngot\n%s", e, a)
			}
		})
	}
}
