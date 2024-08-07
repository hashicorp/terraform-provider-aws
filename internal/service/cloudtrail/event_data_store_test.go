// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudTrailEventDataStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"equals.#":      acctest.Ct1,
						"equals.0":      "Management",
						names.AttrField: "eventCategory",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "Default management events"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudtrail", regexache.MustCompile(`eventdatastore/.+`)),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", "EXTENDABLE_RETENTION_PRICING"),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "2555"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", acctest.CtFalse),
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

func TestAccCloudTrailEventDataStore_billingMode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_billingMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", "FIXED_RETENTION_PRICING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventDataStoreConfig_billingModeUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", "EXTENDABLE_RETENTION_PRICING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCloudTrailEventDataStore_kmsKeyId(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_kmsKeyId(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudtrail", regexache.MustCompile(`eventdatastore/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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

func TestAccCloudTrailEventDataStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudtrail.ResourceEventDataStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudTrailEventDataStore_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
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
				Config: testAccEventDataStoreConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEventDataStoreConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCloudTrailEventDataStore_options(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_options(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "365"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventDataStoreConfig_optionsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrRetentionPeriod, "90"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCloudTrailEventDataStore_advancedEventSelector(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDataStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_advancedSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "s3Custom"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", acctest.Ct4),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "eventName",
						"equals.#":      acctest.Ct1,
						"equals.0":      "DeleteObject",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::S3::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.name", "lambdaLogAllEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.field_selector.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::Lambda::Function",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.2.name", "dynamoDbReadOnlyEvents"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::DynamoDB::Table",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.name", "s3OutpostsWriteOnlyEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.field_selector.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::S3Outposts::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.name", "managementEventsSelector"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.field_selector.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.4.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Management",
					}),
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

func testAccCheckEventDataStoreExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)

		_, err := tfcloudtrail.FindEventDataStoreByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckEventDataStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudtrail_event_data_store" {
				continue
			}

			_, err := tfcloudtrail.FindEventDataStoreByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudTrail Event Data Store %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEventDataStoreConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  termination_protection_enabled = false # For ease of deletion.
}
`, rName)
}

func testAccEventDataStoreConfig_billingMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  billing_mode                   = "FIXED_RETENTION_PRICING"
  termination_protection_enabled = false # For ease of deletion.
}
`, rName)
}

func testAccEventDataStoreConfig_billingModeUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  billing_mode                   = "EXTENDABLE_RETENTION_PRICING"
  termination_protection_enabled = false # For ease of deletion.
}
`, rName)
}

func testAccEventDataStoreConfig_kmsKeyId(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  multi_region = true
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Sid    = "Enable IAM User Permissions"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudtrail_event_data_store" "test" {
  name                           = %[1]q
  kms_key_id                     = aws_kms_key.test.arn
  termination_protection_enabled = false # For ease of deletion.
}
`, rName)
}

func testAccEventDataStoreConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  termination_protection_enabled = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEventDataStoreConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  termination_protection_enabled = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccEventDataStoreConfig_options(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name                 = %[1]q
  multi_region_enabled = false
  organization_enabled = true
  retention_period     = 365

  termination_protection_enabled = true
}
`, rName)
}

func testAccEventDataStoreConfig_optionsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name                 = %[1]q
  multi_region_enabled = true
  organization_enabled = false
  retention_period     = 90

  termination_protection_enabled = false
}
`, rName)
}

func testAccEventDataStoreConfig_advancedSelector(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  termination_protection_enabled = false

  advanced_event_selector {
    name = "s3Custom"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "eventName"
      equals = ["DeleteObject"]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }

  advanced_event_selector {
    name = "lambdaLogAllEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::Lambda::Function"]
    }
  }

  advanced_event_selector {
    name = "dynamoDbReadOnlyEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "readOnly"
      equals = ["true"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::DynamoDB::Table"]
    }
  }

  advanced_event_selector {
    name = "s3OutpostsWriteOnlyEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3Outposts::Object"]
    }
  }

  advanced_event_selector {
    name = "managementEventsSelector"
    field_selector {
      field  = "eventCategory"
      equals = ["Management"]
    }
  }
}
`, rName)
}
