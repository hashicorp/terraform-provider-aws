// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/continuum"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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

// Continuum allows one collector per account and Region, and creating an environment starts a scan,
// so every acceptance test runs serially.
func TestAccContinuum_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Integration": {
			acctest.CtBasic:      testAccIntegration_basic,
			acctest.CtDisappears: testAccIntegration_disappears,
			"updateRole":         testAccIntegration_updateRole,
		},
		"Collector": {
			acctest.CtBasic:      testAccCollector_basic,
			acctest.CtDisappears: testAccCollector_disappears,
		},
		"Environment": {
			acctest.CtBasic:      testAccEnvironment_basic,
			acctest.CtDisappears: testAccEnvironment_disappears,
			"update":             testAccEnvironment_update,
		},
		"CollectorAssociation": {
			acctest.CtBasic:      testAccCollectorAssociation_basic,
			acctest.CtDisappears: testAccCollectorAssociation_disappears,
		},
		"ResourcePolicy": {
			acctest.CtBasic: testAccResourcePolicy_basic,
		},
		"DataSource": {
			acctest.CtBasic: testAccDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

	input := continuum.ListIntegrationsInput{}
	_, err := conn.ListIntegrations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// Integration

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

// Collector

func testAccCollector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetCollectorOutput
	resourceName := "aws_continuum_collector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorConfig_basic(),
				Check:  testAccCheckCollectorExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("AWS_COLLECTOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCollector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetCollectorOutput
	resourceName := "aws_continuum_collector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectorExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceCollector, resourceName),
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

func testAccCheckCollectorExists(ctx context.Context, t *testing.T, n string, v *continuum.GetCollectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		output, err := tfcontinuum.FindCollectorByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCollectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_collector" {
				continue
			}

			_, err := tfcontinuum.FindCollectorByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Collector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCollectorConfig_basic() string {
	return `
resource "aws_continuum_collector" "test" {}
`
}

// Environment

func testAccEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scan_frequency"), knownvalue.StringExact("PT0H")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceEnvironment, resourceName),
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

func testAccEnvironment_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := rName + "-updated"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
			},
			{
				Config: testAccEnvironmentConfig_basic(rName, rNameUpdated, "P7D"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scan_frequency"), knownvalue.StringExact("P7D")),
				},
			},
		},
	})
}

func testAccCheckEnvironmentExists(ctx context.Context, t *testing.T, n string, v *continuum.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		output, err := tfcontinuum.FindEnvironmentByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEnvironmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_environment" {
				continue
			}

			_, err := tfcontinuum.FindEnvironmentByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEnvironmentConfig_basic(rName, envName, scanFrequency string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_basic(rName), fmt.Sprintf(`
resource "aws_continuum_environment" "test" {
  name           = %[1]q
  scan_frequency = %[2]q

  s3_source {
    integration_id = aws_continuum_integration.test.id
    code_prefix    = "code/"
  }
}
`, envName, scanFrequency))
}

// Collector association

func testAccCollectorAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_continuum_collector_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorAssociationConfig_basic(rName),
				Check:  testAccCheckCollectorAssociationExists(ctx, t, resourceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("collector_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCollectorAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_continuum_collector_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectorAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceCollectorAssociation, resourceName),
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

func testAccCheckCollectorAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		_, err := tfcontinuum.FindAssociatedCollector(ctx, conn, rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["collector_arn"])

		return err
	}
}

func testAccCheckCollectorAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_collector_association" {
				continue
			}

			_, err := tfcontinuum.FindAssociatedCollector(ctx, conn, rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["collector_arn"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Collector Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCollectorAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_basic(rName, rName, "PT0H"), fmt.Sprintf(`
resource "aws_continuum_collector" "test" {}

data "aws_iam_policy_document" "collector_trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["continuum.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = [aws_continuum_environment.test.arn]
    }
  }
}

data "aws_iam_policy_document" "collector_access" {
  statement {
    actions   = ["continuum:GetCollectorDataAccessAuthorization"]
    resources = [aws_continuum_collector.test.arn]
  }
}

resource "aws_iam_role" "collector" {
  name               = "%[1]s-collector"
  path               = "/continuum/"
  assume_role_policy = data.aws_iam_policy_document.collector_trust.json
}

resource "aws_iam_role_policy" "collector" {
  role   = aws_iam_role.collector.name
  policy = data.aws_iam_policy_document.collector_access.json
}

resource "aws_continuum_collector_association" "test" {
  environment_id                 = aws_continuum_environment.test.id
  collector_arn                  = aws_continuum_collector.test.arn
  collector_data_access_role_arn = aws_iam_role.collector.arn

  depends_on = [aws_iam_role_policy.collector]
}
`, rName))
}

// Resource policy

func testAccResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_continuum_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("revision_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_resource_policy" {
				continue
			}

			_, err := tfcontinuum.FindResourcePolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCollectorConfig_basic(), `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_continuum_resource_policy" "test" {
  resource_arn = aws_continuum_collector.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root" }
      Action    = "continuum:GetCollectorDataAccessAuthorization"
      Resource  = aws_continuum_collector.test.arn
    }]
  })
}
`)
}

// Data sources

func testAccDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs("data.aws_continuum_integration.by_name", tfjsonpath.New(names.AttrID), "aws_continuum_integration.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs("data.aws_continuum_environment.by_id", tfjsonpath.New(names.AttrName), "aws_continuum_environment.test", tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_scans.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_issues.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.aws_continuum_code_findings.test", tfjsonpath.New("environment_id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_basic(rName, rName, "PT0H"), `
data "aws_continuum_integration" "by_name" {
  name = aws_continuum_integration.test.name
}

data "aws_continuum_environment" "by_id" {
  id = aws_continuum_environment.test.id
}

data "aws_continuum_code_scans" "test" {
  environment_id = aws_continuum_environment.test.id
}

data "aws_continuum_code_issues" "test" {
  environment_id = aws_continuum_environment.test.id
}

data "aws_continuum_code_findings" "test" {
  environment_id = aws_continuum_environment.test.id
}
`)
}
