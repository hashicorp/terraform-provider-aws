// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPAnomalyDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"
	workspaceResourceName := "aws_prometheus_workspace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "evaluation_interval_in_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "avg(up)"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.random_cut_forest.0.sample_size"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.random_cut_forest.0.shingle_size"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_above.0"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_below.0"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "missing_data_action.0.skip", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "missing_data_action.0.mark_as_anomaly"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aps", regexache.MustCompile(`anomalydetector/.+$`)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrID, "workspace_id"),
			},
		},
	})
}

func TestAccAMPAnomalyDetector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfamp.ResourceAnomalyDetector, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccAMPAnomalyDetector_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_update(rName, "60", "avg(up)", "mark_as_anomaly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "avg(up)"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "missing_data_action.0.mark_as_anomaly", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "missing_data_action.0.skip"),
				),
			},
			{ // Testing evaluation_interval_in_seconds update
				Config: testAccAnomalyDetectorConfig_update(rName, "120", "avg(up)", "mark_as_anomaly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "evaluation_interval_in_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "avg(up)"),
					resource.TestCheckResourceAttr(resourceName, "missing_data_action.0.mark_as_anomaly", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "missing_data_action.0.skip"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{ // Testing query update
				Config: testAccAnomalyDetectorConfig_update(rName, "120", "count(up)", "mark_as_anomaly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "evaluation_interval_in_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "count(up)"),
					resource.TestCheckResourceAttr(resourceName, "missing_data_action.0.mark_as_anomaly", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "missing_data_action.0.skip"),
				),
			},
			{ // Testing missing_data_action update
				Config: testAccAnomalyDetectorConfig_update(rName, "120", "count(up)", "skip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, "evaluation_interval_in_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "count(up)"),
					resource.TestCheckResourceAttr(resourceName, "missing_data_action.0.skip", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "missing_data_action.0.mark_as_anomaly"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrID, "workspace_id"),
			},
		},
	})
}

func TestAccAMPAnomalyDetector_randomCutForest(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_randomCutForest(rName, "256", "4", "ratio", "amount"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "avg(up)"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.sample_size", "256"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.shingle_size", "4"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_above.0.ratio", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_below.0.amount", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_above.0.amount"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_below.0.ratio"),
				),
			},
			{ // Update check
				Config: testAccAnomalyDetectorConfig_randomCutForest(rName, "512", "16", "amount", "ratio"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.query", "avg(up)"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.sample_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.shingle_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_above.0.amount", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_below.0.ratio", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_above.0.ratio"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.random_cut_forest.0.ignore_near_expected_from_below.0.amount"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrID, "workspace_id"),
			},
		},
	})
}

func TestAccAMPAnomalyDetector_labels(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_labels(rName, "labelKey1", "labelValue1", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.labelKey1", "labelValue1"),
				),
			},
			{ // Update and add check
				Config: testAccAnomalyDetectorConfig_labels(rName, "starship", "Enterprise", "\"captain\" = \"Picard\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.starship", "Enterprise"),
					resource.TestCheckResourceAttr(resourceName, "labels.captain", "Picard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrID, "workspace_id"),
			},
		},
	})
}

func testAccAnomalyDetectorImportState(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrID, "workspace_id")
}

func testAccCheckAnomalyDetectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_anomaly_detector" {
				continue
			}

			_, err := tfamp.FindAnomalyDetectorByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])
			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.AMP, create.ErrActionCheckingDestroyed, tfamp.ResNameAnomalyDetector, rs.Primary.ID, err)
			}

			return create.Error(names.AMP, create.ErrActionCheckingDestroyed, tfamp.ResNameAnomalyDetector, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAnomalyDetectorExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		_, err := tfamp.FindAnomalyDetectorByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["workspace_id"])
		if err != nil {
			return create.Error(names.AMP, create.ErrActionCheckingExistence, tfamp.ResNameAnomalyDetector, rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccAnomalyDetectorConfig_base = `resource "aws_prometheus_workspace" "test" {}`

func testAccAnomalyDetectorConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAnomalyDetectorConfig_base,
		fmt.Sprintf(`
resource "aws_prometheus_anomaly_detector" "test" {
  alias                          = %[1]q
  workspace_id                   = aws_prometheus_workspace.test.id
  evaluation_interval_in_seconds = 120

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  missing_data_action {
    skip = true
  }
}
`, rName))
}

func testAccAnomalyDetectorConfig_update(rName, eval_time, query, missingDataAction string) string {
	return acctest.ConfigCompose(
		testAccAnomalyDetectorConfig_base,
		fmt.Sprintf(`
resource "aws_prometheus_anomaly_detector" "test" {
  alias                          = %[1]q
  workspace_id                   = aws_prometheus_workspace.test.id
  evaluation_interval_in_seconds = %[2]s

  configuration {
    random_cut_forest {
      query = %[3]q
    }
  }

  missing_data_action {
    %[4]s = true
  }
}
`, rName, eval_time, query, missingDataAction))
}

func testAccAnomalyDetectorConfig_randomCutForest(rName, sampleSize, shingleSize, ignoreAbove, ignoreBelow string) string {
	return acctest.ConfigCompose(
		testAccAnomalyDetectorConfig_base,
		fmt.Sprintf(`
resource "aws_prometheus_anomaly_detector" "test" {
  alias        = %[1]q
  workspace_id = aws_prometheus_workspace.test.id

  configuration {
    random_cut_forest {
      query        = "avg(up)"
      sample_size  = %[2]s
      shingle_size = %[3]s

      ignore_near_expected_from_above {
        %[4]s = 2
      }

      ignore_near_expected_from_below {
        %[5]s = 2
      }
    }
  }

  missing_data_action {
    skip = true
  }
}
`, rName, sampleSize, shingleSize, ignoreAbove, ignoreBelow))
}

func testAccAnomalyDetectorConfig_labels(rName, labelKey1, labelValue1, label2 string) string {
	return acctest.ConfigCompose(
		testAccAnomalyDetectorConfig_base,
		fmt.Sprintf(`
resource "aws_prometheus_anomaly_detector" "test" {
  alias                          = %[1]q
  workspace_id                   = aws_prometheus_workspace.test.id
  evaluation_interval_in_seconds = 120

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  labels = {
    %[2]q = %[3]q
	%[4]s
  }

  missing_data_action {
    skip = true
  }
}
`, rName, labelKey1, labelValue1, label2))
}
