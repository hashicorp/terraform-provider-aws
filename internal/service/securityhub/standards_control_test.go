// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestStandardsControlARNToStandardsSubscriptionARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexache.MustCompile(`expected service securityhub`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexache.MustCompile(`expected at least 3 resource parts`),
		},
		{
			TestName:    "valid ARN",
			InputARN:    "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1",  //lintignore:AWSAT003,AWSAT005
			ExpectedARN: "arn:aws:securityhub:us-west-2:1234567890:subscription/cis-aws-foundations-benchmark/v/1.2.0", //lintignore:AWSAT003,AWSAT005
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfsecurityhub.StandardsControlARNToStandardsSubscriptionARN(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}

func testAccStandardsControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var standardsControl types.StandardsControl
	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlExists(ctx, resourceName, &standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_id", "CIS.1.10"),
					resource.TestCheckResourceAttr(resourceName, "control_status", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "control_status_updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "IAM password policies can prevent the reuse of a given password by the same user. It is recommended that the password policy prevent the reuse of passwords."),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", ""),
					resource.TestCheckResourceAttr(resourceName, "related_requirements.0", "CIS AWS Foundations 1.10"),
					resource.TestCheckResourceAttrSet(resourceName, "remediation_url"),
					resource.TestCheckResourceAttr(resourceName, "severity_rating", "LOW"),
					resource.TestCheckResourceAttr(resourceName, "title", "Ensure IAM password policy prevents password reuse"),
				),
			},
		},
	})
}

func testAccStandardsControl_disabledControlStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var standardsControl types.StandardsControl
	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlConfig_disabledStatus(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlExists(ctx, resourceName, &standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", "We handle password policies within Okta"),
				),
			},
		},
	})
}

func testAccStandardsControl_enabledControlStatusAndDisabledReason(t *testing.T) {
	ctx := acctest.Context(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccStandardsControlConfig_enabledStatus(),
				ExpectError: regexache.MustCompile("InvalidInputException: DisabledReason should not be given for action other than disabling control"),
			},
		},
	})
}

func testAccCheckStandardsControlExists(ctx context.Context, n string, v *types.StandardsControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		standardsSubscriptionARN, err := tfsecurityhub.StandardsControlARNToStandardsSubscriptionARN(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfsecurityhub.FindStandardsControlByTwoPartKey(ctx, conn, standardsSubscriptionARN, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStandardsControlConfig_basic() string {
	return acctest.ConfigCompose(testAccStandardsSubscriptionConfig_basic, `
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.10", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
}
`)
}

func testAccStandardsControlConfig_disabledStatus() string {
	return acctest.ConfigCompose(testAccStandardsSubscriptionConfig_basic, `
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.11", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "DISABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}

func testAccStandardsControlConfig_enabledStatus() string {
	return acctest.ConfigCompose(testAccStandardsSubscriptionConfig_basic, `
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.12", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}
