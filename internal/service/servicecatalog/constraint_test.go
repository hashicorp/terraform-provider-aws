// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// add sweeper to delete known test servicecat constraints

func TestAccServiceCatalogConstraint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_constraint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConstraintDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConstraintConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraintExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NOTIFICATION"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrParameters),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
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

func TestAccServiceCatalogConstraint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_constraint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConstraintDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConstraintConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraintExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceConstraint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogConstraint_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_constraint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConstraintDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConstraintConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				Config: testAccConstraintConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName2),
				),
			},
		},
	})
}

func testAccCheckConstraintDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_constraint" {
				continue
			}

			input := &servicecatalog.DescribeConstraintInput{
				Id: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeConstraintWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Service Catalog Constraint (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Service Catalog Constraint (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckConstraintExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeConstraintWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Constraint (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccConstraintConfig_base(rName string) string {
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
  owner = "ägare"
  type  = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id   = aws_servicecatalog_product.test.id
}
`, rName)
}

func testAccConstraintConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccConstraintConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_servicecatalog_constraint" "test" {
  description  = %[2]q
  portfolio_id = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product_portfolio_association.test.product_id
  type         = "NOTIFICATION"

  parameters = jsonencode({ "NotificationArns" : [aws_sns_topic.test.arn] })

  depends_on = [aws_sns_topic.test]
}
`, rName, description))
}
