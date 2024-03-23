// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcleanrooms "github.com/hashicorp/terraform-provider-aws/internal/service/cleanrooms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCleanRoomsMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_membership.test"

	acctest.AccountID()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_initResources(rName),
			},
			{
				Config: testAccMembershipConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
					resource.TestCheckResourceAttr(resourceName, "query_log_status", TEST_QUERY_LOG_STATUS),
					resource.TestCheckResourceAttr(resourceName, "collaboration_creator_display_name", TEST_CREATOR_DISPLAY_NAME),
					resource.TestCheckResourceAttrSet(resourceName, "collaboration_id"),
					resource.TestCheckResourceAttr(resourceName, "collaboration_name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_result_configuration.0.output_configuration.0.s3.*", map[string]string{
						"bucket":        rName,
						"result_format": TEST_RESULT_FORMAT,
						"key_prefix":    TEST_KEY_PREFIX,
					}),
					resource.TestCheckResourceAttr(resourceName, "member_abilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_abilities.0", "CAN_RECEIVE_RESULTS"),
					resource.TestCheckResourceAttr(resourceName, "payment_configuration.0.query_compute.0.is_responsible", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "query_log_status", TEST_QUERY_LOG_STATUS),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", TEST_TAG),
					acctest.MatchResourceAttrAccountID(resourceName, "collaboration_creator_account_id"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "default_result_configuration.0.role_arn", "iam", regexache.MustCompile("role/"+rName)),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "collaboration_arn", "cleanrooms", "collaboration"),
				),
			},
			{
				Config:                  testAccMembershipConfig_basic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccCleanRoomsMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_initResources(rName),
			},
			{
				Config: testAccMembershipConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcleanrooms.ResourceMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsMembership_mutableProperties(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameSecond := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cleanrooms_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_initDoubledResources(rName, rNameSecond),
			},
			{
				Config: testAccMembershipConfig_mutableProperties(rName, rNameSecond, rName, TEST_QUERY_LOG_STATUS, TEST_RESULT_FORMAT, TEST_KEY_PREFIX, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
				),
			},
			{
				Config: testAccMembershipConfig_mutableProperties(rName, rNameSecond, rNameSecond, "ENABLED", "PARQUET", "updated-prefix", "updated tag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipIsTheSame(resourceName, &membership),
					resource.TestCheckResourceAttr(resourceName, "query_log_status", "ENABLED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_result_configuration.0.output_configuration.0.s3.*", map[string]string{
						"bucket":        rNameSecond,
						"result_format": "PARQUET",
						"key_prefix":    "updated-prefix",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", "updated tag"),
				),
			},
		},
	})
}

func TestAccCleanRoomsMembership_defaultOutputConfigurationWithEmptyAdditionalParameters(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cleanrooms_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_outputConfigurationWithEmptyAdditionalParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
				),
			},
		},
	})
}

func TestAccCleanRoomsMembership_withoutDefaultOutputConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cleanrooms_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_base(rName, TEST_CREATOR_DISPLAY_NAME, "[]", "[\"CAN_QUERY\",\"CAN_RECEIVE_RESULTS\"]", TEST_QUERY_LOG_STATUS, "", TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
				),
			},
		},
	})
}

func TestAccCleanRoomsMembership_addDefaultOutputConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	var membership cleanrooms.GetMembershipOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cleanrooms_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_base(rName, TEST_CREATOR_DISPLAY_NAME, "[]", "[\"CAN_QUERY\",\"CAN_RECEIVE_RESULTS\"]", TEST_QUERY_LOG_STATUS, "", TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
				),
			},
			{
				Config: testAccMembershipConfig_outputConfigurationWithEmptyAdditionalParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMembershipExists(ctx, resourceName, &membership),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_result_configuration.0.output_configuration.0.s3.*", map[string]string{
						"bucket":        rName,
						"result_format": TEST_RESULT_FORMAT,
					}),
				),
			},
		},
	})
}

