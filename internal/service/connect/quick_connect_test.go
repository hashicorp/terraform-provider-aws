package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectQuickConnect_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccQuickConnect_phoneNumber,
		"disappears": testAccQuickConnect_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccQuickConnect_phoneNumber(t *testing.T) {
	var v connect.DescribeQuickConnectOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQuickConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Created", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// update description
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Updated", "+12345678912"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678912"),

					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// update phone number
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Updated", "+12345678913"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQuickConnectExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.quick_connect_type", "PHONE_NUMBER"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_config.0.phone_config.0.phone_number", "+12345678913"),

					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQuickConnect_disappears(t *testing.T) {
	var v connect.DescribeQuickConnectOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQuickConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectConfig_phoneNumber(rName, rName2, "Disappear", "+12345678912"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickConnectExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceQuickConnect(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickConnectExists(resourceName string, function *connect.DescribeQuickConnectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Quick Connect not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Quick Connect ID not set")
		}
		instanceID, quickConnectID, err := tfconnect.QuickConnectParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeQuickConnectInput{
			QuickConnectId: aws.String(quickConnectID),
			InstanceId:     aws.String(instanceID),
		}

		getFunction, err := conn.DescribeQuickConnect(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckQuickConnectDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_quick_connect" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, quickConnectID, err := tfconnect.QuickConnectParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeQuickConnectInput{
			QuickConnectId: aws.String(quickConnectID),
			InstanceId:     aws.String(instanceID),
		}

		_, err = conn.DescribeQuickConnect(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccQuickConnectBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccQuickConnectConfig_phoneNumber(rName, rName2, label string, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = %[3]q
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
  }
}
`, rName2, label, phoneNumber))
}
