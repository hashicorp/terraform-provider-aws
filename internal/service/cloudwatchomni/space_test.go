// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccCloudWatchOmniSpace_serial serializes every Space test.
//
// CloudWatch Omni permits exactly one space per account per region
// ("ConflictException: A space already exists for account <id> in region
// <region>"), so tests sharing a region cannot run concurrently. Running them
// in parallel makes whichever test loses the race fail on CreateSpace.
// envVarDomainID supplies an existing CloudWatch Omni domain to create test
// spaces under. The tests cannot create one: an account may hold only a single
// domain, and in an Organization with CloudWatch Omni trusted access
// account-level domain creation is refused and spaces must use the
// organization domain.
const envVarDomainID = "AWS_CLOUDWATCHOMNI_DOMAIN_ID"

func TestAccCloudWatchOmniSpace_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccCloudWatchOmniSpace_basic,
		acctest.CtDisappears: testAccCloudWatchOmniSpace_disappears,
		"agentCore":          testAccCloudWatchOmniSpace_agentCore,
		"encryption":         testAccCloudWatchOmniSpace_encryption,
		"update":             testAccCloudWatchOmniSpace_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccCloudWatchOmniSpace_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	domainID := acctest.SkipIfEnvVarNotSet(t, envVarDomainID)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_basic(rName, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttrSet(resourceName, "data_access_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "space_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "space_id"),
				ImportStateVerifyIdentifierAttribute: "space_id",
			},
		},
	})
}

func testAccCloudWatchOmniSpace_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	domainID := acctest.SkipIfEnvVarNotSet(t, envVarDomainID)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_basic(rName, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space, t),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceSpace, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCloudWatchOmniSpace_agentCore(t *testing.T) {
	ctx := acctest.Context(t)

	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	domainID := acctest.SkipIfEnvVarNotSet(t, envVarDomainID)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_agentCore(rName, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "agent_core_evaluation_role_arn"),
				),
			},
		},
	})
}

func testAccCloudWatchOmniSpace_encryption(t *testing.T) {
	ctx := acctest.Context(t)

	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	domainID := acctest.SkipIfEnvVarNotSet(t, envVarDomainID)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_encryption(rName, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_strategy", "CUSTOMER_MANAGED"),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_configuration.0.kms_key_arn"),
				),
			},
		},
	})
}

func testAccCloudWatchOmniSpace_update(t *testing.T) {
	ctx := acctest.Context(t)

	var space1, space2 awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	domainID := acctest.SkipIfEnvVarNotSet(t, envVarDomainID)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceConfig_basic(rName, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space1, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccSpaceConfig_basic(rNameUpdated, domainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, resourceName, &space2, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
				),
			},
		},
	})
}

func testAccCheckSpaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_space" {
				continue
			}

			_, err := tfcloudwatchomni.FindSpaceByID(ctx, conn, rs.Primary.Attributes["space_id"], acctest.Region())
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Space %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSpaceExists(ctx context.Context, n string, v *awstypes.Space, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindSpaceByID(ctx, conn, rs.Primary.Attributes["space_id"], acctest.Region())
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

	input := &cloudwatchomni.ListSpacesInput{}
	_, err := conn.ListSpaces(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSpaceConfig_basic(rName, domainID string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_cloudwatchomni_space" "test" {
  name                 = %[1]q
  domain_id            = %[2]q
  data_access_role_arn = aws_iam_role.test.arn
}
`, rName, domainID)
}

func testAccSpaceConfig_agentCore(rName, domainID string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test_data_access" {
  name = "%[1]s-data-access"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role" "test_agent_core" {
  name = "%[1]s-agent-core"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_cloudwatchomni_space" "test" {
  name                           = %[1]q
  domain_id                      = %[2]q
  data_access_role_arn           = aws_iam_role.test_data_access.arn
  agent_core_evaluation_role_arn = aws_iam_role.test_agent_core.arn
}
`, rName, domainID)
}

func testAccSpaceConfig_encryption(rName, domainID string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_kms_key" "test" {
  description = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow CloudWatch Omni Service"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey",
          "kms:Encrypt",
          "kms:GenerateDataKey*",
          "kms:ReEncrypt*"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_cloudwatchomni_space" "test" {
  name                 = %[1]q
  domain_id            = %[2]q
  data_access_role_arn = aws_iam_role.test.arn

  encryption_configuration {
    encryption_strategy = "CUSTOMER_MANAGED"
    kms_key_arn         = aws_kms_key.test.arn
  }
}
`, rName, domainID)
}
