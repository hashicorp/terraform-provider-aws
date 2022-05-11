package events_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
)

func TestAccEventsBus_basic(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	busNameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, "name", busName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v2),
					testAccCheckBusRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busNameModified)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, "name", busNameModified),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccBusConfig_Tags1(busNameModified, "key", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccEventsBus_tags(t *testing.T) {
	var v1, v2, v3, v4 eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_Tags1(busName, "key1", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_Tags2(busName, "key1", "updated", "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v2),
					testAccCheckBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccBusConfig_Tags1(busName, "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v4),
					testAccCheckBusNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccEventsBus_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBusConfig("default"),
				ExpectError: regexp.MustCompile(`cannot be 'default'`),
			},
		},
	})
}

func TestAccEventsBus_disappears(t *testing.T) {
	var v eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfevents.ResourceBus(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventsBus_partnerEventSource(t *testing.T) {
	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var busOutput eventbridge.DescribeEventBusOutput
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusPartnerEventSourceConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(resourceName, &busOutput),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "event_source_name", busName),
					resource.TestCheckResourceAttr(resourceName, "name", busName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckBusDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_bus" {
			continue
		}

		params := eventbridge.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventBus(&params)

		if err == nil {
			return fmt.Errorf("EventBridge event bus (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckBusExists(n string, v *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn
		params := eventbridge.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEventBus(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("EventBridge event bus (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckBusRecreated(i, j *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) == aws.StringValue(j.Arn) {
			return fmt.Errorf("EventBridge event bus not recreated")
		}
		return nil
	}
}

func testAccCheckBusNotRecreated(i, j *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) != aws.StringValue(j.Arn) {
			return fmt.Errorf("EventBridge event bus was recreated")
		}
		return nil
	}
}

func testAccBusConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}
`, name)
}

func testAccBusConfig_Tags1(name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, key, value)
}

func testAccBusConfig_Tags2(name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, key1, value1, key2, value2)
}

func testAccBusPartnerEventSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name              = %[1]q
  event_source_name = %[1]q
}
`, name)
}
