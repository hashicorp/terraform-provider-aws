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

func testAccPhoneNumber_basic(t *testing.T) {
	var v connect.DescribePhoneNumberOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_phone_number.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttrSet(resourceName, "phone_number"),
					resource.TestCheckResourceAttr(resourceName, "status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "status.0.status", connect.PhoneNumberWorkflowStatusClaimed),
					resource.TestCheckResourceAttrPair(resourceName, "target_arn", "aws_connect_instance.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.PhoneNumberTypeDid),
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

func testAccPhoneNumber_disappears(t *testing.T) {
	var v connect.DescribePhoneNumberOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_phone_number.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourcePhoneNumber(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPhoneNumberExists(resourceName string, function *connect.DescribePhoneNumberOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Phone Number not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Phone Number ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribePhoneNumberInput{
			PhoneNumberId: aws.String(rs.Primary.ID),
		}

		getFunction, err := conn.DescribePhoneNumber(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckPhoneNumberDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_phone_number" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribePhoneNumberInput{
			PhoneNumberId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribePhoneNumber(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccPhoneNumberConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccPhoneNumberConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPhoneNumberConfig_base(rName),
		`
resource "aws_connect_phone_number" "test" {
  target_arn   = aws_connect_instance.test.arn
  country_code = "US"
  type         = "DID"
}
`)
}
