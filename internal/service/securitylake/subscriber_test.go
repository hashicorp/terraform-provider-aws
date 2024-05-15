// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSubscriber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					func(s *terraform.State) error {
						id := aws.ToString(subscriber.SubscriberId)
						return acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "securitylake", fmt.Sprintf("subscriber/%s", id))(s)
					},
					resource.TestCheckResourceAttr(resourceName, "resource_share_arn", ""),
					func(s *terraform.State) error {
						id := aws.ToString(subscriber.SubscriberId)
						return acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", fmt.Sprintf("role/AmazonSecurityLake-%s", id))(s)
					},
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, "s3_bucket_arn", "s3", regexache.MustCompile(fmt.Sprintf(`aws-security-data-lake-%s-[a-z0-9]{30}$`, acctest.Region()))),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					func(s *terraform.State) error {
						id := aws.ToString(subscriber.SubscriberId)
						return resource.TestCheckResourceAttr(resourceName, names.AttrID, id)(s)
					},
					resource.TestCheckNoResourceAttr(resourceName, "subscriber_description"),
					resource.TestCheckNoResourceAttr(resourceName, "resource_share_name"),
					resource.TestCheckNoResourceAttr(resourceName, "subscriber_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "2.0"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var subscriber types.SubscriberResource
	resourceName := "aws_securitylake_subscriber.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriber, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSubscriber_customLogSource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceName := randomCustomLogSourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_customLog(rName, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_log_source_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.custom_log_source_resource.0.source_name", "aws_securitylake_custom_log_source.test", "source_name"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.custom_log_source_resource.0.source_version", "aws_securitylake_custom_log_source.test", "source_version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriber_accessType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_accessType(rName, "S3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriber_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriberConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSubscriberConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccSubscriber_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriberConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "VPC_FLOW"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriber_multipleSources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_sources2(rName, "VPC_FLOW", "ROUTE53"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New(names.AttrSource),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"source_name": knownvalue.StringExact("VPC_FLOW"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"source_name": knownvalue.StringExact("ROUTE53"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
						}),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriberConfig_sources2(rName, "ROUTE53", "VPC_FLOW"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccSubscriberConfig_sources1(rName, "ROUTE53"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New(names.AttrSource),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"source_name": knownvalue.StringExact("ROUTE53"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
						}),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubscriberConfig_sources2(rName, "VPC_FLOW", "S3_DATA"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New(names.AttrSource),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"source_name": knownvalue.StringExact("VPC_FLOW"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"source_name": knownvalue.StringExact("S3_DATA"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
						}),
					),
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

func testAccSubscriber_migrate_source(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		CheckDestroy: testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.47.0",
					},
				},
				Config: testAccSubscriberConfig_migrate_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New(names.AttrSource),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"aws_log_source_resource": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"source_name":    knownvalue.StringExact("ROUTE53"),
										"source_version": knownvalue.StringExact("2.0"),
									}),
								}),
								"custom_log_source_resource": knownvalue.ListSizeExact(0),
							}),
						}),
					),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccSubscriberConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckSubscriberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_subscriber" {
				continue
			}

			_, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Subscriber %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubscriberExists(ctx context.Context, n string, v *types.SubscriberResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriberConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name = "ROUTE53"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName))
}

func testAccSubscriberConfig_customLog(rName, sourceName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name        = %[1]q
  subscriber_description = "Example"
  source {
    custom_log_source_resource {
      source_name    = aws_securitylake_custom_log_source.test.source_name
      source_version = aws_securitylake_custom_log_source.test.source_version
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_custom_log_source.test]
}

resource "aws_securitylake_custom_log_source" "test" {
  source_name    = %[2]q
  source_version = "1.5"

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[2]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[2]q
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
"Version": "2012-10-17",
"Statement": [{
	"Action": "sts:AssumeRole",
	"Principal": {
	"Service": "glue.amazonaws.com"
	},
	"Effect": "Allow"
}]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[2]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
	"Version": "2012-10-17",
		"Statement": [{
		"Effect": "Allow",
		"Action": [
		"s3:GetObject",
		"s3:PutObject"
		],
		"Resource": "*"
}]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}
`, rName, sourceName))
}

func testAccSubscriberConfig_accessType(rName, accessType string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  access_type     = %[2]q
  source {
    aws_log_source_resource {
      source_name = "ROUTE53"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName, accessType))
}

func testAccSubscriberConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name    = "ROUTE53"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.name]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName, tag1Key, tag1Value))
}

func testAccSubscriberConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name    = "ROUTE53"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.name]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccSubscriberConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  access_type     = "S3"
  source {
    aws_log_source_resource {
      source_name    = "VPC_FLOW"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "updated"
    principal   = data.aws_caller_identity.current.account_id
  }
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.name]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName))
}

func testAccSubscriberConfig_sources1(rName, source1 string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name = %[2]q
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = %[2]q
  }
  depends_on = [
    aws_securitylake_data_lake.test,
  ]
}

data "aws_region" "current" {}
`, rName, source1))
}

func testAccSubscriberConfig_sources2(rName, source1, source2 string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name = %[2]q
    }
  }
  source {
    aws_log_source_resource {
      source_name = %[3]q
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [
    aws_securitylake_aws_log_source.test,
  ]
}

resource "aws_securitylake_aws_log_source" "test" {
  count = length(local.source_names)

  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = local.source_names[count.index]
  }
  depends_on = [aws_securitylake_data_lake.test]
}

locals {
  source_names = sort([%[2]q, %[3]q])
}

data "aws_region" "current" {}
`, rName, source1, source2))
}

func testAccSubscriberConfig_migrate_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q
  source {
    aws_log_source_resource {
      source_name    = "ROUTE53"
      source_version = "2.0"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.name]
    source_name    = "ROUTE53"
    source_version = "2.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, rName))
}
