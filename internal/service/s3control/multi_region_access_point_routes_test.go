// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlMultiRegionAccessPointRoutes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_s3control_multi_region_access_point_routes.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointRoutesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionAccessPointRoutesExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mrap"), tfknownvalue.GlobalARNRegexp("s3", regexache.MustCompile(`accesspoint/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-1"),
							names.AttrRegion:          knownvalue.StringExact(acctest.Region()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-2"),
							names.AttrRegion:          knownvalue.StringExact(acctest.AlternateRegion()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "mrap"),
				ImportStateVerifyIdentifierAttribute: "mrap",
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPointRoutes_Disappears_multiRegionAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	mrapResourceName := "aws_s3control_multi_region_access_point.test"
	resourceName := "aws_s3control_multi_region_access_point_routes.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointRoutesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointRoutesExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3control.ResourceMultiRegionAccessPoint(), mrapResourceName),
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

func TestAccS3ControlMultiRegionAccessPointRoutes_route(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_s3control_multi_region_access_point_routes.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointRoutesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionAccessPointRoutesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-1"),
							names.AttrRegion:          knownvalue.StringExact(acctest.Region()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-2"),
							names.AttrRegion:          knownvalue.StringExact(acctest.AlternateRegion()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
					})),
				},
			},
			{
				Config: testAccMultiRegionAccessPointRoutesConfig_failover(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionAccessPointRoutesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-1"),
							names.AttrRegion:          knownvalue.StringExact(acctest.Region()),
							"traffic_dial_percentage": knownvalue.Int64Exact(0),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-2"),
							names.AttrRegion:          knownvalue.StringExact(acctest.AlternateRegion()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
					})),
				},
			},
			{
				Config: testAccMultiRegionAccessPointRoutesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionAccessPointRoutesExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-1"),
							names.AttrRegion:          knownvalue.StringExact(acctest.Region()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucket:          knownvalue.StringExact(rName + "-2"),
							names.AttrRegion:          knownvalue.StringExact(acctest.AlternateRegion()),
							"traffic_dial_percentage": knownvalue.Int64Exact(100),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckMultiRegionAccessPointRoutesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		accountID := rs.Primary.Attributes[names.AttrAccountID]
		mrap := rs.Primary.Attributes["mrap"]

		_, err := tfs3control.FindMultiRegionAccessPointRoutesByTwoPartKey(ctx, conn, accountID, mrap)

		return err
	}
}

func testAccMultiRegionAccessPointRoutesConfig_base(rName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  provider = aws

  bucket        = "%[1]s-1"
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  region = %[2]q

  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[1]q

    region {
      bucket = aws_s3_bucket.test1.bucket
    }

    region {
      bucket = aws_s3_bucket.test2.bucket
    }
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccMultiRegionAccessPointRoutesConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMultiRegionAccessPointRoutesConfig_base(rName), `
resource "aws_s3control_multi_region_access_point_routes" "test" {
  mrap = aws_s3control_multi_region_access_point.test.arn

  route {
    bucket                  = aws_s3_bucket.test1.bucket
    region                  = aws_s3_bucket.test1.bucket_region
    traffic_dial_percentage = 100
  }

  route {
    bucket                  = aws_s3_bucket.test2.bucket
    region                  = aws_s3_bucket.test2.bucket_region
    traffic_dial_percentage = 100
  }
}
`)
}

func testAccMultiRegionAccessPointRoutesConfig_failover(rName string) string {
	return acctest.ConfigCompose(testAccMultiRegionAccessPointRoutesConfig_base(rName), `
resource "aws_s3control_multi_region_access_point_routes" "test" {
  mrap = aws_s3control_multi_region_access_point.test.arn

  route {
    bucket                  = aws_s3_bucket.test1.bucket
    region                  = aws_s3_bucket.test1.bucket_region
    traffic_dial_percentage = 0
  }

  route {
    bucket                  = aws_s3_bucket.test2.bucket
    region                  = aws_s3_bucket.test2.bucket_region
    traffic_dial_percentage = 100
  }
}
`)
}
