package aws

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestDiffStringMaps(t *testing.T) {
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

	for i, tc := range cases {
		c, r, u := diffStringMaps(tc.Old, tc.New)
		cm := pointersMapToStringList(c)
		rm := pointersMapToStringList(r)
		um := pointersMapToStringList(u)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
		if !reflect.DeepEqual(um, tc.Unchanged) {
			t.Fatalf("%d: bad unchanged: %#v", i, rm)
		}
	}
}































































func TestCheckYamlString(t *testing.T) {
	var err error
	var actual string

	validYaml := `---
abc:
  def: 123
  xyz:
    -
      a: "ホリネズミ"
      b: "1"
`

	actual, err = checkYamlString(validYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the same YAML string back
	if actual != validYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validYaml)
	}

	invalidYaml := `abc: [`

	actual, err = checkYamlString(invalidYaml)
	if err == nil {
		t.Fatalf("Expected to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the invalid YAML to be shown back to us again.
	if actual != invalidYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, invalidYaml)
	}
}

func TestNormalizeJsonOrYamlString(t *testing.T) {
	var err error
	var actual string

	validNormalizedJson := `{"abc":"1"}`
	actual, err = normalizeJsonOrYamlString(validNormalizedJson)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedJson {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedJson)
	}

	validNormalizedYaml := `abc: 1
`
	actual, err = normalizeJsonOrYamlString(validNormalizedYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedYaml)
	}
}









