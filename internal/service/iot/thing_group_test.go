package iot_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTThingGroup_basic(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "iot", regexp.MustCompile(fmt.Sprintf("thinggroup/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.parent_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.root_to_parent_thing_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func TestAccIoTThingGroup_disappears(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceThingGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingGroup_tags(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
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
				Config: testAccThingGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccThingGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIoTThingGroup_parentGroup(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"
	parentResourceName := "aws_iot_thing_group.parent"
	grandparentResourceName := "aws_iot_thing_group.grandparent"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_parent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttrPair(resourceName, "parent_group_name", parentResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.parent_group_name", parentResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.root_to_parent_groups.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.0.group_arn", grandparentResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.0.group_name", grandparentResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.1.group_arn", parentResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.1.group_name", parentResourceName, "name"),
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

func TestAccIoTThingGroup_properties(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_properties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingGroupConfig_propertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 2"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

func testAccCheckThingGroupExists(n string, v *iot.DescribeThingGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		output, err := tfiot.FindThingGroupByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckThingGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group" {
			continue
		}

		_, err := tfiot.FindThingGroupByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IoT Thing Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccThingGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccThingGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccThingGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccThingGroupConfig_parent(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "grandparent" {
  name = "%[1]s-grandparent"
}

resource "aws_iot_thing_group" "parent" {
  name = "%[1]s-parent"

  parent_group_name = aws_iot_thing_group.grandparent.name
}

resource "aws_iot_thing_group" "test" {
  name = %[1]q

  parent_group_name = aws_iot_thing_group.parent.name
}
`, rName)
}

func testAccThingGroupConfig_properties(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  properties {
    attribute_payload {
      attributes = {
        Key1 = "Value1"
      }
    }

    description = "test description 1"
  }
}
`, rName)
}

func testAccThingGroupConfig_propertiesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  properties {
    attribute_payload {
      attributes = {
        Key2 = "Value2"
        Key3 = "Value3"
      }
    }

    description = "test description 2"
  }
}
`, rName)
}
