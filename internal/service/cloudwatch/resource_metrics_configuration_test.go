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
	"github.com/hashicorp/terraform-plugin-testing/config"
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
				ConfigDirectory: config.StaticDirectory("testdata/ResourceMetricsConfiguration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selection"), knownvalue.ListSizeExact(0)),
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
				ConfigDirectory: config.StaticDirectory("testdata/ResourceMetricsConfiguration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceResourceMetricsConfiguration, resourceName),
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
				ConfigDirectory: config.StaticDirectory("testdata/ResourceMetricsConfiguration/metric_selection/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"include_metrics": acctest.SetOfStringsVariable("valkey.memory.used"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selection").AtSliceIndex(0).AtMapKey("include_metrics"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("valkey.memory.used"),
					})),
				},
			},
			// Add a metric to the selection.
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourceMetricsConfiguration/metric_selection/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"include_metrics": acctest.SetOfStringsVariable("valkey.memory.used", "valkey.commands.processed"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selection").AtSliceIndex(0).AtMapKey("include_metrics"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("valkey.memory.used"),
						knownvalue.StringExact("valkey.commands.processed"),
					})),
				},
			},
			// Removing the block clears the filter so that all available detailed
			// metrics are collected again.
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourceMetricsConfiguration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceMetricsConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metric_selection"), knownvalue.ListSizeExact(0)),
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
