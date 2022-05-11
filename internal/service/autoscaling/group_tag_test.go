package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingGroupTag_basic(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(resourceName),
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

func TestAccAutoScalingGroupTag_disappears(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscaling.ResourceGroupTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingGroupTag_value(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(resourceName),
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
				Config: testAccGroupTagConfig_basic("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckGroupTagDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group_tag" {
			continue
		}

		identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfautoscaling.GetTag(conn, identifier, tfautoscaling.TagResourceTypeGroup, key)

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

func testAccCheckGroupTagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", n)
		}

		identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		_, err = tfautoscaling.GetTag(conn, identifier, tfautoscaling.TagResourceTypeGroup, key)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccGroupTagConfig_basic(key string, value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
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
