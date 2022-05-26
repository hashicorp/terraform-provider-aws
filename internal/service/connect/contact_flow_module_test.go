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

// Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectContactFlowModule_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccContactFlowModule_basic,
		"filename":   testAccContactFlowModule_filename,
		"disappears": testAccContactFlowModule_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccContactFlowModule_basic(t *testing.T) {
	var v connect.DescribeContactFlowModuleOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContactFlowModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Created"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowModuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func testAccContactFlowModule_filename(t *testing.T) {
	var v connect.DescribeContactFlowModuleOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContactFlowModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_filename(rName, rName2, "Created", "test-fixtures/connect_contact_flow_module.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Created"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content_hash",
					"filename",
				},
			},
			{
				Config: testAccContactFlowModuleConfig_filename(rName, rName2, "Updated", "test-fixtures/connect_contact_flow_module_updated.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactFlowModuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func testAccContactFlowModule_disappears(t *testing.T) {
	var v connect.DescribeContactFlowModuleOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContactFlowModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactFlowModuleExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceContactFlowModule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactFlowModuleExists(resourceName string, function *connect.DescribeContactFlowModuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Contact Flow Module not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Contact Flow Module ID not set")
		}
		instanceID, contactFlowModuleID, err := tfconnect.ContactFlowModuleParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeContactFlowModuleInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			InstanceId:          aws.String(instanceID),
		}

		getFunction, err := conn.DescribeContactFlowModule(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckContactFlowModuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_contact_flow_module" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, contactFlowModuleID, err := tfconnect.ContactFlowModuleParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeContactFlowModuleInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			InstanceId:          aws.String(instanceID),
		}

		_, err = conn.DescribeContactFlowModule(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccContactFlowModuleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccContactFlowModuleConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  content = <<JSON
    {
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Parameters": {
					"Text": "Hello contact flow module"
				},
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Type": "MessageParticipant"
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Parameters": {},
				"Transitions": {}
			}
		],
		"Settings": {
			"InputParameters": [],
			"OutputParameters": [],
			"Transitions": [
				{
					"DisplayName": "Success",
					"ReferenceName": "Success",
					"Description": ""
				},
				{
					"DisplayName": "Error",
					"ReferenceName": "Error",
					"Description": ""
				}
			]
		}
	}
    JSON

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label))
}

func testAccContactFlowModuleConfig_filename(rName, rName2 string, label string, filepath string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  description  = %[2]q
  filename     = %[3]q
  content_hash = filebase64sha256(%[3]q)

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label, filepath))
}
