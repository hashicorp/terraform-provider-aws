// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTProvisioningTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "iot", "provisioningtemplate/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiot.ResourceProvisioningTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
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
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
				),
			},
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
				),
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "iot", "provisioningtemplate/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 2),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "iot", "provisioningtemplate/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "For testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/38629.
func TestAccIoTProvisioningTemplate_jitp(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_jitp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(ctx, t, resourceName),
					testAccCheckProvisioningTemplateNumVersions(ctx, t, rName, 1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "iot", "provisioningtemplate/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "JITP"),
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

func testAccCheckProvisioningTemplateExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindProvisioningTemplateByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckProvisioningTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_provisioning_template" {
				continue
			}

			_, err := tfiot.FindProvisioningTemplateByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckProvisioningTemplateNumVersions(ctx context.Context, t *testing.T, name string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

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

func testAccProvisioningTemplateConfig_base(rName string) string {
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
	return acctest.ConfigCompose(testAccProvisioningTemplateConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccProvisioningTemplateConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccProvisioningTemplateConfig_base(rName), fmt.Sprintf(`
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

func testAccProvisioningTemplateConfig_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisioningTemplateConfig_base(rName),
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

func testAccProvisioningTemplateConfig_jitp(rName string) string {
	return acctest.ConfigCompose(testAccProvisioningTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn
  type                  = "JITP"

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
