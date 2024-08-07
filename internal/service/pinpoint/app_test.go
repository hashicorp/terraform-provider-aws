// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("campaign_hook"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"lambda_function_name": knownvalue.StringExact(""),
							names.AttrMode:         knownvalue.StringExact(""),
							"web_url":              knownvalue.StringExact(""),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("limits"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"daily":               knownvalue.Int64Exact(0),
							"maximum_duration":    knownvalue.Int64Exact(0),
							"messages_per_second": knownvalue.Int64Exact(0),
							"total":               knownvalue.Int64Exact(0),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("quiet_time"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"end":   knownvalue.StringExact(""),
							"start": knownvalue.StringExact(""),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
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

func TestAccPinpointApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpoint.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPinpointApp_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
				),
			},
		},
	})
}

func TestAccPinpointApp_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
		},
	})
}

func TestAccPinpointApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
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
				Config: testAccAppConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccPinpointApp_campaignHookLambda(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_campaignHookLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrPair(resourceName, "campaign_hook.0.lambda_function_name", "aws_lambda_function.test", names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("campaign_hook"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"lambda_function_name": knownvalue.NotNull(), // Should be a Pair function, waiting on https://github.com/hashicorp/terraform-plugin-testing/pull/330
							names.AttrMode:         knownvalue.StringExact(pinpoint.ModeDelivery),
							"web_url":              knownvalue.StringExact(""),
						}),
					})),
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

func TestAccPinpointApp_campaignHookEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_campaignHookEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("campaign_hook"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"lambda_function_name": knownvalue.StringExact(""),
							names.AttrMode:         knownvalue.StringExact(""),
							"web_url":              knownvalue.StringExact(""),
						}),
					})),
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

func TestAccPinpointApp_limits(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_limits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("limits"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"daily":               knownvalue.Int64Exact(3),
							"maximum_duration":    knownvalue.Int64Exact(600),
							"messages_per_second": knownvalue.Int64Exact(50),
							"total":               knownvalue.Int64Exact(100),
						}),
					})),
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

func TestAccPinpointApp_limitsEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_limitsEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("limits"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"daily":               knownvalue.Int64Exact(0),
							"maximum_duration":    knownvalue.Int64Exact(0),
							"messages_per_second": knownvalue.Int64Exact(0),
							"total":               knownvalue.Int64Exact(0),
						}),
					})),
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

func TestAccPinpointApp_quietTime(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_quietTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("quiet_time"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"end":   knownvalue.StringExact("03:00"),
							"start": knownvalue.StringExact("00:00"),
						}),
					})),
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

func TestAccPinpointApp_quietTimeEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_quietTimeEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("quiet_time"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"end":   knownvalue.StringExact(""),
							"start": knownvalue.StringExact(""),
						}),
					})),
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

func testAccPreCheckApp(ctx context.Context, t *testing.T) {
	t.Helper()
	acctest.PreCheckPinpointApp(ctx, t)
}

func testAccCheckAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_app" {
				continue
			}

			_, err := tfpinpoint.FindAppByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint App %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, n string, v *pinpoint.ApplicationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		output, err := tfpinpoint.FindAppByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppConfig_nameGenerated() string {
	return `
resource "aws_pinpoint_app" "test" {}
`
}

func testAccAppConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppConfig_campaignHookLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  campaign_hook {
    lambda_function_name = aws_lambda_function.test.arn
    mode                 = "DELIVERY"
  }

  depends_on = [aws_lambda_permission.test]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdapinpoint.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  runtime       = "nodejs16.x"
  publish       = true
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "aws" {}

data "aws_region" "current" {}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromPinpoint"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "pinpoint.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  source_arn    = "arn:${data.aws_partition.current.partition}:mobiletargeting:${data.aws_region.current.name}:${data.aws_caller_identity.aws.account_id}:/apps/*"
}
`, rName)
}

func testAccAppConfig_campaignHookEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  campaign_hook {}
}
`, rName)
}

func testAccAppConfig_limits(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  limits {
    daily               = 3
    maximum_duration    = 600
    messages_per_second = 50
    total               = 100
  }
}
`, rName)
}

func testAccAppConfig_limitsEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  limits {}
}
`, rName)
}

func testAccAppConfig_quietTime(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  quiet_time {
    start = "00:00"
    end   = "03:00"
  }
}
`, rName)
}

func testAccAppConfig_quietTimeEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  quiet_time {}
}
`, rName)
}
