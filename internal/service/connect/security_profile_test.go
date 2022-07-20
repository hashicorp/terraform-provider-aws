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
func TestAccConnectSecurityProfile_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":              testAccSecurityProfile_basic,
		"disappears":         testAccSecurityProfile_disappears,
		"update_permissions": testAccSecurityProfile_updatePermissions,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccSecurityProfile_basic(t *testing.T) {
	var v connect.DescribeSecurityProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", "Created"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccSecurityProfile_updatePermissions(t *testing.T) {
	var v connect.DescribeSecurityProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "TestPermissionsUpdate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test updating permissions
				Config: testAccSecurityProfileConfig_permissions(rName, rName2, "TestPermissionsUpdate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_profile_id"),

					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", "TestPermissionsUpdate"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccSecurityProfile_disappears(t *testing.T) {
	var v connect.DescribeSecurityProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityProfileExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceSecurityProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityProfileExists(resourceName string, function *connect.DescribeSecurityProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Security Profile not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Security Profile ID not set")
		}
		instanceID, securityProfileID, err := tfconnect.SecurityProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeSecurityProfileInput{
			InstanceId:        aws.String(instanceID),
			SecurityProfileId: aws.String(securityProfileID),
		}

		getFunction, err := conn.DescribeSecurityProfile(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckSecurityProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_security_profile" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, securityProfileID, err := tfconnect.SecurityProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeSecurityProfileInput{
			InstanceId:        aws.String(instanceID),
			SecurityProfileId: aws.String(securityProfileID),
		}

		_, err = conn.DescribeSecurityProfile(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccSecurityProfileBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccSecurityProfileConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName2, label))
}

func testAccSecurityProfileConfig_permissions(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  permissions = [
    "BasicAgentAccess",
    "OutboundCallAccess",
  ]

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName2, label))
}
