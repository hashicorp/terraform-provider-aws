// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestLooksLikeJSONString(t *testing.T) {
	t.Parallel()

	looksLikeJSON := ` {"abc":"1"} `
	doesNotLookLikeJSON := `abc: 1`

	if !looksLikeJSONString(looksLikeJSON) {
		t.Errorf("Expected looksLikeJSON to return true for %s", looksLikeJSON)
	}
	if looksLikeJSONString(doesNotLookLikeJSON) {
		t.Errorf("Expected looksLikeJSON to return false for %s", doesNotLookLikeJSON)
	}
}

func TestJSONBytesEqualQuotedAndUnquoted(t *testing.T) {
	t.Parallel()

	unquoted := `{"test": "test"}`
	quoted := "{\"test\": \"test\"}"

	if !JSONBytesEqual([]byte(unquoted), []byte(quoted)) {
		t.Errorf("Expected JSONBytesEqual to return true for %s == %s", unquoted, quoted)
	}

	unquotedDiff := `{"test": "test"}`
	quotedDiff := "{\"test\": \"tested\"}"

	if JSONBytesEqual([]byte(unquotedDiff), []byte(quotedDiff)) {
		t.Errorf("Expected JSONBytesEqual to return false for %s == %s", unquotedDiff, quotedDiff)
	}
}

func TestJSONBytesEqualWhitespaceAndNoWhitespace(t *testing.T) {
	t.Parallel()

	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !JSONBytesEqual([]byte(noWhitespace), []byte(whitespace)) {
		t.Errorf("Expected JSONBytesEqual to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if JSONBytesEqual([]byte(noWhitespaceDiff), []byte(whitespaceDiff)) {
		t.Errorf("Expected JSONBytesEqual to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestSecondJSONUnlessEquivalent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		oldPolicy string
		newPolicy string
		want      string
	}{
		{
			name: "new in random order",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:DescribeKey",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource",
        "kms:CreateKey",
        "kms:Get*",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name: "actual change",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:Describe*",
        "kms:List*",
        "kms:ScheduleKeyDeletion",
        "kms:Get*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:Describe*",
        "kms:List*",
        "kms:ScheduleKeyDeletion",
        "kms:Get*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name:      "empty old",
			oldPolicy: "",
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name: "empty new",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: "",
			want:      "",
		},
	}

	for _, v := range testCases {
		got, err := SecondJSONUnlessEquivalent(v.oldPolicy, v.newPolicy)

		if err != nil {
			t.Fatalf("unexpected error with test case %s: %s", v.name, err)
		}

		if got != v.want {
			t.Fatalf("for test case %s, got %s, wanted %s", v.name, got, v.want)
		}
	}
}

func TestNormalizeJSONOrYAMLString(t *testing.T) {
	t.Parallel()

	var err error
	var actual string

	validNormalizedJSON := `{"abc":"1"}`
	actual, err = NormalizeJSONOrYAMLString(validNormalizedJSON)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedJSON {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedJSON)
	}

	validNormalizedYaml := `abc: 1
`
	actual, err = NormalizeJSONOrYAMLString(validNormalizedYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedYaml)
	}
}

func TestSuppressEquivalentJSONDiffsWhitespaceAndNoWhitespace(t *testing.T) {
	t.Parallel()

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

func TestSuppressEquivalentJSONOrYAMLDiffs(t *testing.T) {
	t.Parallel()

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

func TestLegacyPolicyNormalize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Input    string
		Expected string
		Error    bool
	}{
		{
			Name:     "basic",
			Input:    `{"Statement":{"Action":"*","Effect":"Allow","Resource":"*"},"Version":"2012-10-17"}`,
			Expected: `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`,
			Error:    false,
		},
		{
			Name: "normalWhitespace",
			Input: `{
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  },
  "Version": "2012-10-17"
}
`,
			Expected: `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`,
			Error:    false,
		},
		{
			Name: "badJSON",
			Input: `{
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
  "Version": "2012-10-17"
}
`,
			Expected: `{
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
  "Version": "2012-10-17"
}
`,
			Error: true,
		},
		{
			Name: "principal",
			Input: `{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""    }
  ],
  "Version": "2012-10-17"
}
`,
			Expected: `{"Version":"2012-10-17","Statement":[{"Action":"sts:AssumeRole","Effect":"Allow","Principal":{"Service":"s3.amazonaws.com"},"Sid":""}]}`,
			Error:    false,
		},
		{
			Name: "id",
			Input: `{
  "Id": "Kygo",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""    }
  ],
  "Version": "2012-10-17"
}
`,
			Expected: `{"Version":"2012-10-17","Id":"Kygo","Statement":[{"Action":"sts:AssumeRole","Effect":"Allow","Principal":{"Service":"s3.amazonaws.com"},"Sid":""}]}`,
			Error:    false,
		},
		{
			Name: "newOrder",
			Input: `{
  "Id": "Kygo",
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""    }
  ]
}
`,
			Expected: `{"Version":"2012-10-17","Id":"Kygo","Statement":[{"Action":"sts:AssumeRole","Effect":"Allow","Principal":{"Service":"s3.amazonaws.com"},"Sid":""}]}`,
			Error:    false,
		},
		{
			Name: "complex1",
			Input: `{
"Id": "kms-tf-1",
"Version": "2012-10-17",
"Statement": [
  {
    "Sid": "Enable IAM User Permissions",
    "Effect": "Allow",
    "Principal": {
      "AWS": [
        "arn:aws:iam::012345678901:role/felixjaehn",
        "arn:aws:iam::012345678901:role/garethemery",
        "arn:aws:iam::012345678901:role/kidnap",
        "arn:aws:iam::012345678901:role/paulvandyk",
        "arn:aws:iam::012345678901:role/tinlicker"
      ]
    },
    "Action": [
      "kms:Describe*",
      "kms:Get*",
      "kms:List*",
      "kms:CreateKey",
      "kms:DescribeKey",
      "kms:ScheduleKeyDeletion",
      "kms:TagResource",
      "kms:UntagResource"
    ],
    "Resource": "*"
  }
]
}`,
			Expected: `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Action":["kms:Describe*","kms:Get*","kms:List*","kms:CreateKey","kms:DescribeKey","kms:ScheduleKeyDeletion","kms:TagResource","kms:UntagResource"],"Effect":"Allow","Principal":{"AWS":["arn:aws:iam::012345678901:role/felixjaehn","arn:aws:iam::012345678901:role/garethemery","arn:aws:iam::012345678901:role/kidnap","arn:aws:iam::012345678901:role/paulvandyk","arn:aws:iam::012345678901:role/tinlicker"]},"Resource":"*","Sid":"Enable IAM User Permissions"}]}`,
			Error:    false,
		},
		{
			Name: "complex2",
			Input: `{
"Id": "kms-tf-1",
"ZedsDead": "StillWorse",
"Version": "2012-10-17",
"Statement": [
  {
    "Sid": "Enable IAM User Permissions"
  }
]
}`,
			Expected: `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions"}],"ZedsDead":"StillWorse"}`,
			Error:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			p, err := LegacyPolicyNormalize(tc.Input)

			if tc.Error {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if p != tc.Expected {
				t.Errorf("expected %s, got: %s", tc.Expected, p)
			}
		})
	}
}
