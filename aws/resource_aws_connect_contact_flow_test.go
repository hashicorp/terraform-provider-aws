package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccAwsConnectContactFlow_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccAwsConnectContactFlow_basic,
		"filename":   testAccAwsConnectContactFlow_filename,
		"disappears": testAccAwsConnectContactFlow_disappears_ConnectInstance,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsConnectContactFlow_basic(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rName, rName2, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rName, rName2, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func testAccAwsConnectContactFlow_filename(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfig_filename(rName, rName2, "Created", "testdata/service/connect/connect_contact_flow.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
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
				Config: testAccAwsConnectContactFlowConfig_filename(rName, rName2, "Updated", "testdata/service/connect/connect_contact_flow_updated.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

// Can't delete an contact flow. Test deletion of entire connect instance
func testAccAwsConnectContactFlow_disappears_ConnectInstance(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	// var v2 connect.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"
	instanceResourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsConnectInstance(), instanceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsConnectContactFlowExists(resourceName string, function *connect.DescribeContactFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Contact Flow not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Contact Flow ID not set")
		}
		instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*AWSClient).connectconn

		params := &connect.DescribeContactFlowInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		getFunction, err := conn.DescribeContactFlow(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckAwsConnectContactFlowDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_contact_flow" {
			continue
		}

		conn := acctest.Provider.Meta().(*AWSClient).connectconn

		instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeContactFlowInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		_, experr := conn.DescribeContactFlow(params)
		// Verify the error is what we want
		if experr != nil {
			if awsErr, ok := experr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return experr
		}
	}
	return nil
}

func testAccAwsConnectContactFlowConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccAwsConnectContactFlowConfigBasic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccAwsConnectContactFlowConfigBase(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  type        = "CONTACT_FLOW"
  content     = <<JSON
    {
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [ 
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Type": "MessageParticipant",
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Parameters": {
					"Text": %[2]q
				}
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Transitions": {},
				"Parameters": {}
			}
		]
    }
    JSON
  tags = {
    "Name"   = "Test Contact Flow",
    "Method" = %[2]q
  }
}
`, rName2, label))
}

func testAccAwsConnectContactFlowConfig_filename(rName, rName2 string, label string, filepath string) string {
	return acctest.ConfigCompose(
		testAccAwsConnectContactFlowConfigBase(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  description  = %[2]q
  type         = "CONTACT_FLOW"
  filename     = %[3]q
  content_hash = filebase64sha256(%[3]q)
  tags = {
    "Name"   = "Test Contact Flow",
    "Method" = %[2]q
  }
}
`, rName2, label, filepath))
}
