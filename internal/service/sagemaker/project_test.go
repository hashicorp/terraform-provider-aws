// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_catalog_provisioning_details.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "service_catalog_provisioning_details.0.product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_base(rName),
			},
		},
	})
}

func TestAccSageMakerProject_description(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
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
				Config: testAccProjectConfig_description(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestCheckResourceAttr(resourceName, "project_description", rNameUpdated),
				),
			},
			{
				Config: testAccProjectConfig_base(rName),
			},
		},
	})
}

func TestAccSageMakerProject_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
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
				Config: testAccProjectConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProjectConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProjectConfig_base(rName),
			},
		},
	})
}

func TestAccSageMakerProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &mpg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceProject(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_project" {
				continue
			}

			Project, err := tfsagemaker.FindProjectByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Project (%s): %w", rs.Primary.ID, err)
			}

			if aws.StringValue(Project.ProjectName) == rs.Primary.ID {
				return fmt.Errorf("sagemaker Project %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, n string, mpg *sagemaker.DescribeProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Project ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindProjectByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccProjectConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
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

func testAccProjectConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name = %[1]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }
}
`, rName))
}

func testAccProjectConfig_description(rName, desc string) string {
	return acctest.ConfigCompose(testAccProjectConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_project" "test" {
  project_name        = %[1]q
  project_description = %[2]q

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_constraint.test.product_id
  }
}
`, rName, desc))
}

func testAccProjectConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_base(rName), fmt.Sprintf(`
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

func testAccProjectConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_base(rName), fmt.Sprintf(`
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
