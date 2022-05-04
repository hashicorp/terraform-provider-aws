package iot_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTTopicRuleDestination_basic(t *testing.T) {
	url := fmt.Sprintf("https://%s.example.com/", acctest.RandomSubdomain())
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTopicRuleDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "iot", regexp.MustCompile(`ruledestination/http/.+`)),
					resource.TestCheckResourceAttr(resourceName, "http_url_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_url_configuration.0.confirmation_url", url),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
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

func TestAccIoTTopicRuleDestination_disappears(t *testing.T) {
	url := fmt.Sprintf("https://%s.example.com/", acctest.RandomSubdomain())
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTopicRuleDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceTopicRuleDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTopicRuleDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_topic_rule_destination" {
			continue
		}

		_, err := tfiot.FindTopicRuleDestinationByARN(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IoT Topic Rule Destination %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckTopicRuleDestinationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Topic Rule Destination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		_, err := tfiot.FindTopicRuleDestinationByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccTopicRuleDestinationConfig(url string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule_destination" "test" {
  http_url_configuration {
    confirmation_url = %[1]q
  }
}
`, url)
}
