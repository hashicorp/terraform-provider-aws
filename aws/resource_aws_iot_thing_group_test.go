package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIotThingGroup_basic(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroupConfig_basic(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "properties.description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccAWSIotThingGroup_full(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroupConfig_full(thingGroupName, "42", "this is my thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", fmt.Sprintf("%s_parent", thingGroupName)),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.Answer", "42"),
					resource.TestCheckResourceAttr(resourceName, "properties.description", "this is my thing group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Update attribute
				Config: testAccAWSIotThingGroupConfig_full(thingGroupName, "7", "this is my other thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", fmt.Sprintf("%s_parent", thingGroupName)),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.Answer", "7"),
					resource.TestCheckResourceAttr(resourceName, "properties.description", "this is my other thing group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{ // Remove thing group parent association
				Config: testAccAWSIotThingConfig_basic(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "properties.attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "properties.description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func testAccCheckIotThingGroupExists(n string, thing *iot.DescribeThingGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no IoT Thing Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		input := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeThingGroup(input)
		if err != nil {
			return err
		}

		*thing = *resp

		return nil
	}
}

func testAccCheckAWSIotThingGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group" {
			continue
		}

		input := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingGroup(input)
		if err != nil {
			if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("expected IoT Thing Group to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccAWSIotThingGroupConfig_basic(thingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"
}
`, thingGroupName)
}

func testAccAWSIotThingGroupConfig_full(thingGroupName, answer, description string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "parent" {
  name = "%s_parent"
}

resource "aws_iot_thing_group" "test" {
  name = "%s"

  parent_group_name = "${aws_iot_thing_group.parent.name}"

  properties {
    attributes = {
      One    = "11111"
      Two    = "TwoTwo"
      Answer = "%s"
    }
    description = "%s"
  }

  tags {
    "tagKey" = "tagVal"
  }
}
`, thingGroupName, thingGroupName, answer, description)
}
