// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentBlueprint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out datazone.GetEnvironmentBlueprintOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_properties.0.cloud_formation.0.template_url"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrID,
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprint_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out, outUpdated datazone.GetEnvironmentBlueprintOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfig_description(rName, "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial"),
				),
			},
			{
				Config: testAccEnvironmentBlueprintConfig_description(rName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, t, resourceName, &outUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
					// The update must mutate in place, not replace the blueprint.
					testAccCheckEnvironmentBlueprintNotRecreated(&out, &outUpdated),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrID,
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out datazone.GetEnvironmentBlueprintOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, t, resourceName, &out),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceEnvironmentBlueprint, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprint_userParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out datazone.GetEnvironmentBlueprintOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfig_userParameters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentBlueprintExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.key_name", "costCenter"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.field_type", "String"),
					// The Optional+Computed bools must settle to the configured values.
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.is_editable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.is_optional", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.is_update_supported", acctest.CtTrue),
				),
			},
			{
				// The user_parameters block and its Optional+Computed bools must
				// round-trip without a perpetual diff.
				Config: testAccEnvironmentBlueprintConfig_userParameters(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrID,
			},
		},
	})
}

func testAccCheckEnvironmentBlueprintDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment_blueprint" {
				continue
			}

			_, err := tfdatazone.FindEnvironmentBlueprintByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrID])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DataZone Environment Blueprint %s still exists", rs.Primary.Attributes[names.AttrID])
		}

		return nil
	}
}

func testAccCheckEnvironmentBlueprintExists(ctx context.Context, t *testing.T, n string, v *datazone.GetEnvironmentBlueprintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		out, err := tfdatazone.FindEnvironmentBlueprintByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrID])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckEnvironmentBlueprintNotRecreated(before, after *datazone.GetEnvironmentBlueprintOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return fmt.Errorf("DataZone Environment Blueprint was recreated: before %s, after %s", before, after)
		}

		return nil
	}
}

func testAccEnvironmentBlueprintImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrID]), nil
	}
}

// testAccEnvironmentBlueprintConfig_base provisions the dependencies a custom
// CloudFormation blueprint requires: a V2 (SageMaker Unified Studio) domain with
// its execution and service roles, and an in-account S3 bucket holding the
// CloudFormation template. Custom blueprints are a V2-domain feature, and
// DataZone reads the template as the domain execution role, which must therefore
// carry SageMakerStudioDomainExecutionRolePolicy and the template must live in a
// bucket in the domain account.
func testAccEnvironmentBlueprintConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_version        = "V2"
  domain_execution_role = aws_iam_role.domain_execution.arn
  service_role          = aws_iam_role.domain_service.arn
  skip_deletion_check   = true
}

data "aws_iam_policy_document" "assume_role_domain_execution" {
  statement {
    actions = ["sts:AssumeRole", "sts:TagSession", "sts:SetContext"]

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
  name               = "%[1]s-exec"
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_execution.json
}

data "aws_iam_policy" "domain_execution_role" {
  name = "SageMakerStudioDomainExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_execution" {
  policy_arn = data.aws_iam_policy.domain_execution_role.arn
  role       = aws_iam_role.domain_execution.name
}

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
  name               = "%[1]s-svc"
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_service.json
}

data "aws_iam_policy" "domain_service_role" {
  name = "SageMakerStudioDomainServiceRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_service" {
  policy_arn = data.aws_iam_policy.domain_service_role.arn
  role       = aws_iam_role.domain_service.name
}

resource "aws_s3_bucket" "template" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "template" {
  bucket  = aws_s3_bucket.template.id
  key     = "template.yaml"
  content = <<-YAML
    Resources:
      Index:
        Type: AWS::OpenSearchServerless::Index
        Properties:
          IndexName: textIndex
          CollectionEndpoint: colTest
  YAML
}
`, rName)
}

func testAccEnvironmentBlueprintConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentBlueprintConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment_blueprint" "test" {
  domain_identifier = aws_datazone_domain.test.id
  name              = %[1]q

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_object.template.bucket}.s3.amazonaws.com/${aws_s3_object.template.key}"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.domain_execution]
}
`, rName))
}

func testAccEnvironmentBlueprintConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccEnvironmentBlueprintConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment_blueprint" "test" {
  domain_identifier = aws_datazone_domain.test.id
  name              = %[1]q
  description       = %[2]q

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_object.template.bucket}.s3.amazonaws.com/${aws_s3_object.template.key}"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.domain_execution]
}
`, rName, description))
}

func testAccEnvironmentBlueprintConfig_userParameters(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentBlueprintConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment_blueprint" "test" {
  domain_identifier = aws_datazone_domain.test.id
  name              = %[1]q

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_object.template.bucket}.s3.amazonaws.com/${aws_s3_object.template.key}"
    }
  }

  user_parameters {
    key_name            = "costCenter"
    field_type          = "String"
    default_value       = "engineering"
    is_editable         = true
    is_optional         = true
    is_update_supported = true
  }

  depends_on = [aws_iam_role_policy_attachment.domain_execution]
}
`, rName))
}
