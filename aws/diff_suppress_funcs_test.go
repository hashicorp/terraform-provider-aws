package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSuppressEquivalentJsonDiffsWhitespaceAndNoWhitespace(t *testing.T) {
	d := new(schema.ResourceData)

	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !suppressEquivalentJsonDiffs("", noWhitespace, whitespace, d) {
		t.Errorf("Expected suppressEquivalentJsonDiffs to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if suppressEquivalentJsonDiffs("", noWhitespaceDiff, whitespaceDiff, d) {
		t.Errorf("Expected suppressEquivalentJsonDiffs to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
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
		value := suppressEquivalentTypeStringBoolean("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestSuppressEquivalentJsonOrYamlDiffs(t *testing.T) {
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
		value := suppressEquivalentJsonOrYamlDiffs("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case (%s) to be equivalent", tc.description)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case (%s) to not be equivalent", tc.description)
		}
	}
}

func TestSuppressRoute53ZoneNameWithTrailingDot(t *testing.T) {
	testCases := []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        "example.com",
			new:        "example.com",
			equivalent: true,
		},
		{
			old:        "example.com.",
			new:        "example.com.",
			equivalent: true,
		},
		{
			old:        "example.com.",
			new:        "example.com",
			equivalent: true,
		},
		{
			old:        "example.com",
			new:        "example.com.",
			equivalent: true,
		},
		{
			old:        ".",
			new:        "",
			equivalent: false,
		},
		{
			old:        "",
			new:        ".",
			equivalent: false,
		},
		{
			old:        ".",
			new:        ".",
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		value := suppressRoute53ZoneNameWithTrailingDot("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestSuppressDmsReplicationTaskSettingsDiffs(t *testing.T) {
	testCases := []struct {
		description string
		equivalent  bool
		old         string
		new         string
	}{
		{
			description: `JSON change`,
			equivalent:  false,
			old: `
{
	"ControlTablesSettings": {
		"ControlSchema": "dms_control",
		"HistoryTimeslotInMinutes": 10,
		"HistoryTableEnabled": true,
		"SuspendedTablesTableEnabled": true,
		"StatusTableEnabled": false
	}
}
`,
			new: `
{
	"ControlTablesSettings": {
		"ControlSchema": "dms_control",
		"HistoryTimeslotInMinutes": 5,
		"HistoryTableEnabled": true,
		"SuspendedTablesTableEnabled": true,
		"StatusTableEnabled": false
	}
}
`,
		},
		{
			description: `historyTimeslotInMinutes ignored`,
			equivalent:  true,
			old: `
{
	"ControlTablesSettings": {
		"historyTimeslotInMinutes": 5,
		"ControlSchema": "dms_control",
		"HistoryTimeslotInMinutes": 5,
		"HistoryTableEnabled": true,
		"SuspendedTablesTableEnabled": true,
		"StatusTableEnabled": false
	}
}
`,
			new: `
{
	"ControlTablesSettings": {
		"ControlSchema": "dms_control",
		"HistoryTimeslotInMinutes": 5,
		"HistoryTableEnabled": true,
		"SuspendedTablesTableEnabled": true,
		"StatusTableEnabled": false
	}
}
`,
		},
		{
			description: `CloudWatchLogGroup, CloudWatchLogStream ignored`,
			equivalent:  true,
			old: `
{
	"Logging": {
		"EnableLogging": true,
		"LogComponents": [
			{
				"Id": "SOURCE_UNLOAD",
				"Severity": "LOGGER_SEVERITY_DEFAULT"
			},
			{
				"Id": "TARGET_LOAD",
				"Severity": "LOGGER_SEVERITY_DEFAULT"
			},
			{
				"Id": "SOURCE_CAPTURE",
				"Severity": "LOGGER_SEVERITY_DEFAULT"
			},
			{
				"Id": "TARGET_APPLY",
				"Severity": "LOGGER_SEVERITY_DEFAULT"
			},
			{
				"Id": "TASK_MANAGER",
				"Severity": "LOGGER_SEVERITY_DEFAULT"
			}
		],
		"CloudWatchLogGroup": "dms-tasks-task-name",
		"CloudWatchLogStream": "dms-task-ABCDEF123456789"
	}
}
`,
			new: `
{
    "Logging": {
        "EnableLogging": true,
        "LogComponents": [
            {
                "Id": "SOURCE_UNLOAD",
                "Severity": "LOGGER_SEVERITY_DEFAULT"
            },
            {
                "Id": "TARGET_LOAD",
                "Severity": "LOGGER_SEVERITY_DEFAULT"
            },
            {
                "Id": "SOURCE_CAPTURE",
                "Severity": "LOGGER_SEVERITY_DEFAULT"
            },
            {
                "Id": "TARGET_APPLY",
                "Severity": "LOGGER_SEVERITY_DEFAULT"
            },
            {
                "Id": "TASK_MANAGER",
                "Severity": "LOGGER_SEVERITY_DEFAULT"
            }
        ]
	}
}
`,
		},
	}

	for _, tc := range testCases {
		value := suppressDmsReplicationTaskSettingsDiffs("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case (%s) to be equivalent", tc.description)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case (%s) to not be equivalent", tc.description)
		}
	}
}
