// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"encoding/json"
	"reflect"
	"slices"
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
			value: acctest.Ct12Digit,
			valid: true,
		},
		"unique_id": {
			value: "AROAS5MHDZS6NEXAMPLE",
			valid: false,
		},
	}

	for name, testcase := range testcases {
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
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: "1"},
				{Test: "StringLike", Variable: "s3:prefix", Values: "one/"},
			},
			want: []byte(`{"ArnNotLike":{"aws:PrincipalArn":"1"},"StringLike":{"s3:prefix":"one/"}}`),
		},
		"multiple condition multiple values": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: []string{"1", "2"}},
				{Test: "StringLike", Variable: "s3:prefix", Values: []string{"one/", "two/"}},
			},
			want: []byte(`{"ArnNotLike":{"aws:PrincipalArn":["1","2"]},"StringLike":{"s3:prefix":["one/","two/"]}}`),
		},
		"multiple condition mixed value lengths": {
			cs: tfiam.IAMPolicyStatementConditionSet{
				{Test: "ArnNotLike", Variable: "aws:PrincipalArn", Values: "1"},
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

func TestIAMPolicyStatementPrincipalSet_UnmarshalJSON(t *testing.T) { // nosemgrep:ci.iam-in-func-name
	t.Parallel()

	testcases := map[string]struct {
		b       []byte
		want    tfiam.IAMPolicyStatementPrincipalSet
		wantErr bool
	}{
		"wildcard, wildcard": {
			b: []byte(`{"*": "*"}`),
			want: tfiam.IAMPolicyStatementPrincipalSet{
				{Type: "*", Identifiers: "*"},
			},
		},
		"single key, wildcard": {b: []byte(`{"AWS": "*"}`),
			want: tfiam.IAMPolicyStatementPrincipalSet{
				{Type: "AWS", Identifiers: "*"},
			},
		},
		"single key, single value": {
			b: []byte(`{"AWS": "111122223333"}`),
			want: tfiam.IAMPolicyStatementPrincipalSet{
				{Type: "AWS", Identifiers: "111122223333"},
			},
		},
		"single key, multiple value": {
			b: []byte(`{"AWS": ["111122223333", "444455556666"]}`),
			want: tfiam.IAMPolicyStatementPrincipalSet{
				{Type: "AWS", Identifiers: []string{"111122223333", "444455556666"}},
			},
		},
		"multiple key": {
			b: []byte(`{
  "AWS": "111122223333",
  "CanonicalUser": "abcdef123456"
}`,
			),
			want: tfiam.IAMPolicyStatementPrincipalSet{
				{Type: "AWS", Identifiers: "111122223333"},
				{Type: "CanonicalUser", Identifiers: "abcdef123456"},
			},
		},
		"invalid json": {
			b:       []byte(`{{{"*"}`),
			wantErr: true,
		},
		"invalid data type": {
			b:       []byte(`["AWS"]`),
			wantErr: true,
		},
		"invalid value type": {
			b:       []byte(`{"AWS": 0}`),
			wantErr: true,
		},
		"invalid array value type": {
			b:       []byte(`{"AWS": ["111122223333", 0]}`),
			wantErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got tfiam.IAMPolicyStatementPrincipalSet
			err := got.UnmarshalJSON(tc.b)
			if (err != nil) != tc.wantErr {
				t.Errorf("IAMPolicyStatementPrincipalSet.UnmarshalJSON() error = %v, wantErr %t", err, tc.wantErr)
				return
			}
			// Sort both slices by Type to ensure deterministic comparison
			// (JSON object key iteration order is non-deterministic)
			sortByType := func(a, b tfiam.IAMPolicyStatementPrincipal) int {
				if a.Type < b.Type {
					return -1
				}
				if a.Type > b.Type {
					return 1
				}
				return 0
			}
			slices.SortFunc(got, sortByType)
			slices.SortFunc(tc.want, sortByType)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("IAMPolicyStatementPrincipalSet.UnmarshalJSON() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestIAMPolicyDoc_UnmarshalJSON(t *testing.T) { // nosemgrep:ci.iam-in-func-name
	t.Parallel()

	testcases := map[string]struct {
		b       string
		want    *tfiam.IAMPolicyDoc
		wantErr bool
	}{
		"statement array": {
			b: `{"Version": "2012-10-17", "Statement": [{"Sid": "One", "Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"}, {"Sid": "Two", "Effect": "Deny", "Action": ["s3:PutObject", "s3:DeleteObject"], "Resource": "*"}]}`,
			want: &tfiam.IAMPolicyDoc{
				Version: "2012-10-17",
				Statements: []*tfiam.IAMPolicyStatement{
					{Sid: "One", Effect: "Allow", Actions: "s3:GetObject", Resources: "*"},
					{Sid: "Two", Effect: "Deny", Actions: []any{"s3:PutObject", "s3:DeleteObject"}, Resources: "*"},
				},
			},
		},
		"statement object": {
			b: `{"Version": "2012-10-17", "Id": "Example", "Statement": {"Effect": "Allow", "Action": ["acm:DescribeCertificate", "acm:ListCertificates"], "Resource": "*"}}`,
			want: &tfiam.IAMPolicyDoc{
				Version: "2012-10-17",
				Id:      "Example",
				Statements: []*tfiam.IAMPolicyStatement{
					{Effect: "Allow", Actions: []any{"acm:DescribeCertificate", "acm:ListCertificates"}, Resources: "*"},
				},
			},
		},
		"statement object with whitespace": {
			b: `{"Statement":
				{"Effect": "Allow", "Action": "*", "Resource": "*"}}`,
			want: &tfiam.IAMPolicyDoc{
				Statements: []*tfiam.IAMPolicyStatement{
					{Effect: "Allow", Actions: "*", Resources: "*"},
				},
			},
		},
		"empty statement array": {
			b: `{"Version": "2012-10-17", "Statement": []}`,
			want: &tfiam.IAMPolicyDoc{
				Version:    "2012-10-17",
				Statements: []*tfiam.IAMPolicyStatement{},
			},
		},
		"no statement": {
			b: `{"Version": "2012-10-17"}`,
			want: &tfiam.IAMPolicyDoc{
				Version: "2012-10-17",
			},
		},
		"null statement": {
			b: `{"Version": "2012-10-17", "Statement": null}`,
			want: &tfiam.IAMPolicyDoc{
				Version: "2012-10-17",
			},
		},
		"invalid json": {
			b:       `{"Statement": {`,
			wantErr: true,
		},
		"invalid statement type": {
			b:       `{"Statement": "Allow"}`,
			wantErr: true,
		},
		"invalid statement object": {
			b:       `{"Statement": {"Principal": ["AWS"]}}`,
			wantErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &tfiam.IAMPolicyDoc{}
			err := json.Unmarshal([]byte(tc.b), got)
			if (err != nil) != tc.wantErr {
				t.Fatalf("json.Unmarshal() error = %v, wantErr %t", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("json.Unmarshal() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestIAMPolicyDoc_mergeSingleStatementObject(t *testing.T) { // nosemgrep:ci.iam-in-func-name
	t.Parallel()

	sources := []string{
		`{"Version": "2012-10-17", "Statement": [{"Sid": "Array", "Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"}]}`,
		`{"Version": "2012-10-17", "Statement": {"Sid": "Object", "Effect": "Allow", "Action": "acm:ListCertificates", "Resource": "*"}}`,
	}

	merged := &tfiam.IAMPolicyDoc{}
	for _, source := range sources {
		doc := &tfiam.IAMPolicyDoc{}
		if err := json.Unmarshal([]byte(source), doc); err != nil {
			t.Fatalf("json.Unmarshal(%s): %s", source, err)
		}
		merged.Merge(doc)
	}

	got, err := json.Marshal(merged)
	if err != nil {
		t.Fatalf("json.Marshal(): %s", err)
	}

	want := `{"Version":"2012-10-17","Statement":[{"Sid":"Array","Effect":"Allow","Action":"s3:GetObject","Resource":"*"},{"Sid":"Object","Effect":"Allow","Action":"acm:ListCertificates","Resource":"*"}]}`
	if string(got) != want {
		t.Errorf("json.Marshal() = %s, want %s", got, want)
	}
}
