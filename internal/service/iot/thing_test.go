package iot_test

import (
	"fmt"
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

func TestAccIoTThing_basic(t *testing.T) {
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
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

func TestAccIoTThing_full(t *testing.T) {
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	typeName := fmt.Sprintf("tf_acc_type_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingConfig_full(thingName, typeName, "42"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "42"),
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
				Config: testAccThingConfig_full(thingName, typeName, "differentOne"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "differentOne"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{ // Remove thing type association
				Config: testAccThingConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func testAccCheckThingExists(n string, v *iot.DescribeThingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		output, err := tfiot.FindThingByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckThingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing" {
			continue
		}

		_, err := tfiot.FindThingByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IoT Thing %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccThingConfig_basic(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"
}
`, thingName)
}

func testAccThingConfig_full(thingName, typeName, answer string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"

  attributes = {
    One    = "11111"
    Two    = "TwoTwo"
    Answer = "%s"
  }

  thing_type_name = aws_iot_thing_type.test.name
}

resource "aws_iot_thing_type" "test" {
  name = "%s"
}
`, thingName, answer, typeName)
}
