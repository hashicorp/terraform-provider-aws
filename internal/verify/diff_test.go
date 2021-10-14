package verify

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSuppressEquivalentJSONDiffsWhitespaceAndNoWhitespace(t *testing.T) {
	d := new(schema.ResourceData)

	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !SuppressEquivalentJSONDiffs("", noWhitespace, whitespace, d) {
		t.Errorf("Expected SuppressEquivalentJSONDiffs to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if SuppressEquivalentJSONDiffs("", noWhitespaceDiff, whitespaceDiff, d) {
		t.Errorf("Expected SuppressEquivalentJSONDiffs to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestSuppressEquivalentTypeStringBoolean(t *testing.T) {
	testCases := []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        "false",
			new:        "0",
			equivalent: true,
		},
		{
			old:        "true",
			new:        "1",
			equivalent: true,
		},
		{
			old:        "",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "",
			new:        "1",
			equivalent: false,
		},
	}

	for i, tc := range testCases {
		value := SuppressEquivalentTypeStringBoolean("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestSuppressEquivalentJSONOrYAMLDiffs(t *testing.T) {
	testCases := []struct {
		description string
		equivalent  bool
		old         string
		new         string
	}{
		{
			description: `JSON no change`,
			equivalent:  true,
			old: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `JSON whitespace`,
			equivalent:  true,
			old:         `{"Resources":{"TestVpc":{"Type":"AWS::EC2::VPC","Properties":{"CidrBlock":"10.0.0.0/16"}}},"Outputs":{"TestVpcID":{"Value":{"Ref":"TestVpc"}}}}`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `JSON change`,
			equivalent:  false,
			old: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "172.16.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `YAML no change`,
			equivalent:  true,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
		{
			description: `YAML whitespace`,
			equivalent:  false,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16

Outputs:
  TestVpcID:
    Value: !Ref TestVpc

`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
		{
			description: `YAML change`,
			equivalent:  false,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 172.16.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
	}

	for _, tc := range testCases {
		value := SuppressEquivalentJSONOrYAMLDiffs("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case (%s) to be equivalent", tc.description)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case (%s) to not be equivalent", tc.description)
		}
	}
}

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
		c, r, u := DiffStringMaps(tc.Old, tc.New)
		cm := PointersMapToStringList(c)
		rm := PointersMapToStringList(r)
		um := PointersMapToStringList(u)
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
