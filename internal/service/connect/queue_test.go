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
func TestAccConnectQueue_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                         testAccQueue_basic,
		"disappears":                    testAccQueue_disappears,
		"update_hours_of_operation_id":  testAccQueue_updateHoursOfOperationId,
		"update_max_contacts":           testAccQueue_updateMaxContacts,
		"update_outbound_caller_config": testAccQueue_updateOutboundCallerConfig,
		"update_status":                 testAccQueue_updateStatus,
		"update_quick_connect_ids":      testAccQueue_updateQuickConnectIds,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccQueue_basic(t *testing.T) {
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalDescription := "Created"
	updatedDescription := "Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_basic(rName, rName2, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQueue_disappears(t *testing.T) {
	t.Skip("Queues do not support deletion today")
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccQueue_updateHoursOfOperationId(t *testing.T) {
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQueue_updateMaxContacts(t *testing.T) {
	t.Skip("A bug in the service API has been reported")
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalMaxContacts := "1"
	updatedMaxContacts := "2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_maxContacts(rName, rName2, originalMaxContacts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "max_contacts", originalMaxContacts),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_maxContacts(rName, rName2, updatedMaxContacts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "max_contacts", updatedMaxContacts),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQueue_updateOutboundCallerConfig(t *testing.T) {
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalOutboundCallerIdName := "original"
	updatedOutboundCallerIdName := "updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_outboundCaller(rName, rName2, originalOutboundCallerIdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.0.outbound_caller_id_name", originalOutboundCallerIdName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_outboundCaller(rName, rName2, updatedOutboundCallerIdName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.0.outbound_caller_id_name", updatedOutboundCallerIdName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQueue_updateStatus(t *testing.T) {
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalStatus := connect.QueueStatusEnabled
	updatedStatus := connect.QueueStatusDisabled

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_status(rName, rName2, originalStatus),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", originalStatus),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_status(rName, rName2, updatedStatus),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", updatedStatus),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccQueue_updateQuickConnectIds(t *testing.T) {
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	description := "test queue integrations with quick connects"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				// start with no quick connects associated with the queue
				Config: testAccQueueConfig_basic(rName, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// associate one quick connect to the queue
				Config: testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "quick_connect_ids.0", "aws_connect_quick_connect.test1", "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// associate two quick connects to the queue
				Config: testAccQueueConfig_quickConnect2(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// remove one quick connect
				Config: testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName4),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "quick_connect_ids.0", "aws_connect_quick_connect.test1", "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckQueueExists(resourceName string, function *connect.DescribeQueueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Queue not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Queue ID not set")
		}
		instanceID, queueID, err := tfconnect.QueueParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeQueueInput{
			QueueId:    aws.String(queueID),
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeQueue(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckQueueDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_queue" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, queueID, err := tfconnect.QueueParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeQueueInput{
			QueueId:    aws.String(queueID),
			InstanceId: aws.String(instanceID),
		}

		_, err = conn.DescribeQueue(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccQueueBaseConfig(rName string) string {
	// Use the aws_connect_hours_of_operation data source with the default "Basic Hours" that comes with connect instances.
	// Because if a resource is used, Terraform will not be able to delete it since queues do not have support for the delete api
	// yet but still references hours_of_operation_id. However, using the data source will result in the failure of the
	// disppears test (removed till delete api is available) for the connect instance (We test disappears on the Connect instance
	// instead of the queue since the queue does not support delete). The error is:
	// Step 1/1 error: Error running post-apply plan: exit status 1
	// Error: error finding Connect Hours of Operation Summary by name (Basic Hours): ResourceNotFoundException: Instance not found
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

data "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Basic Hours"
}
`, rName)
}

func testAccQueueConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, label))
}

func testAccQueueConfig_hoursOfOperation(rName, rName2, selectHoursOfOperationId string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
locals {
  select_hours_of_operation_id = %[2]q
}

resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = "Example aws_connect_hours_of_operation to test updates"
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }
}

resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update hours_of_operation_id"
  hours_of_operation_id = local.select_hours_of_operation_id == "first" ? data.aws_connect_hours_of_operation.test.hours_of_operation_id : aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, selectHoursOfOperationId))
}

//lint:ignore U1000 Ignore unused function temporarily
func testAccQueueConfig_maxContacts(rName, rName2, maxContacts string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update max contacts"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
  max_contacts          = %[2]q

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, maxContacts))
}

func testAccQueueConfig_outboundCaller(rName, rName2, OutboundCallerIdName string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update outbound caller config"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  outbound_caller_config {
    outbound_caller_id_name = %[2]q
  }

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, OutboundCallerIdName))
}

func testAccQueueConfig_status(rName, rName2, status string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update status"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
  status                = %[2]q

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, status))
}

func testAccQueueQuickConnectBaseConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_quick_connect" "test1" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = "Test Quick Connect 1"

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = "+12345678912"
    }
  }

  tags = {
    "Name" = "Test Quick Connect 1"
  }
}

resource "aws_connect_quick_connect" "test2" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Quick Connect 2"

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = "+12345678913"
    }
  }

  tags = {
    "Name" = "Test Quick Connect 2"
  }
}
`, rName, rName2)
}

func testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, label string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		testAccQueueQuickConnectBaseConfig(rName2, rName3),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  quick_connect_ids = [
    aws_connect_quick_connect.test1.quick_connect_id,
  ]

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName4, label))
}

func testAccQueueConfig_quickConnect2(rName, rName2, rName3, rName4, label string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		testAccQueueQuickConnectBaseConfig(rName2, rName3),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  quick_connect_ids = [
    aws_connect_quick_connect.test1.quick_connect_id,
    aws_connect_quick_connect.test2.quick_connect_id,
  ]

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName4, label))
}
