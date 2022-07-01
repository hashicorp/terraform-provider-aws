package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTProvisioningTemplate_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisioningTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					testAccCheckProvisioningTemplateNumVersions(rName, 1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisioningTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceProvisioningTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisioningTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					testAccCheckProvisioningTemplateNumVersions(rName, 1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProvisioningTemplateConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckProvisioningTemplateNumVersions(rName, 1),
				),
			},
			{
				Config: testAccProvisioningTemplateConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckProvisioningTemplateNumVersions(rName, 1),
				),
			},
		},
	})
}

func TestAccIoTProvisioningTemplate_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_provisioning_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisioningTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisioningTemplateExists(resourceName),
					testAccCheckProvisioningTemplateNumVersions(rName, 1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckProvisioningTemplateExists(resourceName),
					testAccCheckProvisioningTemplateNumVersions(rName, 2),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "For testing"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pre_provisioning_hook.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
				),
			},
		},
	})
}

func testAccCheckProvisioningTemplateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Provisioning Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		_, err := tfiot.FindProvisioningTemplateByName(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckProvisioningTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_provisioning_template" {
			continue
		}

		_, err := tfiot.FindProvisioningTemplateByName(context.TODO(), conn, rs.Primary.ID)

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

func testAccCheckProvisioningTemplateNumVersions(name string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		var got int
		err := conn.ListProvisioningTemplateVersionsPages(
			&iot.ListProvisioningTemplateVersionsInput{TemplateName: aws.String(name)},
			func(page *iot.ListProvisioningTemplateVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				got += len(page.Versions)

				return !lastPage
			})

		if err != nil {
			return err
		}

		if got != want {
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
	return acctest.ConfigCompose(testAccProvisioningTemplateBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_provisioning_template" "test" {
  name                  = %[1]q
  provisioning_role_arn = aws_iam_role.test.arn
  description           = "For testing"
  enabled               = true

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