func testAccCheckMembershipExists(ctx context.Context, name string, membership *cleanrooms.GetMembershipOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameMembership, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameMembership, name, errors.New("not set"))
		}

		client := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		resp, err := client.GetMembership(ctx, &cleanrooms.GetMembershipInput{
			MembershipIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, rs.Primary.ID, err)
		}

		*membership = *resp

		return nil
	}
}

func testAccCheckMembershipDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != tfcleanrooms.ResNameMembership {
				continue
			}

			_, err := conn.GetMembership(ctx, &cleanrooms.GetMembershipInput{
				MembershipIdentifier: aws.String(rs.Primary.ID),
			})

			if err == nil {
				return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameMembership, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckMembershipIsTheSame(name string, membership *cleanrooms.GetMembershipOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return checkMembershipIsTheSame(name, membership, s)
	}
}

func checkMembershipIsTheSame(name string, membership *cleanrooms.GetMembershipOutput, s *terraform.State) error {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not found"))
	}

	if rs.Primary.ID == "" {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not set"))
	}

	if rs.Primary.ID != *membership.Membership.Id {
		return fmt.Errorf("New configured table: %s created instead of updating: %s", rs.Primary.ID, *membership.Membership.Id)
	}

	return nil
}

const TEST_MEMBERSHIP_CREATOR_MEMBER_ABILITIES = "[\"CAN_QUERY\"]"
const TEST_MEMBERSHIP_MEMBER_ABILITIES = "[\"CAN_RECEIVE_RESULTS\"]"
const TEST_RESULT_FORMAT = "CSV"
const TEST_KEY_PREFIX = "test"

// We need to first initialize the IAM resources and S3 bucket, because there is a race condition
// between the permissions attached to the role and creation of the membership that uses it.
// If the membership is created before the permissions are actually attached, the membership creation will fail.
// Adding depends_on to the membership resource does not help, because the policy attachment doesn't come
// into effect immediately.
func testAccMembershipConfig_initResources(rName string) string {
	return acctest.ConfigCompose(
		testAccMembershipConfig_partitionDataSource(),
		testAccMembershipConfig_iamResources(rName),
		testAccMembershipConfig_s3Bucket(rName),
	)
}

func testAccMembershipConfig_initDoubledResources(rName string, rNameSecond string) string {
	return acctest.ConfigCompose(
		testAccMembershipConfig_partitionDataSource(),
		testAccMembershipConfig_iamResources(rName),
		testAccMembershipConfig_iamResources(rNameSecond),
		testAccMembershipConfig_s3Bucket(rName),
		testAccMembershipConfig_s3Bucket(rNameSecond),
	)
}

func testAccMembershipConfig_basic(rName string) string {
	defaultResultConfiguration := testAccMembershipConfig_defaultOutputConfiguration(rName, true, TEST_RESULT_FORMAT, TEST_KEY_PREFIX)
	return acctest.ConfigCompose(
		testAccMembershipConfig_partitionDataSource(),
		testAccMembershipConfig_iamResources(rName),
		testAccMembershipConfig_s3Bucket(rName),
		testAccMembershipConfig_base(rName, TEST_CREATOR_DISPLAY_NAME, TEST_MEMBERSHIP_CREATOR_MEMBER_ABILITIES, TEST_MEMBERSHIP_MEMBER_ABILITIES, TEST_QUERY_LOG_STATUS, defaultResultConfiguration, TEST_TAG),
	)
}

