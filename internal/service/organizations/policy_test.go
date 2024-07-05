// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	content1 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	content2 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "s3:*", "Resource": "*"}}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_required(rName, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "organizations", regexache.MustCompile("policy/o-.+/service_control_policy/p-.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				Config: testAccPolicyConfig_required(rName, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/5073
func testAccPolicy_concurrent(t *testing.T) {
	ctx := acctest.Context(t)
	var policy1, policy2, policy3, policy4, policy5 awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_organizations_policy.test1"
	resourceName2 := "aws_organizations_policy.test2"
	resourceName3 := "aws_organizations_policy.test3"
	resourceName4 := "aws_organizations_policy.test4"
	resourceName5 := "aws_organizations_policy.test5"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_concurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName1, &policy1),
					testAccCheckPolicyExists(ctx, resourceName2, &policy2),
					testAccCheckPolicyExists(ctx, resourceName3, &policy3),
					testAccCheckPolicyExists(ctx, resourceName4, &policy4),
					testAccCheckPolicyExists(ctx, resourceName5, &policy5),
				),
			},
		},
	})
}

func testAccPolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccPolicyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func testAccPolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
			{
				Config: testAccPolicyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPolicyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccPolicy_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	content := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyNoDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_skipDestroy(rName, content),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "organizations", regexache.MustCompile("policy/o-.+/service_control_policy/p-.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
		},
	})
}

func testAccPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPolicy_type_AI_OPT_OUT(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out_syntax.html
	AiOptOutPolicyContent := `{ "services": { "rekognition": { "opt_out_policy": { "@@assign": "optOut" } }, "lex": { "opt_out_policy": { "@@assign": "optIn" } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, AiOptOutPolicyContent, string(awstypes.PolicyTypeAiservicesOptOutPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeAiservicesOptOutPolicy)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func testAccPolicy_type_Backup(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_backup_syntax.html
	backupPolicyContent := fmt.Sprintf(`{
   "plans":{
      "PII_Backup_Plan":{
         "regions":{
            "@@assign":[
               "%[1]s",
               "%[2]s"
            ]
         },
         "rules":{
            "Hourly":{
               "schedule_expression":{
                  "@@assign":"cron(0 5/1 ? * * *)"
               },
               "start_backup_window_minutes":{
                  "@@assign":"480"
               },
               "complete_backup_window_minutes":{
                  "@@assign":"10080"
               },
               "lifecycle":{
                  "move_to_cold_storage_after_days":{
                     "@@assign":"180"
                  },
                  "delete_after_days":{
                     "@@assign":"270"
                  }
               },
               "target_backup_vault_name":{
                  "@@assign":"FortKnox"
               },
               "copy_actions":{
                  "arn:%[3]s:backup:%[1]s:$account:backup-vault:secondary_vault":{
                     "target_backup_vault_arn":{
                        "@@assign":"arn:%[3]s:backup:%[1]s:$account:backup-vault:secondary_vault"
                     },
                     "lifecycle":{
                        "delete_after_days":{
                           "@@assign":"100"
                        },
                        "move_to_cold_storage_after_days":{
                           "@@assign":"10"
                        }
                     }
                  }
               }
            }
         },
         "selections":{
            "tags":{
               "datatype":{
                  "iam_role_arn":{
                     "@@assign":"arn:%[3]s:iam::$account:role/MyIamRole"
                  },
                  "tag_key":{
                     "@@assign":"dataType"
                  },
                  "tag_value":{
                     "@@assign":[
                        "PII",
                        "RED"
                     ]
                  }
               }
            }
         }
      }
   }
}`, acctest.AlternateRegion(), acctest.Region(), acctest.Partition())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, backupPolicyContent, string(awstypes.PolicyTypeBackupPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeBackupPolicy)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func testAccPolicy_type_SCP(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, serviceControlPolicyContent, string(awstypes.PolicyTypeServiceControlPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				Config: testAccPolicyConfig_required(rName, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func testAccPolicy_type_Tag(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	tagPolicyContent := `{ "tags": { "Product": { "tag_key": { "@@assign": "Product" }, "enforced_for": { "@@assign": [ "ec2:instance" ] } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, tagPolicyContent, string(awstypes.PolicyTypeTagPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeTagPolicy)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func testAccPolicy_importManagedPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_organizations_policy.test"

	resourceID := "p-FullAWSAccess"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_managedSetup,
			},
			{
				Config:        testAccPolicyConfig_managed,
				ResourceName:  resourceName,
				ImportStateId: resourceID,
				ImportState:   true,
				ExpectError:   regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("AWS-managed Organizations policy (%s) cannot be imported.", resourceID))),
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_policy" {
				continue
			}

			_, err := tforganizations.FindPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Organizations Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckPolicyNoDestroy is a variant of the CheckDestroy function to be used when
// skip_destroy is true and the policy should still exist after destroy completes
func testAccCheckPolicyNoDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_policy" {
				continue
			}

			_, err := tforganizations.FindPolicyByID(ctx, conn, rs.Primary.ID)

			if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) {
				// The organization was destroyed, so we can safely assume the policy
				// skipped during destruction was as well
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *awstypes.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		output, err := tforganizations.FindPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPolicyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  description = %[1]q
  name        = %[2]q

  depends_on = [aws_organizations_organization.test]
}
`, description, rName)
}

func testAccPolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = %[1]q

  depends_on = [aws_organizations_organization.test]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPolicyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = %[1]q

  depends_on = [aws_organizations_organization.test]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPolicyConfig_required(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q

  depends_on = [aws_organizations_organization.test]
}
`, strconv.Quote(content), rName)
}

func testAccPolicyConfig_concurrent(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test1" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "cloudtrail:StopLogging",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s1"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test2" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "ec2:DeleteFlowLogs",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s2"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test3" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "logs:DeleteLogGroup",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s3"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test4" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "config:DeleteConfigRule",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s4"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test5" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "iam:DeleteRolePermissionsBoundary",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s5"

  depends_on = [aws_organizations_organization.test]
}
`, rName)
}

func testAccPolicyConfig_type(rName, content, policyType string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q
  type    = %[3]q

  depends_on = [aws_organizations_organization.test]
}
`, strconv.Quote(content), rName, policyType)
}

func testAccPolicyConfig_skipDestroy(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q

  skip_destroy = true

  depends_on = [aws_organizations_organization.test]
}
`, strconv.Quote(content), rName)
}

const testAccPolicyConfig_managedSetup = `
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}
`

const testAccPolicyConfig_managed = `
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}

resource "aws_organizations_policy" "test" {
  name = "FullAWSAccess"
}
`
