// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
)

// The per-resource account tests use the IAM role in TF_ACC_ASSUME_ROLE_ARN:
// its account is the resources' account and its name the role to assume there.
// The role must be assumable from the credentials the tests run with.
func testAccAccountRole(t *testing.T) (accountID, roleName string) {
	t.Helper()

	acctest.PreCheckAssumeRoleARN(t)
	v, err := arn.Parse(os.Getenv(envvar.AccAssumeRoleARN))
	if err != nil {
		t.Fatalf("parsing %s: %s", envvar.AccAssumeRoleARN, err)
	}

	return v.AccountID, strings.TrimPrefix(v.Resource, "role/")
}

func TestAccProvider_Account_none(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_account(rName, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The account of every resource is recorded, also when it is not configured.
					acctest.CheckResourceAttrAccountID(ctx, "aws_ssm_parameter.test", "account"),
					acctest.CheckResourceAttrAccountID(ctx, "aws_vpc_security_group_ingress_rule.test", "account"),
					acctest.CheckResourceAttrAccountID(ctx, "data.aws_caller_identity.test", "account"),
					acctest.CheckResourceAttrAccountID(ctx, "data.aws_caller_identity.test", "account_id"),
				),
			},
		},
	})
}

func TestAccProvider_Account_other(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accountID, roleName := testAccAccountRole(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// Without a block that applies to the account there is no way to reach it.
				Config:      testAccProviderConfig_account(rName, accountID, ""),
				ExpectError: regexache.MustCompile(`no assume_role_in_account block`),
				SkipFunc:    func() (bool, error) { return accountID == acctest.AccountID(ctx), nil },
			},
			{
				Config: testAccProviderConfig_account(rName, accountID, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_ssm_parameter.test", "account", accountID),
					resource.TestCheckResourceAttr("aws_vpc_security_group_ingress_rule.test", "account", accountID),
					resource.TestCheckResourceAttr("data.aws_caller_identity.test", "account_id", accountID),
					resource.TestMatchResourceAttr("aws_ssm_parameter.test", "arn", regexache.MustCompile(`^arn:[^:]+:ssm:[^:]+:\d{12}:parameter/`)),
				),
			},
			{
				// Another session name is another role session, and no change to any resource.
				Config: testAccProviderConfig_account(rName, accountID, roleName, "session_name = \"tf-acc-test\""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccProvider_Account_change(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accountID, roleName := testAccAccountRole(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_account(rName, "", ""),
			},
			{
				// A change of account replaces the resources.
				Config:   testAccProviderConfig_account(rName, accountID, roleName),
				SkipFunc: func() (bool, error) { return accountID == acctest.AccountID(ctx), nil },
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_ssm_parameter.test", plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction("aws_vpc_security_group_ingress_rule.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_ssm_parameter.test", "account", accountID),
					resource.TestCheckResourceAttr("aws_vpc_security_group_ingress_rule.test", "account", accountID),
				),
			},
			{
				// A resource in another account has to say so.
				Config:      testAccProviderConfig_account(rName, "", roleName),
				SkipFunc:    func() (bool, error) { return accountID == acctest.AccountID(ctx), nil },
				ExpectError: regexache.MustCompile(`which is not the provider's account`),
			},
		},
	})
}

// testAccProviderConfig_account manages Plugin SDK v2 resources, a Plugin Framework resource
// and a Plugin Framework data source in an account.
func testAccProviderConfig_account(rName, accountID, roleName string, roleArguments ...string) string {
	var provider, account string
	if roleName != "" {
		//lintignore:AT004
		provider = fmt.Sprintf(`
provider "aws" {
  assume_role_in_account {
    role_name = %[1]q
    %[2]s
  }
}
`, roleName, strings.Join(roleArguments, "\n"))
	}
	if accountID != "" {
		account = fmt.Sprintf("account = %q", accountID)
	}

	return provider + fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "test"
  %[2]s
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  %[2]s

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_vpc.test.default_security_group_id
  cidr_ipv4         = "10.0.0.0/8"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  %[2]s
}

data "aws_caller_identity" "test" {
  %[2]s
}
`, rName, account)
}
