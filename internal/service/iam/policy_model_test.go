// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestPolicyHasValidAWSPrincipals(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	testcases := map[string]struct {
		json  string
		valid bool
		err   func(t *testing.T, err error)
	}{
		"single_arn": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "arn:aws:iam::123456789012:role/role-name"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			valid: true,
		},
		names.AttrAccountID: {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "123456789012"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`,
			valid: true,
		},
		"wildcard": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`,
			valid: true,
		},
		"unique_id": {json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "AROAS5MHDZS6NEXAMPLE"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`,
			valid: false,
		},
		"non_AWS_principal": {json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`,
			valid: true,
		},
		"multiple_arns": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": [
          "arn:aws:iam::123456789012:role/role-name",
          "arn:aws:iam::123456789012:role/another-role-name"
        ]
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			valid: true,
		},
		"mixed_principals": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": [
          "arn:aws:iam::123456789012:role/role-name",
          "AROAS5MHDZS6NEXAMPLE"
        ]
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			valid: true,
		},
		"multiple_statements_valid": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "arn:aws:iam::123456789012:role/role-name"
      },
      "Action": "*",
      "Resource": "*"
    },
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "arn:aws:iam::123456789012:role/another-role-name"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			valid: true,
		},
		"multiple_statements_invalid": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "arn:aws:iam::123456789012:role/role-name"
      },
      "Action": "*",
      "Resource": "*"
    },
    {
      "Effect":"Allow",
      "Principal":{
        "AWS": "AROAS5MHDZS6NEXAMPLE"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			valid: false,
		},
		"empty_string": {
			json: "",
			err: func(t *testing.T, err error) {
				if !errs.IsA[*json.SyntaxError](err) {
					t.Fatalf("expected JSON syntax error, got %#v", err)
				}
			},
		},
		"invalid_json": {
			json: `{
  "Statement":[
    {
      "Effect":"Allow"
      "Principal":{
        "AWS": "arn:aws:iam::123456789012:role/role-name"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, // lintignore:AWSAT005
			err: func(t *testing.T, err error) {
				if !errs.IsA[*json.SyntaxError](err) {
					t.Fatalf("expected JSON syntax error, got %#v", err)
				}
			},
		},
	}

	for name, testcase := range testcases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valid, err := tfiam.PolicyHasValidAWSPrincipals(testcase.json)

			if testcase.err == nil {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, not none")
				}
				testcase.err(t, err)
			}

			if a, e := valid, testcase.valid; a != e {
				t.Fatalf("expected %t, got %t", e, a)
			}
		})
	}
}

func TestIsValidAWSPrincipal(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	testcases := map[string]struct {
		value string
		valid bool
	}{
		names.AttrRoleARN: {
			value: "arn:aws:iam::123456789012:role/role-name", // lintignore:AWSAT005
			valid: true,
		},
		"root_arn": {
			value: "arn:aws:iam::123456789012:root", // lintignore:AWSAT005
			valid: true,
		},
		names.AttrAccountID: {
			value: "123456789012",
			valid: true,
		},
		"unique_id": {
			value: "AROAS5MHDZS6NEXAMPLE",
			valid: false,
		},
	}

	for name, testcase := range testcases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := tfiam.IsValidPolicyAWSPrincipal(testcase.value)

			if e := testcase.valid; a != e {
				t.Fatalf("expected %t, got %t", e, a)
			}
		})
	}
}

func TestIAMPolicyStatementConditionSet_MarshalJSON(t *testing.T) { // nosemgrep:ci.iam-in-func-name
	t.Parallel()

	testcases := map[string]struct {
		cs      tfiam.IAMPolicyStatementConditionSet
		want    []byte
		wantErr bool
	}{
		"invalid value type": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: 1},
			},
			wantErr: true,
		},
		"single condition single value": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
			},
			want: []byte(`{"StringLike":{"s3:prefix":"one/"}}`),
		},
		"single condition multiple values": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/"]}}`),
		},
		// Multiple distinct conditions
		"multiple condition single value": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: acctest.Ct1},
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
			},
			want: []byte(`{"ArnNotLike":{"aws:PrincipalArn":"1"},"StringLike":{"s3:prefix":"one/"}}`),
		},
		"multiple condition multiple values": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: []string{acctest.Ct1, acctest.Ct2}},
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
			},
			want: []byte(`{"ArnNotLike":{"aws:PrincipalArn":["1","2"]},"StringLike":{"s3:prefix":["one/","two/"]}}`),
		},
		"multiple condition mixed value lengths": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: acctest.Ct1},
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
			},
			want: []byte(`{"ArnNotLike":{"aws:PrincipalArn":"1"},"StringLike":{"s3:prefix":["one/","two/"]}}`),
		},
		// Multiple conditions with duplicated `test` arguments
		"duplicate condition test single value": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
				{Test: "StringLike", Variable: "s3:versionid", Values: "abc123"},
			},
			want: []byte(`{"StringLike":{"s3:prefix":"one/","s3:versionid":"abc123"}}`),
		},
		"duplicate condition test multiple values": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
				{Test: "StringLike", Variable: "s3:versionid", Values: []string{"abc123", "def456"}},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/"],"s3:versionid":["abc123","def456"]}}`),
		},
		"duplicate condition test mixed value lengths": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
				{Test: "StringLike", Variable: "s3:versionid", Values: []string{"abc123", "def456"}},
			},
			want: []byte(`{"StringLike":{"s3:prefix":"one/","s3:versionid":["abc123","def456"]}}`),
		},
		"duplicate condition test mixed value lengths reversed": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
				{Test: "StringLike", Variable: "s3:versionid", Values: "abc123"},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/"],"s3:versionid":"abc123"}}`),
		},
		// Multiple conditions with duplicated `test` and `variable` arguments
		"duplicate condition test and variable single value": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
				{Test: "StringLike", Variable: "s3:prefix", Values: "two/"},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/"]}}`),
		},
		"duplicate condition test and variable multiple values": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"three/", "four/"}},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/","three/","four/"]}}`),
		},
		"duplicate condition test and variable mixed value lengths": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"three/", "four/"}},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","three/","four/"]}}`),
		},
		"duplicate condition test and variable mixed value lengths reversed": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
				{Test: "StringLike", Variable: "s3:prefix", Values: "three/"},
			},
			want: []byte(`{"StringLike":{"s3:prefix":["one/","two/","three/"]}}`),
		},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.cs.MarshalJSON()
			if (err != nil) != tc.wantErr {
				t.Errorf("IAMPolicyStatementConditionSet.MarshalJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("IAMPolicyStatementConditionSet.MarshalJSON() = %v, want %v", string(got), string(tc.want))
			}
		})
	}
}

func TestPolicyUnmarshalServicePrincipalOrder(t *testing.T) {
	t.Parallel()

	policy1 := `
		  {
			"Action": "sts:AssumeRole",
			"Principal": {
			  "Service": ["lambda.amazonaws.com", "service2.amazonaws.com"]
			},
			"Effect": "Allow",
			"Sid": ""
		  }`
	// Service order is different, but should be the same object for terraform
	policy2 := `
		  {
			"Action": "sts:AssumeRole",
			"Principal": {
			  "Service": ["service2.amazonaws.com", "lambda.amazonaws.com"]
			},
			"Effect": "Allow",
			"Sid": ""
		  }`

	var data1 tfiam.IAMPolicyStatement
	var data2 tfiam.IAMPolicyStatement
	err := json.Unmarshal([]byte(policy1), &data1)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(policy2), &data2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data1, data2) {
		t.Fatalf("should be equal, but was:\n%#v\nVS\n%#v\n", data1, data2)
	}
}
