package cloudtrail_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEventDataStore_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDataStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"equals.#": "1",
						"equals.0": "Management",
						"field":    "eventCategory",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "Default management events"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudtrail", regexp.MustCompile(`eventdatastore/.+`)),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "2555"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", "false"),
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

func TestAccEventDataStore_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDataStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudtrail.ResourceEventDataStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventDataStore_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDataStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
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
				Config: testAccEventDataStoreTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEventDataStoreTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEventDataStore_options(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationManagementAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDataStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "365"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventDataStoreOptionsUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDataStoreExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_region_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "organization_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "retention_period", "90"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection_enabled", "false"),
				),
			},
		},
	})
}

func TestAccEventDataStore_advancedEventSelector(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail_event_data_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEventDataStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDataStoreConfig_advancedEventSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "s3Custom"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "eventName",
						"equals.#": "1",
						"equals.0": "DeleteObject",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::S3::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.name", "lambdaLogAllEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.field_selector.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::Lambda::Function",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.2.name", "dynamoDbReadOnlyEvents"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::DynamoDB::Table",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.name", "s3OutpostsWriteOnlyEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.field_selector.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::S3Outposts::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.name", "managementEventsSelector"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.field_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.4.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Management",
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

func testAccCheckEventDataStoreExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudTrail Event Data Store ARN is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn

		_, err := tfcloudtrail.FindEventDataStoreByARN(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckEventDataStoreDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudtrail_event_data_store" {
			continue
		}

		_, err := tfcloudtrail.FindEventDataStoreByARN(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		return fmt.Errorf("CloudTrail Event Data Store %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccEventDataStoreConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail_event_data_store" "test" {
  name = %[1]q

  termination_protection_enabled = false # For ease of deletion.
}
`, rName)
}

func testAccEventDataStoreTagsConfig1(rName, tagKey1, tagValue1 string) string {
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

func testAccEventDataStoreTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccEventDataStoreOptionsConfig(rName string) string {
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

func testAccEventDataStoreOptionsUpdatedConfig(rName string) string {
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

func testAccEventDataStoreConfig_advancedEventSelector(rName string) string {
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
