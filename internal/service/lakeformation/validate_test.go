// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflf "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidPrincipal(t *testing.T) {
	t.Parallel()

	v := ""
	_, errors := tflf.ValidPrincipal(v, names.AttrARN)
	if len(errors) == 0 {
		t.Fatalf("%q should not be validated as a principal %d: %q", v, len(errors), errors)
	}

	validNames := []string{
		"IAM_ALLOWED_PRINCIPALS",     // Special principal
		"123456789012:IAMPrincipals", // Special principal, Example Account ID (Valid looking but not real)
		acctest.Ct12Digit,            // lintignore:AWSAT005          // Example Account ID (Valid looking but not real)
		"111122223333",               // lintignore:AWSAT005          // Example Account ID (Valid looking but not real)
		"arn:aws-us-gov:iam::357342307427:role/tf-acc-test-3217321001347236965",          // lintignore:AWSAT005          // IAM Role
		"arn:aws:iam::123456789012:user/David",                                           // lintignore:AWSAT005          // IAM User
		"arn:aws:iam::123456789012:federated-user/David",                                 // lintignore:AWSAT005          // IAM Federated User
		"arn:aws-us-gov:iam:us-west-2:357342307427:role/tf-acc-test-3217321001347236965", // lintignore:AWSAT003,AWSAT005 // Non-global IAM Role?
		"arn:aws:iam:us-east-1:123456789012:user/David",                                  // lintignore:AWSAT003,AWSAT005 // Non-global IAM User?
		"arn:aws:iam::111122223333:saml-provider/idp1:group/data-scientists",             // lintignore:AWSAT005          // SAML group
		"arn:aws:iam::111122223333:saml-provider/idp1:user/Paul",                         // lintignore:AWSAT005          // SAML user
		"arn:aws:quicksight:us-east-1:111122223333:group/default/data_scientists",        // lintignore:AWSAT003,AWSAT005 // quicksight group
		"arn:aws:organizations::111122223333:organization/o-abcdefghijkl",                // lintignore:AWSAT005          // organization
		"arn:aws:organizations::111122223333:ou/o-abcdefghijkl/ou-ab00-cdefgh",           // lintignore:AWSAT005          // ou
	}
	for _, v := range validNames {
		_, errors := tflf.ValidPrincipal(v, names.AttrARN)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid principal: %q", v, errors)
		}
	}

	invalidNames := []string{
		"IAM_NOT_ALLOWED_PRINCIPALS", // doesn't exist
		names.AttrARN,
		"1234567890125",               //not an account id
		"IAMPrincipals",               // incorrect representation
		"1234567890125:IAMPrincipals", // incorrect representation, account id invalid length
		"1234567890125:IAMPrincipal",
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
	}
	for _, v := range invalidNames {
		_, errors := tflf.ValidPrincipal(v, names.AttrARN)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid principal", v)
		}
	}
}

func TestValidCatalogID(t *testing.T) {
	t.Parallel()

	validCatalogIDs := []string{
		acctest.Ct12Digit, // Standard AWS account ID
		"111122223333",    // Another valid account ID
		acctest.Ct12Digit + ":s3tablescatalog/my-table-bucket",   // S3 Tables catalog ID
		"111122223333:s3tablescatalog/test-bucket",               // Another S3 Tables catalog ID
		acctest.Ct12Digit + ":s3tablescatalog/bucket.with.dots",  // S3 Tables catalog ID with dots
		acctest.Ct12Digit + ":s3tablescatalog/bucket-with-dash",  // S3 Tables catalog ID with dashes
		acctest.Ct12Digit + ":s3tablescatalog/bucket_with_under", // S3 Tables catalog ID with underscores
	}

	for _, v := range validCatalogIDs {
		_, errors := tflf.ValidCatalogID(v, names.AttrCatalogID)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid catalog ID: %q", v, errors)
		}
	}

	invalidCatalogIDs := []string{
		"",                                      // Empty string
		"12345678901",                           // Too short account ID
		"1234567890123",                         // Too long account ID
		"12345678901a",                          // Invalid account ID with letter
		acctest.Ct12Digit + ":invalid/format",   // Invalid format
		acctest.Ct12Digit + ":s3tablescatalog/", // Missing bucket name
		acctest.Ct12Digit + ":s3tablescatalog/-invalid", // Bucket name starts with dash
		acctest.Ct12Digit + ":s3tablescatalog/invalid-", // Bucket name ends with dash
		acctest.Ct12Digit + ":s3tablescatalog/.invalid", // Bucket name starts with dot
		acctest.Ct12Digit + ":s3tablescatalog/invalid.", // Bucket name ends with dot
		acctest.Ct12Digit + ":s3tablescatalog/_invalid", // Bucket name starts with underscore
		acctest.Ct12Digit + ":s3tablescatalog/invalid_", // Bucket name ends with underscore
		"12345678901a:s3tablescatalog/valid-bucket",     // Invalid account ID in S3 Tables format
		acctest.Ct12Digit + ":notstables/bucket",        // Wrong service type
		"not-account:s3tablescatalog/bucket",            // Invalid account format
	}

	for _, v := range invalidCatalogIDs {
		_, errors := tflf.ValidCatalogID(v, names.AttrCatalogID)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid catalog ID", v)
		}
	}
}
