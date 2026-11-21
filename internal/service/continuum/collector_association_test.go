// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"context"
	"fmt"
	"testing"

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
