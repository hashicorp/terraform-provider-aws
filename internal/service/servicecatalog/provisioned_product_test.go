// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccServiceCatalogProvisionedProduct_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, servicecatalog.ServiceName, regexache.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC ID",
						names.AttrKey:         "VpcID",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]*regexp.Regexp{
						names.AttrValue: regexache.MustCompile(`vpc-.+`),
					}),
					resource.TestCheckResourceAttrPair(resourceName, "path_id", "data.aws_servicecatalog_launch_paths.test", "summaries.0.path_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", "aws_servicecatalog_product.test", "provisioning_artifact_parameters.0.name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, servicecatalog.StatusAvailable),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CFN_STACK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"ignore_errors",
					"provisioning_artifact_name",
					"provisioning_parameters",
					"retain_physical_resources",
				},
			},
		},
	})
}

// TestAccServiceCatalogProvisionedProduct_update verifies the resource update
// of only a change in provisioning_parameters
func TestAccServiceCatalogProvisionedProduct_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.10.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, servicecatalog.ServiceName, regexache.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC ID",
						names.AttrKey:         "VpcID",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]*regexp.Regexp{
						names.AttrValue: regexache.MustCompile(`vpc-.+`),
					}),
					resource.TestCheckResourceAttrPair(resourceName, "path_id", "data.aws_servicecatalog_launch_paths.test", "summaries.0.path_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", "aws_servicecatalog_product.test", "provisioning_artifact_parameters.0.name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, servicecatalog.StatusAvailable),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CFN_STACK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"ignore_errors",
					"provisioning_artifact_name",
					"provisioning_parameters",
					"retain_physical_resources",
				},
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_stackSetProvisioningPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, "10.1.0.0/16", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.failure_tolerance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.max_concurrency_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.regions.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"ignore_errors",
					"provisioning_artifact_name",
					"provisioning_parameters",
					"retain_physical_resources",
					"stack_set_provisioning_preferences",
				},
			},
			{
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, "10.1.0.0/16", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.failure_tolerance_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.max_concurrency_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.regions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_ProductName_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	productName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	productNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_productName(rName, "10.1.0.0/16", productName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttrPair(resourceName, "product_name", "aws_servicecatalog_product.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
				),
			},
			{
				// update the product name, but keep provisioned product name as-is to trigger an in-place update
				Config: testAccProvisionedProductConfig_productName(rName, "10.1.0.0/16", productNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttrPair(resourceName, "product_name", "aws_servicecatalog_product.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"ignore_errors",
					"product_name",
					"provisioning_artifact_name",
					"provisioning_parameters",
					"retain_physical_resources",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/26271
func TestAccServiceCatalogProvisionedProduct_ProvisioningArtifactName_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	productResourceName := "aws_servicecatalog_product.test"
	artifactResourceName := "aws_servicecatalog_provisioning_artifact.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	artifactName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod1, pprod2 servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{

				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod1),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", productResourceName, "provisioning_artifact_parameters.0.name"),
				),
			},
			{
				Config: testAccProvisionedProductConfig_ProvisionedArtifactName_update(rName, "10.1.0.0/16", artifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod2),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", artifactResourceName, names.AttrName),
					testAccCheckProvisionedProductProvisioningArtifactIDChanged(&pprod1, &pprod2),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_computedOutputs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_computedOutputs(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC ID",
						names.AttrKey:         "VpcID",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC CIDR",
						names.AttrKey:         "VPCPrimaryCIDR",
						names.AttrValue:       "10.1.0.0/16",
					}),
				),
			},
			{
				Config: testAccProvisionedProductConfig_computedOutputs(rName, "10.1.0.1/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC ID",
						names.AttrKey:         "VpcID",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrDescription: "VPC CIDR",
						names.AttrKey:         "VPCPrimaryCIDR",
						names.AttrValue:       "10.1.0.1/16",
					}),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceProvisionedProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_errorOnCreate(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccProvisionedProductConfig_error(rName, "10.1.0.0/16"),
				ExpectError: regexache.MustCompile(`AmazonCloudFormationException  Unresolved resource dependencies \[MyVPC\] in the Outputs block of the template`),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_errorOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
				),
			},
			{
				Config:      testAccProvisionedProductConfig_error(rName, "10.1.0.0/16"),
				ExpectError: regexache.MustCompile(`AmazonCloudFormationException  Unresolved resource dependencies \[MyVPC\] in the Outputs block of the template`),
			},
			{
				// Check we can still run a complete apply after the previous update error
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_productTagUpdateAfterError(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_productTagUpdateAfterError_valid(rName, bucketName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.version", "1.0"),
					acctest.S3BucketHasTag(ctx, bucketName, names.AttrVersion, "1.0"),
				),
			},
			{
				Config:      testAccProvisionedProductConfig_productTagUpdateAfterError_confict(rName, bucketName, "1.5"),
				ExpectError: regexache.MustCompile(`BucketAlreadyOwnedByYou`),
			},
			{
				Config: testAccProvisionedProductConfig_productTagUpdateAfterError_valid(rName, bucketName, "1.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.version", "1.5"),
					acctest.S3BucketHasTag(ctx, bucketName, names.AttrVersion, "1.5"),
				),
			},
		},
	})
}

func testAccCheckProvisionedProductDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_provisioned_product" {
				continue
			}

			input := &servicecatalog.DescribeProvisionedProductInput{
				Id:             aws.String(rs.Primary.ID),
				AcceptLanguage: aws.String(rs.Primary.Attributes["accept_language"]),
			}
			_, err := conn.DescribeProvisionedProductWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Catalog Provisioned Product (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisionedProductExists(ctx context.Context, resourceName string, pprod *servicecatalog.ProvisionedProductDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		out, err := tfservicecatalog.WaitProvisionedProductReady(ctx, conn, tfservicecatalog.AcceptLanguageEnglish, rs.Primary.ID, "", tfservicecatalog.ProvisionedProductReadyTimeout)
		if err != nil {
			return fmt.Errorf("describing Service Catalog Provisioned Product (%s): %w", rs.Primary.ID, err)
		}

		*pprod = *out.ProvisionedProductDetail

		return nil
	}
}

// testAccCheckProvisionedProductProvisioningArtifactIDChanged verifies that the provisioned artifact
// ID differs between two provisioned products. If either provisioned product details or the provisioned
// artifact ID are null, the check will fail.
func testAccCheckProvisionedProductProvisioningArtifactIDChanged(pprod1, pprod2 *servicecatalog.ProvisionedProductDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pprod1 == nil || pprod2 == nil ||
			pprod1.ProvisioningArtifactId == nil ||
			pprod2.ProvisioningArtifactId == nil {
			return fmt.Errorf("provisioned product provisioning artifact ID is nil")
		}
		if aws.StringValue(pprod1.ProvisioningArtifactId) == aws.StringValue(pprod2.ProvisioningArtifactId) {
			return fmt.Errorf("provisioned product provisioning artifact ID has not changed")
		}

		return nil
	}
}

func testAccProvisionedProductPortfolioBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_constraint" "test" {
  description  = %[1]q
  portfolio_id = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product_portfolio_association.test.product_id
  type         = "RESOURCE_UPDATE"

  parameters = jsonencode({
    Version = "2.0"
    Properties = {
      TagUpdateOnProvisionedProduct = "ALLOWED"
    }
  })
}

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_principal_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product.test.id
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = data.aws_iam_session_context.current.issuer_arn # unfortunately, you cannot get launch_path for arbitrary role - only caller
}