// Changing rNameToPointTo switches between referenced S3 bucket and IAM role in the default result configuration
func testAccMembershipConfig_mutableProperties(rName string, rNameSecond string, rNameToPointTo string, queryLogStatus string, resultFormat string, keyPrefix string, tagValue string) string {
	defaultResultConfiguration := testAccMembershipConfig_defaultOutputConfiguration(rNameToPointTo, true, resultFormat, keyPrefix)
	return acctest.ConfigCompose(
		testAccMembershipConfig_partitionDataSource(),
		testAccMembershipConfig_iamResources(rName),
		testAccMembershipConfig_iamResources(rNameSecond),
		testAccMembershipConfig_s3Bucket(rName),
		testAccMembershipConfig_s3Bucket(rNameSecond),
		testAccMembershipConfig_base(rName, TEST_CREATOR_DISPLAY_NAME, TEST_MEMBERSHIP_CREATOR_MEMBER_ABILITIES, TEST_MEMBERSHIP_MEMBER_ABILITIES, queryLogStatus, defaultResultConfiguration, tagValue),
	)
}

func testAccMembershipConfig_outputConfigurationWithEmptyAdditionalParameters(rName string) string {
	defaultResultConfiguration := testAccMembershipConfig_defaultOutputConfiguration(rName, false, TEST_RESULT_FORMAT, "")
	return acctest.ConfigCompose(
		testAccMembershipConfig_s3Bucket(rName),
		testAccMembershipConfig_base(rName, TEST_CREATOR_DISPLAY_NAME, "[]", "[\"CAN_QUERY\",\"CAN_RECEIVE_RESULTS\"]", TEST_QUERY_LOG_STATUS, defaultResultConfiguration, TEST_TAG),
	)
}

func testAccMembershipConfig_defaultOutputConfiguration(rName string, includeRoleArn bool, resultFormat string, keyPrefix string) string {
	roleArnEntry := ""
	if includeRoleArn {
		roleArnEntry = fmt.Sprintf("role_arn = aws_iam_role.test_%[1]s.arn", rName)
	}
	keyPrefixEntry := ""
	if keyPrefix != "" {
		keyPrefixEntry = fmt.Sprintf("key_prefix = %[1]q", keyPrefix)
	}

	return fmt.Sprintf(`
  default_result_configuration {
		
	%[1]s

    output_configuration {
	  s3 {
	    bucket        = aws_s3_bucket.test_%[2]s.bucket
        result_format = %[3]q
				
        %[4]s

      }
    }
  }`, roleArnEntry, rName, resultFormat, keyPrefixEntry)
}

func testAccMembershipConfig_partitionDataSource() string {
	return `
data "aws_partition" "current" {}`
}

func testAccMembershipConfig_iamResources(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test_assume_role_policy_%[1]s" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["cleanrooms.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "test_s3_policy_%[1]s" {
  statement {
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::%[1]s"]
  }

  statement {
    actions   = ["s3:PutObject"]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"]
  }
}

resource "aws_iam_policy" "test_%[1]s" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.test_s3_policy_%[1]s.json
}

resource "aws_iam_role" "test_%[1]s" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy_%[1]s.json
}

resource "aws_iam_role_policy_attachment" "test_%[1]s" {
  role       = aws_iam_role.test_%[1]s.id
  policy_arn = aws_iam_policy.test_%[1]s.arn
}`, rName)
}

func testAccMembershipConfig_s3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test_%[1]s" {
  bucket = %[1]q
}`, rName)
}

func testAccMembershipConfig_base(rName string, creatorDisplayName string, creatorMemberAbilities string,
	memberAbilities string, queryLogStatus string, defaultResultConfiguration string, tagValue string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_cleanrooms_collaboration" "test" {
  name                     = %[1]q
  description              = "test"
  creator_display_name     = %[2]q
  creator_member_abilities = %[3]s
  query_log_status         = "ENABLED"

  member {
    account_id       = data.aws_caller_identity.test.account_id
    display_name     = "Other member"
    member_abilities = %[4]s
  }

  provider = awsalternate
}

resource "aws_cleanrooms_membership" "test" {
  collaboration_id = aws_cleanrooms_collaboration.test.id
  query_log_status = %[5]q

  %[6]s

  tags = {
    Project = %[7]q
  }
}
	`, rName, creatorDisplayName, creatorMemberAbilities, memberAbilities, queryLogStatus, defaultResultConfiguration, tagValue)
}
