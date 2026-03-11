// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogProvisionedProduct_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusAvailable)),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.10.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					// One output will default to the launched CloudFormation Stack (provisioned outside terraform).
					// While another output will describe the output parameter configured in the S3 object resource,
					// which we can check as follows.
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusAvailable)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, "10.1.0.0/16", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccProvisionedProductConfig_stackSetprovisioningPreferences(rName, "10.1.0.0/16", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.failure_tolerance_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.max_concurrency_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.0.regions.#", "1"),
				),
			},
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "stack_set_provisioning_preferences.#", "0"),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_ProductName_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	productName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	productNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_productName(rName, "10.1.0.0/16", productName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttrPair(resourceName, "product_name", "aws_servicecatalog_product.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", names.AttrID),
				),
			},
			{
				// update the product name, but keep provisioned product name as-is to trigger an in-place update
				Config: testAccProvisionedProductConfig_productName(rName, "10.1.0.0/16", productNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	artifactName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod1, pprod2 awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{

				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod1),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", productResourceName, "provisioning_artifact_parameters.0.name"),
				),
			},
			{
				Config: testAccProvisionedProductConfig_ProvisionedArtifactName_update(rName, "10.1.0.0/16", artifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod2),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_computedOutputs(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "3"),
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
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "3"),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicecatalog.ResourceProvisionedProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_errorOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_basic(rName, "10.1.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
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
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
				),
			},
		},
	})
}

func TestAccServiceCatalogProvisionedProduct_productTagUpdateAfterError(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var pprod awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedProductConfig_productTagUpdateAfterError_valid(rName, bucketName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &pprod),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.version", "1.5"),
					acctest.S3BucketHasTag(ctx, bucketName, names.AttrVersion, "1.5"),
				),
			},
		},
	})
}

// Validates that a provisioned product in tainted status properly triggers an update
// on subsequent applies.
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/42585
func TestAccServiceCatalogProvisionedProduct_retryTaintedUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_provisioned_product.test"
	artifactsDataSourceName := "data.aws_servicecatalog_provisioning_artifacts.product_artifacts"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	initialArtifactID := "provisioning_artifact_details.0.id"
	newArtifactID := "provisioning_artifact_details.1.id"
	var v awstypes.ProvisionedProductDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1 - Setup
			{
				Config: testAccProvisionedProductConfig_retryTaintedUpdate(rName, false, false, "original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_id", artifactsDataSourceName, initialArtifactID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("AVAILABLE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioning_parameters"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("FailureSimulation"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact(acctest.CtFalse),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("ExtraParam"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact("original"),
						}),
					})),
				},
			},
			// Step 2 - Trigger a failure, leaving the provisioned product tainted
			{
				Config:      testAccProvisionedProductConfig_retryTaintedUpdate(rName, true, true, "updated"),
				ExpectError: regexache.MustCompile(`The following resource\(s\) failed to update:`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioning_parameters"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrKey:        knownvalue.StringExact("FailureSimulation"),
								"use_previous_value": knownvalue.Bool(false),
								names.AttrValue:      knownvalue.StringExact(acctest.CtTrue),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrKey:        knownvalue.StringExact("ExtraParam"),
								"use_previous_value": knownvalue.Bool(false),
								names.AttrValue:      knownvalue.StringExact("updated"),
							}),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("TAINTED")),
					// Verify state is rolled back to the parameters from the original setup run
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioning_parameters"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("FailureSimulation"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact(acctest.CtFalse),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("ExtraParam"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact("original"),
						}),
					})),
				},
			},
			// Step 3 - Verify an update is planned, even without configuration changes
			{
				Config: testAccProvisionedProductConfig_retryTaintedUpdate(rName, true, true, "updated"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ExpectError: regexache.MustCompile(`The following resource\(s\) failed to update:`),
			},
			// Step 4 - Resolve the failure, verifying an update is completed
			{
				Config: testAccProvisionedProductConfig_retryTaintedUpdate(rName, true, false, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedProductExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_id", artifactsDataSourceName, newArtifactID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("AVAILABLE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioning_parameters"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("FailureSimulation"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact(acctest.CtFalse),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:        knownvalue.StringExact("ExtraParam"),
							"use_previous_value": knownvalue.Bool(false),
							names.AttrValue:      knownvalue.StringExact("updated"),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckProvisionedProductDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_provisioned_product" {
				continue
			}

			_, err := tfservicecatalog.FindProvisionedProductByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["accept_language"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Catalog Provisioned Product %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisionedProductExists(ctx context.Context, t *testing.T, n string, v *awstypes.ProvisionedProductDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		output, err := tfservicecatalog.FindProvisionedProductByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["accept_language"])

		if err != nil {
			return err
		}

		*v = *output.ProvisionedProductDetail

		return nil
	}
}

// testAccCheckProvisionedProductProvisioningArtifactIDChanged verifies that the provisioned artifact
// ID differs between two provisioned products. If either provisioned product details or the provisioned
// artifact ID are null, the check will fail.
func testAccCheckProvisionedProductProvisioningArtifactIDChanged(pprod1, pprod2 *awstypes.ProvisionedProductDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pprod1 == nil || pprod2 == nil ||
			pprod1.ProvisioningArtifactId == nil ||
			pprod2.ProvisioningArtifactId == nil {
			return fmt.Errorf("provisioned product provisioning artifact ID is nil")
		}
		if aws.ToString(pprod1.ProvisioningArtifactId) == aws.ToString(pprod2.ProvisioningArtifactId) {
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
    regions                 = [data.aws_region.current.region]
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

func testAccProvisionedProductConfig_retryTaintedUpdate(rName string, useNewVersion bool, simulateFailure bool, extraParam string) string {
	return acctest.ConfigCompose(
		testAccProvisionedProductPortfolioBaseConfig(rName),
		fmt.Sprintf(`
locals {
  initial_provisioning_artifact = data.aws_servicecatalog_provisioning_artifacts.product_artifacts.provisioning_artifact_details[0]
  new_provisioning_artifact     = data.aws_servicecatalog_provisioning_artifacts.product_artifacts.provisioning_artifact_details[1]
}

resource "aws_servicecatalog_provisioned_product" "test" {
  name                     = %[1]q
  product_id               = aws_servicecatalog_product.test.id
  provisioning_artifact_id = %[2]t ? local.new_provisioning_artifact.id : local.initial_provisioning_artifact.id

  provisioning_parameters {
    key   = "FailureSimulation"
    value = "%[3]t"
  }

  provisioning_parameters {
    key   = "ExtraParam"
    value = %[4]q
  }
}

resource "aws_servicecatalog_product" "test" {
  description = %[1]q
  name        = %[1]q
  owner       = "test"
  type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    name         = "%[1]s - Initial"
    description  = "Initial"
    template_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type         = "CLOUD_FORMATION_TEMPLATE"
  }
}

resource "aws_servicecatalog_provisioning_artifact" "new_version" {
  product_id = aws_servicecatalog_product.test.id

  name         = "%[1]s - New"
  description  = "New"
  template_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  type         = "CLOUD_FORMATION_TEMPLATE"
}

data "aws_servicecatalog_provisioning_artifacts" "product_artifacts" {
  product_id = aws_servicecatalog_product.test.id

  depends_on = [aws_servicecatalog_provisioning_artifact.new_version]
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "product_template.yaml"

  source = "${path.module}/testdata/retry-tainted-update/product_template.yaml"
}
`, rName, useNewVersion, simulateFailure, extraParam))
}
