// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"regexp"
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
)

func TestAccServiceCatalogProvisionedProduct_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", servicecatalog.ServiceName, regexp.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC ID",
						"key":         "VpcID",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]*regexp.Regexp{
						"value": regexp.MustCompile(`vpc-.+`),
					}),
					resource.TestCheckResourceAttrPair(resourceName, "path_id", "data.aws_servicecatalog_launch_paths.test", "summaries.0.path_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", "aws_servicecatalog_product.test", "provisioning_artifact_parameters.0.name"),
					resource.TestCheckResourceAttr(resourceName, "status", servicecatalog.StatusAvailable),
					resource.TestCheckResourceAttr(resourceName, "type", "CFN_STACK"),
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.10.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", servicecatalog.ServiceName, regexp.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC ID",
						"key":         "VpcID",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]*regexp.Regexp{
						"value": regexp.MustCompile(`vpc-.+`),
					}),
					resource.TestCheckResourceAttrPair(resourceName, "path_id", "data.aws_servicecatalog_launch_paths.test", "summaries.0.path_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", "aws_servicecatalog_product.test", "provisioning_artifact_parameters.0.name"),
					resource.TestCheckResourceAttr(resourceName, "status", servicecatalog.StatusAvailable),
					resource.TestCheckResourceAttr(resourceName, "type", "CFN_STACK"),
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.failure_tolerance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.max_concurrency_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.regions.#", "1"),
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
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.failure_tolerance_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.max_concurrency_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.regions.#", "1"),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", "0"),
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_productName(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16", productName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttrPair(resourceName, "product_name", "aws_servicecatalog_product.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
				),
			},
			{
				// update the product name, but keep provisioned product name as-is to trigger an in-place update
				Config: testAccProvisionedProductConfig_productName(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16", productNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttrPair(resourceName, "product_name", "aws_servicecatalog_product.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod1, pprod2 servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{

				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod1),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", productResourceName, "provisioning_artifact_parameters.0.name"),
				),
			},
			{
				Config: testAccProvisionedProductConfig_ProvisionedArtifactName_update(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16", artifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod2),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", artifactResourceName, "name"),
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_computedOutputs(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC ID",
						"key":         "VpcID",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC CIDR",
						"key":         "VPCPrimaryCIDR",
						"value":       "10.1.0.0/16",
					}),
				),
			},
			{
				Config: testAccProvisionedProductConfig_computedOutputs(rName, domain, acctest.DefaultEmailAddress, "10.1.0.1/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC ID",
						"key":         "VpcID",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"description": "VPC CIDR",
						"key":         "VPCPrimaryCIDR",
						"value":       "10.1.0.1/16",
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
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceProvisionedProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_tags(rName, "Name", rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccProvisionedProductConfig_tags(rName, "NotName", rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.NotName", rName),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_errorOnCreate(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccProvisionedProductConfig_error(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				ExpectError: regexp.MustCompile(`AmazonCloudFormationException  Unresolved resource dependencies \[MyVPC\] in the Outputs block of the template`),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_errorOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())
	var pprod servicecatalog.ProvisionedProductDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
				),
			},
			{
				Config:      testAccProvisionedProductConfig_error(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				ExpectError: regexp.MustCompile(`AmazonCloudFormationException  Unresolved resource dependencies \[MyVPC\] in the Outputs block of the template`),
			},
			{
				// Check we can still run a complete apply after the previous update error
				Config: testAccProvisionedProductConfig_basic(rName, domain, acctest.DefaultEmailAddress, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, resourceName, &pprod),
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
  portfolio_id = aws_servicecatalog_principal_portfolio_association.test.portfolio_id # avoid depends_on
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
  product_id = aws_servicecatalog_product_portfolio_association.test.product_id # avoid depends_on
}
`, rName)
}

func testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email string) string {
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
`, rName, domain, email))
}

func testAccProvisionedProductPhysicalTemplateIDBaseConfig(rName, domain, email string) string {
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
  support_email       = %[3]q
  support_url         = %[2]q

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
`, rName, domain, email))
}

func testAccProvisionedProductConfig_basic(rName, domain, email, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email),
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

func testAccProvisionedProductConfig_computedOutputs(rName, domain, email, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductPhysicalTemplateIDBaseConfig(rName, domain, email),
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

func testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, domain, email, vpcCidr string, failureToleranceCount, maxConcurrencyCount int) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email),
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

func testAccProvisionedProductConfig_productName(rName, domain, email, vpcCidr, productName string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(productName, domain, email),
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

func testAccProvisionedProductConfig_ProvisionedArtifactName_update(rName, domain, email, vpcCidr, artifactName string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email),
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
}
`, rName, vpcCidr, artifactName))
}

// Because the `provisioning_parameter` "LeaveMeEmpty" is not empty, this configuration results in an error.
// The `status_message` will be:
// AmazonCloudFormationException  Unresolved resource dependencies [MyVPC] in the Outputs block of the template
func testAccProvisionedProductConfig_error(rName, domain, email, vpcCidr string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email),
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

func testAccProvisionedProductConfig_tags(rName, tagKey, tagValue, domain, email string) string {
	return acctest.ConfigCompose(testAccProvisionedProductTemplateURLBaseConfig(rName, domain, email),
		fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_constraint.test.product_id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "VPCPrimaryCIDR"
    value = "10.2.0.0/16"
  }

  provisioning_parameters {
    key   = "LeaveMeEmpty"
    value = ""
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue))
}
