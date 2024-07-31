// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTProvisioningTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "FLEET_PROVISIONING"),
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

func TestAccIoTProvisioningTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceProvisioningTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProvisioningTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 1),
				),
			},
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 1),
				),
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProvisioningTemplateConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, rName, 2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "For testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
				),
			},
		},
	})
}

func testAccCheckProvisioningTemplateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Provisioning Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindProvisioningTemplateByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckProvisioningTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_provisioning_template" {
				continue
			}

			_, err := tfiot.FindProvisioningTemplateByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Provisioning Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisioningTemplateNumVersions(ctx context.Context, name string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		var got int
		out, err := conn.ListProvisioningTemplateVersions(ctx, &iot.ListProvisioningTemplateVersionsInput{TemplateName: aws.String(name)})

		if err != nil {
			return err
		}

		if len(out.Versions) != want {
			return fmt.Errorf("Incorrect version count for IoT Provisioning Template %s; got: %d, want: %d", name, got, want)
		}

		return nil
	}
}

func testAccProvisioningTemplateBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

data "aws_iam_policy_document" "device" {
  statement {
    actions   = ["iot:Subscribe"]
    resources = ["*"]
  }
}

resource "aws_iot_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.device.json
}
`, rName)
}

func testAccProvisioningTemplateConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProvisioningTemplateBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.test.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })
}
`, rName))
}

func testAccProvisioningTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccProvisioningTemplateBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.test.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccProvisioningTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccProvisioningTemplateBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.test.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccProvisioningTemplateConfig_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisioningTemplateBaseConfig(rName),
		testAccProvisioningTemplateConfig_preProvisioningHook(rName),
		fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn
  description           = "For testing"
  enabled               = true

  pre_provisioning_hook {
    target_arn = aws_lambda_function.test.arn
  }

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Inactive"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.test.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })
}
`, rName))
}

func testAccProvisioningTemplateConfig_preProvisioningHook(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromIot"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "iot.amazonaws.com"
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambda-preprovisioninghook.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambda-preprovisioninghook.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test2.arn
  handler          = "lambda-preprovisioninghook.handler"
  runtime          = "nodejs20.x"
}
`, rName)
}
