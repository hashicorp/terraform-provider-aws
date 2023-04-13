package dynamodb

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

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