data "aws_servicecatalog_launch_paths" "test" {
  product_id = aws_servicecatalog_product_portfolio_association.test.product_id
}
`, rName)
}

func testAccProvisionedProductTemplateURLBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisionedProductPortfolioBaseConfig(rName),
		fmt.Sprintf(`
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
      VPCPrimaryCIDR = {
        Type = "String"
      }
      LeaveMeEmpty = {
        Type        = "String"
        Description = "Make sure that empty values come through. Issue #21349"
      }
    }

    "Conditions" = {
      "IsEmptyParameter" = {
        "Fn::Equals" = [
          {
            "Ref" = "LeaveMeEmpty"
          },
          "",
        ]
      }
    }

    Resources = {
      MyVPC = {
        Type      = "AWS::EC2::VPC"
        Condition = "IsEmptyParameter"
        Properties = {
          CidrBlock = { Ref = "VPCPrimaryCIDR" }
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

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }
}
`, rName))
}

func testAccProvisionedProductTemplateURLSimpleBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisionedProductPortfolioBaseConfig(rName),
		fmt.Sprintf(`
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
      BucketName = {
        Type = "String"
      }
    }

    Resources = {
      MyS3Bucket = {
        Type = "AWS::S3::Bucket"
        Properties = {
          BucketName = { Ref = "BucketName" }
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description = %[1]q
  distributor = "distributör"
  name        = %[1]q
  owner       = "ägare"
  type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }
}
`, rName))
}

func testAccProvisionedProductPhysicalTemplateIDBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisionedProductPortfolioBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Parameters = {
      VPCPrimaryCIDR = {
        Type    = "String"
        Default = "10.0.0.0/16"
      }
      LeaveMeEmpty = {
        Type        = "String"
        Description = "Make sure that empty values come through. Issue #21349"
        Default     = ""
      }
    }

    "Conditions" = {
      "IsEmptyParameter" = {
        "Fn::Equals" = [
          {
            "Ref" = "LeaveMeEmpty"
          },
          "",
        ]
      }
    }

    Resources = {
      MyVPC = {
        Type      = "AWS::EC2::VPC"
        Condition = "IsEmptyParameter"
        Properties = {
          CidrBlock = { Ref = "VPCPrimaryCIDR" }
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

      VPCPrimaryCIDR = {
        Description = "VPC CIDR"
        Value = {
          Ref = "VPCPrimaryCIDR"
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

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_physical_id        = aws_cloudformation_stack.test.id
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccProvisionedProductConfig_basic(rName, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }

  # Leave this here to test tag behavior on Update
  tags = {
    Name = %[1]q
  }
}
`, rName, vpcCidr))
}

func testAccProvisionedProductConfig_computedOutputs(rName, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductPhysicalTemplateIDBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }
}
`, rName, vpcCidr))
}

func testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, vpcCidr string, failureToleranceCount, maxConcurrencyCount int) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  stack_set_provisioning_preferences {
    accounts                = [data.aws_caller_identity.current.account_id]
    regions                 = [data.aws_region.current.name]
    failure_tolerance_count = %[3]d
    max_concurrency_count   = %[4]d
  }

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }
}
`, rName, vpcCidr, failureToleranceCount, maxConcurrencyCount))
}

func testAccProvisionedProductConfig_productName(rName, vpcCidr, productName string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(productName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_name               = aws_servicecatalog_product.test.name
  provisioning_artifact_name = aws_servicecatalog_product.test.provisioning_artifact_parameters[0].name
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }
}
`, rName, vpcCidr))
}

func testAccProvisionedProductConfig_ProvisionedArtifactName_update(rName, vpcCidr, artifactName string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  product_id   = aws_servicecatalog_product.test.id
  template_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  name         = %[3]q
  type         = "CLOUD_FORMATION_TEMPLATE"
}

resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = aws_servicecatalog_provisioning_artifact.test.name
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }

  # Leave this here to test tag behavior on Update
  tags = {
    Name = %[1]q
  }
}
`, rName, vpcCidr, artifactName))
}

// Because the `provisioning_parameter` "LeaveMeEmpty" is not empty, this configuration results in an error.
// The `status_message` will be:
// AmazonCloudFormationException  Unresolved resource dependencies [MyVPC] in the Outputs block of the template
func testAccProvisionedProductConfig_error(rName, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = %[2]q
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = "NotEmpty"
  }
}
`, rName, vpcCidr))
}

func testAccProvisionedProductConfig_productTagUpdateAfterError_valid(rName, bucketName, tagValue string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLSimpleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "BucketName"
    value = %[2]q
  }

  tags = {
    version = %[3]q
  }
}
`, rName, bucketName, tagValue))
}

func testAccProvisionedProductConfig_productTagUpdateAfterError_confict(rName, conflictingBucketName, tagValue string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLSimpleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "BucketName"
    value = aws_s3_bucket.conflict.bucket
  }

  tags = {
    version = %[3]q
  }
}

resource "aws_s3_bucket" "conflict" {
  bucket = %[2]q
}
`, rName, conflictingBucketName, tagValue))
}
