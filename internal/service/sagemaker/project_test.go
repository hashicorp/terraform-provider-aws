package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSageMakerProject_basic(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
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
			{
				Config: testAccProjectBaseConfig(rName),
			},
		},
	})
}

func TestAccSageMakerProject_description(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestCheckResourceAttr(resourceName, "project_description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectDescription(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestCheckResourceAttr(resourceName, "project_description", rNameUpdated),
				),
			},
			{
				Config: testAccProjectBaseConfig(rName),
			},
		},
	})
}

func TestAccSageMakerProject_tags(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectTagsConfig1(rName, "key1", "value1"),
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
				Config: testAccProjectTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccProjectTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccProjectBaseConfig(rName),
			},
		},
	})
}

func TestAccSageMakerProject_disappears(t *testing.T) {
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
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

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker Project (%s): %w", rs.Primary.ID, err)
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

func testAccProjectBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Parameters = {
      SageMakerProjectName = {
        Type        = "String"
        Description = "Name of the project"
      }
      SageMakerProjectId = {
        Type        = "String"
        Description = "Service generated Id of the project."
      }
    }

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
  name  = %[1]q
  owner = %[1]q
  type  = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "cloudformation:CreateStack",
          "cloudformation:DeleteStack",
          "cloudformation:DescribeStackEvents",
          "cloudformation:DescribeStacks",
          "cloudformation:GetTemplateSummary",
          "cloudformation:SetStackPolicy",
          "cloudformation:ValidateTemplate",
          "cloudformation:UpdateStack",
          "s3:GetObject",
          "servicecatalog:*",
          "ec2:*"
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_servicecatalog_constraint" "test" {
  portfolio_id = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product_portfolio_association.test.product_id
  type         = "LAUNCH"

  parameters = jsonencode({
    "RoleArn" : aws_iam_role.test.arn
  })

  depends_on = [aws_iam_role_policy.test]
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "test" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  principal_arn = data.aws_iam_session_context.test.issuer_arn
}
`, rName)
}

func testAccProjectBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccProjectBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }
}
`, rName))
}

func testAccProjectDescription(rName, desc string) string {
	return acctest.ConfigCompose(testAccProjectBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name        = %[1]q
  project_description = %[2]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }
}
`, rName, desc))
}

func testAccProjectTagsConfig1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccProjectBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccProjectTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccProjectBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
