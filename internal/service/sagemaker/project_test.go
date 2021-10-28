package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerProject_basic(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_catalog_provisioning_details.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "service_catalog_provisioning_details.0.product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
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

func TestAccSageMakerProject_description(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_description", rName),
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

func TestAccSageMakerProject_tags(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccProjectTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerProject_disappears(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceProject(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_project" {
			continue
		}

		Project, err := tfsagemaker.FindProjectByName(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "does not exist") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Project (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(Project.ProjectName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Project %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckProjectExists(n string, mpg *sagemaker.DescribeProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Project ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindProjectByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccProjectBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id   = aws_servicecatalog_product.test.id
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "servicecatalog.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
}

resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_product_portfolio_association.test.product_id
  }
}
`, rName)
}

func testAccProjectDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name        = %[1]q
  project_description = %[1]q
}
`, rName)
}

func testAccProjectTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProjectTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
