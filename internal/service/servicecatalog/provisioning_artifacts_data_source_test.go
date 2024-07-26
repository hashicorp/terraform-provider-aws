// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogProvisioningArtifactsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicecatalog_provisioning_artifacts.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisioningArtifactsDataSourceConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttrPair(dataSourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(dataSourceName, "provisioning_artifact_details.0.description"),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
		},
	})
}

func testAccProvisioningArtifactsDataSourceConfig_base(rName, domain string) string {
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
`, rName, domain, acctest.DefaultEmailAddress)
}

func testAccProvisioningArtifactsDataSourceConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccProvisioningArtifactsDataSourceConfig_base(rName, domain), `
data "aws_servicecatalog_provisioning_artifacts" "test" {
  product_id = aws_servicecatalog_product.test.id
}
`)
}
