package sns

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
		"account_id": {
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

			valid, err := policyHasValidAWSPrincipals(testcase.json)

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
		"role_arn": {
			value: "arn:aws:iam::123456789012:role/role-name", // lintignore:AWSAT005
			valid: true,
		},
		"root_arn": {
			value: "arn:aws:iam::123456789012:root", // lintignore:AWSAT005
			valid: true,
		},
		"account_id": {
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

			a := isValidAWSPrincipal(testcase.value)

			if e := testcase.valid; a != e {
				t.Fatalf("expected %t, got %t", e, a)
			}
		})
	}
}
