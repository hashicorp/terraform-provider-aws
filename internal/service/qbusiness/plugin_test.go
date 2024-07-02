// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQBusinessPlugin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var plugin qbusiness.GetPluginOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPlugin(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPluginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "plugin_id"),
					resource.TestCheckResourceAttrSet(resourceName, "basic_auth_configuration.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "basic_auth_configuration.0.secret_arn"),
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

func TestAccQBusinessPlugin_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var plugin qbusiness.GetPluginOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPlugin(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPluginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourcePlugin, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessPlugin_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var plugin qbusiness.GetPluginOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPlugin(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPluginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig_tags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPluginConfig_tags(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, "value2updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value2updated"),
				),
			},
		},
	})
}

func TestAccQBusinessPlugin_customPlugin(t *testing.T) {
	ctx := acctest.Context(t)
	var plugin qbusiness.GetPluginOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPlugin(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPluginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPluginConfig_customPlugin(rName, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					resource.TestCheckResourceAttr(resourceName, "no_auth_configuration.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
				),
			},
			{
				Config: testAccPluginConfig_customPlugin(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPluginExists(ctx, resourceName, &plugin),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
				),
			},
		},
	})
}

func testAccPreCheckPlugin(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

	input := &qbusiness.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPluginDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_plugin" {
				continue
			}

			_, err := tfqbusiness.FindPluginByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q Plugin %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPluginExists(ctx context.Context, n string, v *qbusiness.GetPluginOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindPluginByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPluginConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
variable "credentials" {
  default = {
    username = "username"
    password = "password"
  }
  type = map(string)
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode(var.credentials)
}

resource "aws_iam_policy" "test" {
  policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["secretsmanager:GetSecretValue"]
          Effect   = "Allow"
          Resource = aws_secretsmanager_secret.test.arn
        }
      ]
  })
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rName))
}

func testAccPluginConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPluginConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_plugin" "test" {
  application_id = aws_qbusiness_app.test.id
  basic_auth_configuration {
    role_arn   = aws_iam_role.test.arn
    secret_arn = aws_secretsmanager_secret.test.arn
  }
  display_name = %[1]q
  server_url   = "https://yourinstance.service-now.com"
  state        = "ENABLED"
  type         = "SERVICE_NOW"
}
`, rName))
}

func testAccPluginConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccPluginConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_plugin" "test" {
  application_id = aws_qbusiness_app.test.id
  basic_auth_configuration {
    role_arn   = aws_iam_role.test.arn
    secret_arn = aws_secretsmanager_secret.test.arn
  }
  display_name = %[1]q
  server_url   = "https://yourinstance.service-now.com"
  state        = "ENABLED"
  type         = "SERVICE_NOW"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccPluginConfig_customPlugin(rName, state string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_plugin" "test" {
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  state          = %[2]q
  type           = "CUSTOM"
  custom_plugin_configuration {
    api_schema_type = "OPEN_API_V3"
    description     = "Plugin description"
    payload         = <<SCHEMA
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
servers:
  - url: https://api.example.com/
paths:
  /strings:
    get:
      description: Test
      responses:
        '200':
          description: A JSON array of strings
          content:
            application/json:
              schema: 
                type: array
                items: 
                  type: string
SCHEMA
  }
}
`, rName, state))
}
