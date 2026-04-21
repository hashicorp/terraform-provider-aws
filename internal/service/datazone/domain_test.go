// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "portal_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "datazone", "domain/{id}"),
					resource.TestCheckResourceAttr(resourceName, "domain_version", "V1"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrServiceRole),
					resource.TestCheckResourceAttrSet(resourceName, "root_domain_unit_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user", "skip_deletion_check"},
			},
			{
				Config: testAccDomainConfig_description(rName, "test_description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "portal_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "datazone", "domain/{id}"),
					resource.TestCheckResourceAttr(resourceName, "domain_version", "V1"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrServiceRole),
					resource.TestCheckResourceAttrSet(resourceName, "root_domain_unit_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test_description"),
				),
			},
		},
	})
}

func TestAccDataZoneDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneDomain_kms_key_identifier(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kms_key_identifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_identifier"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user", "skip_deletion_check"},
			},
		},
	})
}

func TestAccDataZoneDomain_description(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"
	description := "This is a description"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user", "skip_deletion_check"},
			},
		},
	})
}

func TestAccDataZoneDomain_single_sign_on(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_single_sign_on(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// we do not set single_sign_on if it's the default value
				ImportStateVerifyIgnore: []string{"single_sign_on", "skip_deletion_check"},
			},
		},
	})
}

func TestAccDataZoneDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDataZoneDomain_domainVersionV2(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.GetDomainOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_domainVersionV2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "portal_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "datazone", "domain/{id}"),
					resource.TestCheckResourceAttr(resourceName, "domain_version", "V2"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_execution_role", "aws_iam_role.domain_execution", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.domain_service", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user", "skip_deletion_check"},
			},
		},
	})
}

func testAccCheckDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_domain" {
				continue
			}

			_, err := tfdatazone.FindDomainByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataZone Domain (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, t *testing.T, n string, v *datazone.GetDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		output, err := tfdatazone.FindDomainByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

	input := &datazone.ListDomainsInput{}
	_, err := conn.ListDomains(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDomainConfigDomainExecutionRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "domain_execution_role" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = %[1]q
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}
`, rName)
}

func testAccDomainConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}
`, rName),
	)
}

func testAccDomainConfig_kms_key_identifier(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  kms_key_identifier    = aws_kms_key.test.arn
}
`, rName),
	)
}

func testAccDomainConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  description           = %[2]q
}
`, rName, description),
	)
}

func testAccDomainConfig_single_sign_on(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  single_sign_on {
    type = "DISABLED"
  }
}
`, rName),
	)
}

func testAccDomainConfig_tags(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue),
	)
}

func testAccDomainConfig_tags2(rName, tagKey, tagValue, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey, tagValue, tagKey2, tagValue2),
	)
}

func testAccDomainConfig_domainVersionV2(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

# IAM role for Domain Execution
data "aws_iam_policy_document" "assume_role_domain_execution" {
  statement {
    actions = [
      "sts:AssumeRole",
      "sts:TagSession",
      "sts:SetContext"
    ]
    principals {
      type        = "Service"
      identifiers = ["datazone.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
    condition {
      test     = "ForAllValues:StringLike"
      values   = ["datazone*"]
      variable = "aws:TagKeys"
    }
  }
}

resource "aws_iam_role" "domain_execution" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_execution.json
  name               = "%[1]s-domain-execution-role"
}

data "aws_iam_policy" "domain_execution_role" {
  name = "SageMakerStudioDomainExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_execution" {
  policy_arn = data.aws_iam_policy.domain_execution_role.arn
  role       = aws_iam_role.domain_execution.name
}

# IAM role for Domain Service
data "aws_iam_policy_document" "assume_role_domain_service" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["datazone.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "domain_service" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_service.json
  name               = "%[1]s-domain-service-role"
}

data "aws_iam_policy" "domain_service_role" {
  name = "SageMakerStudioDomainServiceRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_service" {
  policy_arn = data.aws_iam_policy.domain_service_role.arn
  role       = aws_iam_role.domain_service.name
}

# DataZone Domain V2
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution.arn
  domain_version        = "V2"
  service_role          = aws_iam_role.domain_service.arn
}
`, rName)
}
