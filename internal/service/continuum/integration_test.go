// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/continuum"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcontinuum "github.com/hashicorp/terraform-provider-aws/internal/service/continuum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetIntegrationOutput
	resourceName := "aws_continuum_integration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check:  testAccCheckIntegrationExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("integration_type"), knownvalue.StringExact("S3")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3.0.external_id"},
			},
		},
	})
}

func testAccIntegration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetIntegrationOutput
	resourceName := "aws_continuum_integration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceIntegration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccIntegration_updateRole(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetIntegrationOutput
	resourceName := "aws_continuum_integration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_role(rName, "test"),
				Check:  testAccCheckIntegrationExists(ctx, t, resourceName, &v),
			},
			{
				Config: testAccIntegrationConfig_role(rName, "test2"),
				Check:  testAccCheckIntegrationExists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckIntegrationExists(ctx context.Context, t *testing.T, n string, v *continuum.GetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		output, err := tfcontinuum.FindIntegrationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIntegrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_integration" {
				continue
			}

			_, err := tfcontinuum.FindIntegrationByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Integration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccIntegrationConfig_base creates a bucket holding a placeholder code package and the roles
// that Continuum assumes to read it. The package is not a valid zip, so scans fail fast.
func testAccIntegrationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "code" {
  bucket     = aws_s3_bucket.test.id
  key        = "code/app.zip"
  content    = "placeholder"
  depends_on = [aws_s3_bucket_versioning.test]
}

data "aws_iam_policy_document" "trust" {
  statement {
    sid     = "ContinuumAccess"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["continuum.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [%[1]q]
    }
  }
}

data "aws_iam_policy_document" "access" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.test.arn]
  }

  statement {
    actions   = ["s3:GetObject", "s3:GetObjectVersion"]
    resources = ["${aws_s3_bucket.test.arn}/code/*", "${aws_s3_bucket.test.arn}/context/*"]
  }

  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:GetObjectVersion"]
    resources = ["${aws_s3_bucket.test.arn}/continuum-patches/*"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/continuum/integrations/"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  path               = "/continuum/integrations/"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test2.name
  policy = data.aws_iam_policy_document.access.json
}
`, rName)
}

func testAccIntegrationConfig_basic(rName string) string {
	return testAccIntegrationConfig_role(rName, "test")
}

func testAccIntegrationConfig_role(rName, role string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_base(rName), fmt.Sprintf(`
resource "aws_continuum_integration" "test" {
  name = %[1]q

  s3 {
    bucket      = aws_s3_bucket.test.id
    role_arn    = aws_iam_role.%[2]s.arn
    external_id = %[1]q
  }

  depends_on = [aws_iam_role_policy.test, aws_iam_role_policy.test2, aws_s3_object.code]
}
`, rName, role))
}
