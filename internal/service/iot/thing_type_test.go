package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIoTThingType_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists("aws_iot_thing_type.foo"),
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "name", fmt.Sprintf("tf_acc_iot_thing_type_%d", rInt)),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "tags.%", "0"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIoTThingType_full(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists("aws_iot_thing_type.foo"),
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.description", "MyDescription"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.searchable_attributes.#", "3"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "true"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "tags_all.%", "1"),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingTypeConfig_fullUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "false"),
				),
			},
		},
	})
}

func TestAccIoTThingType_tags(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resourceName := "aws_iot_thing_type.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_tags1(rName, "key1", "user@example"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists("aws_iot_thing_type.foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "user@example"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingTypeConfig_tags2(rName, "key1", "user@example", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists("aws_iot_thing_type.foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "user@example"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccThingTypeConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists("aws_iot_thing_type.foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckThingTypeExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
		input := &iot.ListThingTypesInput{}

		output, err := conn.ListThingTypes(input)

		if err != nil {
			return err
		}

		for _, rule := range output.ThingTypes {
			if aws.StringValue(rule.ThingTypeName) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("IoT Topic Rule (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckThingTypeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_type" {
			continue
		}

		params := &iot.DescribeThingTypeInput{
			ThingTypeName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingType(params)
		if err == nil {
			return fmt.Errorf("Expected IoT Thing Type to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccThingTypeConfig_basic(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name = "tf_acc_iot_thing_type_%d"
}
`, rName)
}

func testAccThingTypeConfig_full(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = true

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }

  tags = {
    testtag = "MyTagValue"
  }
}
`, rName)
}

func testAccThingTypeConfig_fullUpdated(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = false

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }

  tags = {
    testtag = "MyTagValue"
  }
}
`, rName)
}

func testAccThingTypeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%[1]s"
  deprecated = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccThingTypeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%[1]s"
  deprecated = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
