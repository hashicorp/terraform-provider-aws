package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEc2Tag_basic(t *testing.T) {
	var tag ec2.TagDescription
	rBgpAsn := randIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName, &tag),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
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

func TestAccAWSEc2Tag_disappears(t *testing.T) {
	var tag ec2.TagDescription
	rBgpAsn := randIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName, &tag),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2Tag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2Tag_Value(t *testing.T) {
	var tag ec2.TagDescription
	rBgpAsn := randIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName, &tag),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName, &tag),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckEc2TagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_tag" {
			continue
		}

		resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []*string{aws.String(resourceID)},
				},
				{
					Name:   aws.String("key"),
					Values: []*string{aws.String(key)},
				},
			},
		}

		output, err := conn.DescribeTags(input)

		if err != nil {
			return err
		}

		var tag *ec2.TagDescription

		for _, outputTag := range output.Tags {
			if aws.StringValue(outputTag.Key) == key {
				tag = outputTag
				break
			}
		}

		if tag != nil {
			return fmt.Errorf("Tag (%s) for resource (%s) still exists", key, resourceID)
		}
	}

	return nil
}

func testAccCheckEc2TagExists(n string, tag *ec2.TagDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []*string{aws.String(resourceID)},
				},
				{
					Name:   aws.String("key"),
					Values: []*string{aws.String(key)},
				},
			},
		}

		output, err := conn.DescribeTags(input)

		if err != nil {
			return err
		}

		for _, outputTag := range output.Tags {
			if aws.StringValue(outputTag.Key) == key {
				*tag = *outputTag
				break
			}
		}

		if tag == nil {
			return fmt.Errorf("Tag (%s) for resource (%s) not found", key, resourceID)
		}

		return nil
	}
}

func testAccEc2TagConfig(rBgpAsn int, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_vpn_connection.test.transit_gateway_attachment_id
  key         = %[2]q
  value       = %[3]q
}
`, rBgpAsn, key, value)
}
