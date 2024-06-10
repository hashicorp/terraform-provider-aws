// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"maps"
	"slices"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExpandTableItemAttributes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected map[string]awstypes.AttributeValue
	}{
		"B": {
			input: fmt.Sprintf(`{"attr":{"B":"%s"}}`, base64.StdEncoding.EncodeToString([]byte("blob"))),
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberB{
					Value: []byte("blob"),
				},
			},
		},
		"BOOL": {
			input: `{"true":{"BOOL":true},"false":{"BOOL":false}}`,
			expected: map[string]awstypes.AttributeValue{
				acctest.CtTrue: &awstypes.AttributeValueMemberBOOL{
					Value: true,
				},
				acctest.CtFalse: &awstypes.AttributeValueMemberBOOL{
					Value: false,
				},
			},
		},
		"BS": {
			input: fmt.Sprintf(`{"attr":{"BS":["%[1]s","%[2]s"]}}`,
				base64.StdEncoding.EncodeToString([]byte("blob1")),
				base64.StdEncoding.EncodeToString([]byte("blob2")),
			),
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberBS{
					Value: [][]byte{
						[]byte("blob1"),
						[]byte("blob2"),
					},
				},
			},
		},
		"L": {
			input: `{"attr":{"L":[{"S":"one"},{"N":"2"}]}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberL{
					Value: []awstypes.AttributeValue{
						&awstypes.AttributeValueMemberS{Value: "one"},
						&awstypes.AttributeValueMemberN{Value: acctest.Ct2},
					},
				},
			},
		},
		"M": {
			input: `{"attr":{"M":{"one":{"S":"one"},"two":{"N":"2"}}}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberM{
					Value: map[string]awstypes.AttributeValue{
						"one": &awstypes.AttributeValueMemberS{Value: "one"},
						"two": &awstypes.AttributeValueMemberN{Value: acctest.Ct2},
					},
				},
			},
		},
		"N": {
			input: `{"attr":{"N":"123"}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberN{
					Value: "123",
				},
			},
		},
		"NS": {
			input: `{"attr":{"NS":["42.2","-19"]}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberNS{
					Value: []string{"42.2", "-19"},
				},
			},
		},
		"NULL": {
			input: `{"attr":{"NULL":true}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberNULL{
					Value: true,
				},
			},
		},
		"S": {
			input: `{"attr":{"S":"value"}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberS{
					Value: names.AttrValue,
				},
			},
		},
		"SS": {
			input: `{"attr":{"SS":["one","two"]}}`,
			expected: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberSS{
					Value: []string{"one", "two"},
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := tfdynamodb.ExpandTableItemAttributes(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !maps.EqualFunc(actual, tc.expected, attributeValuesEqual) {
				t.Fatalf("expected\n%s\ngot\n%s", tc.expected, actual)
			}
		})
	}
}

func attributeValuesEqual(a, b awstypes.AttributeValue) bool {
	switch a := a.(type) {
	case *awstypes.AttributeValueMemberB:
		return bytes.Equal(a.Value, b.(*awstypes.AttributeValueMemberB).Value)
	case *awstypes.AttributeValueMemberBOOL:
		return a.Value == b.(*awstypes.AttributeValueMemberBOOL).Value
	case *awstypes.AttributeValueMemberBS:
		return slices.EqualFunc(a.Value, b.(*awstypes.AttributeValueMemberBS).Value, func(x, y []byte) bool {
			return bytes.Equal(x, y)
		})
	case *awstypes.AttributeValueMemberL:
		return slices.EqualFunc(a.Value, b.(*awstypes.AttributeValueMemberL).Value, attributeValuesEqual)
	case *awstypes.AttributeValueMemberM:
		return maps.EqualFunc(a.Value, b.(*awstypes.AttributeValueMemberM).Value, attributeValuesEqual)
	case *awstypes.AttributeValueMemberN:
		return a.Value == b.(*awstypes.AttributeValueMemberN).Value
	case *awstypes.AttributeValueMemberNS:
		return slices.EqualFunc(a.Value, b.(*awstypes.AttributeValueMemberNS).Value, func(x, y string) bool {
			return x == y
		})
	case *awstypes.AttributeValueMemberNULL:
		return a.Value == b.(*awstypes.AttributeValueMemberNULL).Value
	case *awstypes.AttributeValueMemberS:
		return a.Value == b.(*awstypes.AttributeValueMemberS).Value
	case *awstypes.AttributeValueMemberSS:
		return slices.EqualFunc(a.Value, b.(*awstypes.AttributeValueMemberSS).Value, func(x, y string) bool {
			return x == y
		})
	}

	return false
}

func TestFlattenTableItemAttributes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		attrs    map[string]awstypes.AttributeValue
		expected string
	}{
		"B": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberB{
					Value: []byte("blob"),
				},
			},
			expected: fmt.Sprintf(`{"attr":{"B":"%s"}}`, base64.StdEncoding.EncodeToString([]byte("blob"))),
		},
		"BOOL": {
			attrs: map[string]awstypes.AttributeValue{
				acctest.CtTrue: &awstypes.AttributeValueMemberBOOL{
					Value: true,
				},
				acctest.CtFalse: &awstypes.AttributeValueMemberBOOL{
					Value: false,
				},
			},
			expected: `{"true":{"BOOL":true},"false":{"BOOL":false}}`,
		},
		"BS": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberBS{
					Value: [][]byte{
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
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberL{
					Value: []awstypes.AttributeValue{
						&awstypes.AttributeValueMemberS{Value: "one"},
						&awstypes.AttributeValueMemberN{Value: acctest.Ct2},
					},
				},
			},
			expected: `{"attr":{"L":[{"S":"one"},{"N":"2"}]}}`,
		},
		"M": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberM{
					Value: map[string]awstypes.AttributeValue{
						"one": &awstypes.AttributeValueMemberS{Value: "one"},
						"two": &awstypes.AttributeValueMemberN{Value: acctest.Ct2},
					},
				},
			},
			expected: `{"attr":{"M":{"one":{"S":"one"},"two":{"N":"2"}}}}`,
		},
		"N": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberN{
					Value: "123",
				},
			},
			expected: `{"attr":{"N":"123"}}`,
		},
		"NS": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberNS{
					Value: []string{"42.2", "-19"},
				},
			},
			expected: `{"attr":{"NS":["42.2","-19"]}}`,
		},
		"NULL": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberNULL{
					Value: true,
				},
			},
			expected: `{"attr":{"NULL":true}}`,
		},
		"S": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberS{
					Value: names.AttrValue,
				},
			},
			expected: `{"attr":{"S":"value"}}`,
		},
		"SS": {
			attrs: map[string]awstypes.AttributeValue{
				"attr": &awstypes.AttributeValueMemberSS{
					Value: []string{"one", "two"},
				},
			},
			expected: `{"attr":{"SS":["one","two"]}}`,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := tfdynamodb.FlattenTableItemAttributes(tc.attrs)
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
