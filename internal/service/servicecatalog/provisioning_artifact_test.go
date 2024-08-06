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

// add sweeper to delete known test servicecat provisioning artifacts

func TestAccServiceCatalogProvisioningArtifact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningArtifactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningArtifactConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisioningArtifactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "disable_template_validation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_id"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, servicecatalog.ProductTypeCloudFormationTemplate),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_url",
				},
			},
		},
	})
}

func TestAccServiceCatalogProvisioningArtifact_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningArtifactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningArtifactConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisioningArtifactExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceProvisioningArtifact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogProvisioningArtifact_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningArtifactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningArtifactConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisioningArtifactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-2", rName)),
				),
			},
			{
				Config: testAccProvisioningArtifactConfig_update(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "accept_language", "jp"),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s-3", rName)),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDeprecated),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-3", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_url",
				},
			},
		},
	})
}

func TestAccServiceCatalogProvisioningArtifact_physicalID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisioningArtifactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningArtifactConfig_physicalID(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisioningArtifactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "disable_template_validation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "template_physical_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, servicecatalog.ProductTypeCloudFormationTemplate),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_physical_id",
				},
			},
		},
	})
}

func testAccCheckProvisioningArtifactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_provisioning_artifact" {
				continue
			}

			artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(rs.Primary.ID)

			if err != nil {
				return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
			}

			input := &servicecatalog.DescribeProvisioningArtifactInput{
				ProductId:              aws.String(productID),
				ProvisioningArtifactId: aws.String(artifactID),
			}

			output, err := conn.DescribeProvisioningArtifactWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Service Catalog Provisioning Artifact (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckProvisioningArtifactExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
		}

		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
		}

		_, err = conn.DescribeProvisioningArtifactWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccProvisioningArtifactTemplateURLBaseConfig(rName, domain string) string {
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
  description         = %[1]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[1]q
  support_email       = %[3]q
  support_url         = %[2]q

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, domain, acctest.DefaultEmailAddress)
}

func testAccProvisioningArtifactConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccProvisioningArtifactTemplateURLBaseConfig(rName, domain), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "en"
  active                      = true
  description                 = %[1]q
  disable_template_validation = true
  guidance                    = "DEFAULT"
  name                        = "%[1]s-2"
  product_id                  = aws_servicecatalog_product.test.id
  template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}

func testAccProvisioningArtifactConfig_update(rName, domain string) string {
	return acctest.ConfigCompose(testAccProvisioningArtifactTemplateURLBaseConfig(rName, domain), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "jp"
  active                      = false
  description                 = "%[1]s-3"
  disable_template_validation = true
  guidance                    = "DEPRECATED"
  name                        = "%[1]s-3"
  product_id                  = aws_servicecatalog_product.test.id
  template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}

func testAccProvisioningArtifactPhysicalIDBaseConfig(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
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
  description         = %[1]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = %[3]q
  support_url         = %[2]q

  provisioning_artifact_parameters {
    description          = "artefaktbeskrivning"
    name                 = %[1]q
    template_physical_id = aws_cloudformation_stack.test.id
    type                 = "CLOUD_FORMATION_TEMPLATE"
  }
}
`, rName, domain, acctest.DefaultEmailAddress)
}

func testAccProvisioningArtifactConfig_physicalID(rName, domain string) string {
	return acctest.ConfigCompose(testAccProvisioningArtifactPhysicalIDBaseConfig(rName, domain), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "en"
  active                      = true
  description                 = %[1]q
  disable_template_validation = false
  guidance                    = "DEFAULT"
  name                        = "%[1]s-2"
  product_id                  = aws_servicecatalog_product.test.id
  template_physical_id        = aws_cloudformation_stack.test.id
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}
