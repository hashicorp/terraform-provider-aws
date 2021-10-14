package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tagresource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSAutoscalingGroupTag_basic(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1"),
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

func TestAccAWSAutoscalingGroupTag_disappears(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAutoscalingGroupTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAutoscalingGroupTag_Value(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckAutoscalingGroupTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group_tag" {
			continue
		}

		identifier, key, err := tagresource.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = keyvaluetags.AutoscalingGetTag(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, key)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("AutoScaling Group (%s) tag (%s) still exists", identifier, key)
	}

	return nil
}

func testAccCheckAutoscalingGroupTagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", n)
		}

		identifier, key, err := tagresource.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		_, err = keyvaluetags.AutoscalingGetTag(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, key)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAutoscalingGroupTagConfig(key string, value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name_prefix   = "terraform-test-"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.nano"
}

resource "aws_autoscaling_group" "test" {
  lifecycle {
    ignore_changes = [tag]
  }

  availability_zones = [data.aws_availability_zones.available.names[0]]

  min_size = 0
  max_size = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}

resource "aws_autoscaling_group_tag" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.name

  tag {
    key   = %[1]q
    value = %[2]q

    propagate_at_launch = true
  }
}
`, key, value))
}
