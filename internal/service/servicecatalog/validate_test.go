// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidSharePrincipal(t *testing.T) {
	t.Parallel()

	v := ""
	_, errors := validSharePrincipal(v, names.AttrARN)
	if len(errors) == 0 {
		t.Fatalf("%q should not be validated as a principal %d: %q", v, len(errors), errors)
	}

	validNames := []string{
		"123456789012", // lintignore:AWSAT005          // Example Account ID (Valid looking but not real)
		"111122223333", // lintignore:AWSAT005          // Example Account ID (Valid looking but not real)
		"arn:aws:organizations::111122223333:organization/o-abcdefghijkl",             // lintignore:AWSAT005    // organization
		"arn:aws:organizations::111122223333:ou/o-abcdefghijkl/ou-ab00-cdefgh",        // lintignore:AWSAT005    // ou
		"arn:aws-us-gov:organizations::111122223333:ou/o-abcdefghijkl/ou-ab00-cdefgh", // lintignore:AWSAT005    // GovCloud ou
	}
	for _, v := range validNames {
		_, errors := validSharePrincipal(v, names.AttrARN)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid principal: %q", v, errors)
		}
	}

	invalidNames := []string{
		"IAM_ALLOWED_PRINCIPALS",     // Special principal
		"IAM_NOT_ALLOWED_PRINCIPALS", // doesn't exist
		names.AttrARN,
		"1234567890125", //not an account id
		"arn:aws",
		"arn:aws:logs",            //lintignore:AWSAT005
		"arn:aws:logs:region:*:*", //lintignore:AWSAT005
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment", // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess",                                 // lintignore:AWSAT005          // not a user or role
		"arn:aws:rds:eu-west-1:123456789012:db:mysql-db",                                   // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                               // lintignore:AWSAT005          // not a user or role
		"arn:aws:events:us-east-1:319201112229:rule/rule_name",                             // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws-us-gov:ec2:us-gov-west-1:123456789012:instance/i-12345678",                // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws-us-gov:s3:::bucket/object",                                                // lintignore:AWSAT005          // not a user or role
		"arn:aws-us-gov:iam::357342307427:role/tf-acc-test-3217321001347236965",            // lintignore:AWSAT005          // IAM Role
		"arn:aws:iam::123456789012:user/David",                                             // lintignore:AWSAT005          // IAM User
		"arn:aws-us-gov:iam:us-west-2:357342307427:role/tf-acc-test-3217321001347236965",   // lintignore:AWSAT003,AWSAT005 // Non-global IAM Role?
		"arn:aws:iam:us-east-1:123456789012:user/David",                                    // lintignore:AWSAT003,AWSAT005 // Non-global IAM User?
		"arn:aws:iam::111122223333:saml-provider/idp1:group/data-scientists",               // lintignore:AWSAT005          // SAML group
		"arn:aws:iam::111122223333:saml-provider/idp1:user/Paul",                           // lintignore:AWSAT005          // SAML user
		"arn:aws:quicksight:us-east-1:111122223333:group/default/data_scientists",          // lintignore:AWSAT003,AWSAT005 // quicksight group
	}
	for _, v := range invalidNames {
		_, errors := validSharePrincipal(v, names.AttrARN)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid principal", v)
		}
	}
}
