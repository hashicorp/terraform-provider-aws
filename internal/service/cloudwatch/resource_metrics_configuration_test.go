// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The resource metrics configuration API rejects resource types that are not
// eligible for detailed metrics with this validation error.
const (
	errCodeValidationError            = "ValidationError"
	errMessageUnsupportedResourceType = "Unsupported resource type"
)

func TestAccCloudWatchResourceMetricsConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_resource_metrics_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckResourceMetricsConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceMetricsConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMetricsConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selections"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAccCloudWatchResourceMetricsConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_resource_metrics_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckResourceMetricsConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceMetricsConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMetricsConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceResourceMetricsConfiguration, resourceName),
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

func TestAccCloudWatchResourceMetricsConfiguration_metricSelections(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_resource_metrics_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckResourceMetricsConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceMetricsConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMetricsConfigurationConfig_metricSelections(rName, `"CacheHits"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selections").AtSliceIndex(0).AtMapKey("include_metrics"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("CacheHits"),
					})),
				},
			},
			// Add a metric to the selection.
			{
				Config: testAccResourceMetricsConfigurationConfig_metricSelections(rName, `"CacheHits", "CacheMisses"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selections").AtSliceIndex(0).AtMapKey("include_metrics"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("CacheHits"),
						knownvalue.StringExact("CacheMisses"),
					})),
				},
			},
			// Removing the block clears the filter so that all available detailed
			// metrics are collected again.
			{
				Config: testAccResourceMetricsConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selections"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccCheckResourceMetricsConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_resource_metrics_configuration" {
				continue
			}

			_, err := tfcloudwatch.FindResourceMetricsConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Resource Metrics Configuration %s still exists", rs.Primary.Attributes[names.AttrResourceARN])
		}

		return nil
	}
}

func testAccCheckResourceMetricsConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		_, err := tfcloudwatch.FindResourceMetricsConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])

		return err
	}
}

// testAccPreCheckResourceMetricsConfiguration verifies that resource metrics
// configurations can be managed in the current account, Region and partition.
//
// There is no list operation for this API, so the check reads the configuration
// of a resource that does not exist. A ResourceNotFoundException means the API
// is reachable and the resource type is eligible for detailed metrics, while a
// validation error means the resource type is not eligible here.
func testAccPreCheckResourceMetricsConfiguration(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

	input := cloudwatch.GetResourceMetricsConfigurationInput{
		ResourceArn: aws.String(arn.ARN{
			Partition: acctest.Partition(),
			Service:   "elasticache",
			Region:    acctest.Region(),
			AccountID: acctest.AccountID(ctx),
			Resource:  "replicationgroup:tf-acc-test-does-not-exist",
		}.String()),
	}
	_, err := conn.GetResourceMetricsConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if tfawserr.ErrMessageContains(err, errCodeValidationError, errMessageUnsupportedResourceType) {
		t.Skipf("skipping acceptance testing: resource metrics configurations are not supported for this resource type: %s", err)
	}

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccResourceMetricsConfigurationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "terraform-provider-aws acceptance testing"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  apply_immediately    = true
}
`, rName)
}

func testAccResourceMetricsConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceMetricsConfigurationConfig_base(rName), `
resource "aws_cloudwatch_resource_metrics_configuration" "test" {
  resource_arn = aws_elasticache_replication_group.test.arn
}
`)
}

func testAccResourceMetricsConfigurationConfig_metricSelections(rName, includeMetrics string) string {
	return acctest.ConfigCompose(testAccResourceMetricsConfigurationConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_resource_metrics_configuration" "test" {
  resource_arn = aws_elasticache_replication_group.test.arn

  metric_selections {
    include_metrics = [%[1]s]
  }
}
`, includeMetrics))
}
