// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	content1 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	content2 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "s3:*", "Resource": "*"}}`
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_required(rName, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile("policy/"+organizationIDRegexPattern+"/service_control_policy/p-[0-9a-z]{8}")),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				Config: testAccPolicyConfig_required(rName, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_organizations_policy.test1"
	resourceName2 := "aws_organizations_policy.test2"
	resourceName3 := "aws_organizations_policy.test3"
	resourceName4 := "aws_organizations_policy.test4"
	resourceName5 := "aws_organizations_policy.test5"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_concurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName1, &policy1),
					testAccCheckPolicyExists(ctx, t, resourceName2, &policy2),
					testAccCheckPolicyExists(ctx, t, resourceName3, &policy3),
					testAccCheckPolicyExists(ctx, t, resourceName4, &policy4),
					testAccCheckPolicyExists(ctx, t, resourceName5, &policy5),
				),
			},
		},
	})
}

func testAccPolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccPolicyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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

func testAccPolicy_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	content := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyNoDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_skipDestroy(rName, content),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile("policy/"+organizationIDRegexPattern+"/service_control_policy/p-.+$")),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.CheckSDKResourceDisappears(ctx, t, tforganizations.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPolicy_type_AI_OPT_OUT(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out_syntax.html
	AiOptOutPolicyContent := `{ "services": { "rekognition": { "opt_out_policy": { "@@assign": "optOut" } }, "lex": { "opt_out_policy": { "@@assign": "optIn" } } } }`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, AiOptOutPolicyContent, string(awstypes.PolicyTypeAiservicesOptOutPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, backupPolicyContent, string(awstypes.PolicyTypeBackupPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, serviceControlPolicyContent, string(awstypes.PolicyTypeServiceControlPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				Config: testAccPolicyConfig_required(rName, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	tagPolicyContent := `{ "tags": { "Product": { "tag_key": { "@@assign": "Product" }, "enforced_for": { "@@assign": [ "ec2:instance" ] } } } }`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, tagPolicyContent, string(awstypes.PolicyTypeTagPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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

func testAccPolicy_type_SecurityHub(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_security_hub_syntax.html
	inspectorPolicyContent := `{
    "securityhub" : {
      "enable_in_regions" : {
        "@@assign" : [
          "ALL_SUPPORTED"
        ]
      },
      "disable_in_regions" : {
        "@@assign" : []
      }
    }
}`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, inspectorPolicyContent, string(awstypes.PolicyTypeSecurityhubPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeSecurityhubPolicy)),
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

func testAccPolicy_type_Inspector(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_inspector_syntax.html
	//lintignore:AWSAT003
	inspectorPolicyContent := `{
    "inspector" : {
      "enablement" : {
        "ec2_scanning" : {
          "enable_in_regions" : {
            "@@assign" : ["us-east-1", "us-west-2"]
          },
          "disable_in_regions" : {
            "@@assign" : ["eu-west-1"]
          }
        }
      }
    }
}`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, inspectorPolicyContent, string(awstypes.PolicyTypeInspectorPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeInspectorPolicy)),
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

func testAccPolicy_type_UpgradeRollout(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_upgrade_syntax.html
	upgradeRolloutPolicyContent := `{
      "upgrade_rollout" : {
        "default" : {
          "patch_order" : {
            "@@assign" : "last"
          }
        },
        "tags" : {
          "my_patch_order_tag" : {
            "tag_values" : {
              "tag1" : {
                "patch_order" : {
                  "@@assign" : "first"
                }
              },
              "tag2" : {
                "patch_order" : {
                  "@@assign" : "second"
                }
              },
              "tag3" : {
                "patch_order" : {
                  "@@assign" : "last"
                }
              }
            }
          }
        }
      }
}`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, upgradeRolloutPolicyContent, string(awstypes.PolicyTypeUpgradeRolloutPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeUpgradeRolloutPolicy)),
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

func testAccPolicy_type_S3(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_s3_syntax.html
	s3PolicyContent := `{
    "s3_attributes": {
        "public_access_block_configuration": {
            "@@assign": "all"
        }
    }
}`

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type(rName, s3PolicyContent, string(awstypes.PolicyTypeS3Policy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeS3Policy)),
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

func testAccPolicy_type_Bedrock(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.Policy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_type_Bedrock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.PolicyTypeBedrockPolicy)),
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
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

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_policy" {
				continue
			}

			_, err := tforganizations.FindPolicyByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
func testAccCheckPolicyNoDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

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

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string, v *awstypes.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

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
}
`, description, rName)
}

func testAccPolicyConfig_required(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q
}
`, strconv.Quote(content), rName)
}

func testAccPolicyConfig_concurrent(rName string) string {
	return fmt.Sprintf(`
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
}
`, rName)
}

func testAccPolicyConfig_type(rName, content, policyType string) string {
	return fmt.Sprintf(`
resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q
  type    = %[3]q
}
`, strconv.Quote(content), rName, policyType)
}

func testAccPolicyConfig_type_Bedrock(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
  }

  contextual_grounding_policy_config {
    filters_config {
      threshold = 0.4
      type      = "GROUNDING"
    }
  }

  sensitive_information_policy_config {
    pii_entities_config {
      action = "BLOCK"
      type   = "NAME"
    }
    pii_entities_config {
      action = "BLOCK"
      type   = "DRIVER_ID"
    }
    pii_entities_config {
      action = "ANONYMIZE"
      type   = "USERNAME"
    }
    regexes_config {
      action      = "BLOCK"
      description = "example regex"
      name        = "regex_example"
      pattern     = "^\\d{3}-\\d{2}-\\d{4}$"
    }
  }

  topic_policy_config {
    topics_config {
      name       = "investment_topic"
      examples   = ["Where should I invest my money ?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns ."
    }
  }

  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
}

resource "aws_bedrock_guardrail_version" "test" {
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
}

resource "aws_organizations_policy" "test" {
  # Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_bedrock_syntax.html
  content = jsonencode({
    "bedrock" : {
      "guardrail_inference" : {
        (data.aws_region.current.region) : {
          "config_1" : {
            "identifier" : {
              "@@assign" : "${aws_bedrock_guardrail.test.guardrail_arn}:${aws_bedrock_guardrail_version.test.version}"
            },
            "input_tags" : {
              "@@assign" : "honor"
            }
          }
        }
      }
    }
  })
  name = %[1]q
  type = "BEDROCK_POLICY"
}
`, rName)
}

func testAccPolicyConfig_skipDestroy(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_policy" "test" {
  content = %[1]s
  name    = %[2]q

  skip_destroy = true
}

`, strconv.Quote(content), rName)
}

const testAccPolicyConfig_managed = `
resource "aws_organizations_policy" "test" {
  name = "FullAWSAccess"
}
`
