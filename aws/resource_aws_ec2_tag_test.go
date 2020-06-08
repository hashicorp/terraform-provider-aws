package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEc2ResourceTag_basic(t *testing.T) {
	var tag ec2.TagDescription

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigIgnoreTagsKeys1("Name") + testAccEc2ResourceTagConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2ResourceTagExists(
						"aws_ec2_tag.test", &tag),
					resource.TestCheckResourceAttr("aws_ec2_tag.test", "key", "Name"),
					resource.TestCheckResourceAttr("aws_ec2_tag.test", "value", "Hello World"),
				),
			},
		},
	})
}

func TestAccAWSEc2ResourceTag_OutOfBandDelete(t *testing.T) {
	var tag ec2.TagDescription

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ResourceTagConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2ResourceTagExists("aws_ec2_tag.test", &tag),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2Tag(), "aws_ec2_tag.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEc2ResourceTagExists(n string, tag *ec2.TagDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		id, key, err := extractResourceIDAndKeyFromEc2TagID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error getting ID or key from EC2 tag ID: %s", rs.Primary.ID)
		}
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeTags(&ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []*string{aws.String(id)},
				},
				{
					Name:   aws.String("key"),
					Values: []*string{aws.String(key)},
				},
			},
		})

		if err != nil {
			return err
		}

		if len(resp.Tags) == 0 {
			return fmt.Errorf("No tags found")
		}

		*tag = *resp.Tags[0]
		//		return fmt.Errorf("Tag found %s => %s", aws.StringValue(tag.Key), aws.StringValue(tag.Value))

		return nil
	}
}

const testAccEc2ResourceTagConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_ec2_tag" "test" {
  resource_id = "${aws_vpc.test.id}"
  key         = "Name"
  value       = "Hello World"
}
`
