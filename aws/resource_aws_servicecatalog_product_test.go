package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "distributor", "distributör"),
					resource.TestCheckResourceAttr(resourceName, "has_default_path", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", "ägare"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_url"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.ProductStatusCreated),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_email", "support@example.com"),
					resource.TestCheckResourceAttr(resourceName, "support_url", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateTags(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_updateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "natural"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeProductAsAdmin(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Product (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProductExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProductAsAdmin(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProductConfigTemplateURLBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "AWS Service Catalog sample template. Creates an Amazon EC2 instance running the Amazon Linux AMI. The AMI is chosen based on the region in which the stack is run. This example creates an EC2 security group for the instance to give you SSH access. **WARNING** This template creates an Amazon EC2 instance. You will be billed for the AWS resources used if you create a stack from this template.",
  "Parameters": {
    "KeyName": {
      "Description": "Name of an existing EC2 key pair for SSH access to the EC2 instance.",
      "Type": "AWS::EC2::KeyPair::KeyName"
    },
    "InstanceType": {
      "Description": "EC2 instance type.",
      "Type": "String",
      "Default": "t2.micro",
      "AllowedValues": [
        "t2.micro",
        "t2.small",
        "t2.medium",
        "m3.medium",
        "m3.large",
        "m3.xlarge",
        "m3.2xlarge"
      ]
    },
    "SSHLocation": {
      "Description": "The IP address range that can SSH to the EC2 instance.",
      "Type": "String",
      "MinLength": "9",
      "MaxLength": "18",
      "Default": "0.0.0.0/0",
      "AllowedPattern": "(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})/(\\d{1,2})",
      "ConstraintDescription": "Must be a valid IP CIDR range of the form x.x.x.x/x."
    }
  },
  "Metadata": {
    "AWS::CloudFormation::Interface": {
      "ParameterGroups": [
        {
          "Label": {
            "default": "Instance configuration"
          },
          "Parameters": [
            "InstanceType"
          ]
        },
        {
          "Label": {
            "default": "Security configuration"
          },
          "Parameters": [
            "KeyName",
            "SSHLocation"
          ]
        }
      ],
      "ParameterLabels": {
        "InstanceType": {
          "default": "Server size:"
        },
        "KeyName": {
          "default": "Key pair:"
        },
        "SSHLocation": {
          "default": "CIDR range:"
        }
      }
    }
  },
  "Mappings": {
    "AWSRegionArch2AMI": {
      "us-east-1": {
        "HVM64": "ami-08842d60"
      },
      "us-west-2": {
        "HVM64": "ami-8786c6b7"
      },
      "us-west-1": {
        "HVM64": "ami-cfa8a18a"
      },
      "eu-west-1": {
        "HVM64": "ami-748e2903"
      },
      "ap-southeast-1": {
        "HVM64": "ami-d6e1c584"
      },
      "ap-northeast-1": {
        "HVM64": "ami-35072834"
      },
      "ap-southeast-2": {
        "HVM64": "ami-fd4724c7"
      },
      "sa-east-1": {
        "HVM64": "ami-956cc688"
      },
      "cn-north-1": {
        "HVM64": "ami-ac57c595"
      },
      "eu-central-1": {
        "HVM64": "ami-b43503a9"
      }
    }
  },
  "Resources": {
    "EC2Instance": {
      "Type": "AWS::EC2::Instance",
      "Properties": {
        "InstanceType": {
          "Ref": "InstanceType"
        },
        "SecurityGroups": [
          {
            "Ref": "InstanceSecurityGroup"
          }
        ],
        "KeyName": {
          "Ref": "KeyName"
        },
        "ImageId": {
          "Fn::FindInMap": [
            "AWSRegionArch2AMI",
            {
              "Ref": "AWS::Region"
            },
            "HVM64"
          ]
        }
      }
    },
    "InstanceSecurityGroup": {
      "Type": "AWS::EC2::SecurityGroup",
      "Properties": {
        "GroupDescription": "Enable SSH access via port 22",
        "SecurityGroupIngress": [
          {
            "IpProtocol": "tcp",
            "FromPort": "22",
            "ToPort": "22",
            "CidrIp": {
              "Ref": "SSHLocation"
            }
          }
        ]
      }
    }
  },
  "Outputs": {
    "PublicDNSName": {
      "Description": "Public DNS name of the new EC2 instance",
      "Value": {
        "Fn::GetAtt": [
          "EC2Instance",
          "PublicDnsName"
        ]
      }
    },
    "PublicIPAddress": {
      "Description": "Public IP address of the new EC2 instance",
      "Value": {
        "Fn::GetAtt": [
          "EC2Instance",
          "PublicIp"
        ]
      }
    }
  }
}
EOF
}
`, rName)
}

func testAccAWSServiceCatalogProductConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
	disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSServiceCatalogProductConfig_updateTags(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
	disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Yak         = %[1]q
	Environment = "natural"
  }
}
`, rName))
}
