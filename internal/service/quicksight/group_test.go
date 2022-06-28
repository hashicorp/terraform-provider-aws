package quicksight_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightGroup_basic(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("group/default/%s", rName1)),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("group/default/%s", rName2)),
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

func TestAccQuickSightGroup_withDescription(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_description(rName, "Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 1"),
				),
			},
			{
				Config: testAccGroupConfig_description(rName, "Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 2"),
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

func TestAccQuickSightGroup_disappears(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupDisappears(&group),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupExists(resourceName string, group *quicksight.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, groupName, err := tfquicksight.GroupParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		input := &quicksight.DescribeGroupInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		}

		output, err := conn.DescribeGroup(input)

		if err != nil {
			return err
		}

		if output == nil || output.Group == nil {
			return fmt.Errorf("QuickSight Group (%s) not found", rs.Primary.ID)
		}

		*group = *output.Group

		return nil
	}
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_group" {
			continue
		}

		awsAccountID, namespace, groupName, err := tfquicksight.GroupParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeGroup(&quicksight.DescribeGroupInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		})
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("QuickSight Group '%s' was not deleted properly", rs.Primary.ID)
	}

	return nil
}

func testAccCheckGroupDisappears(v *quicksight.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		arn, err := arn.Parse(aws.StringValue(v.Arn))
		if err != nil {
			return err
		}

		parts := strings.SplitN(arn.Resource, "/", 3)

		input := &quicksight.DeleteGroupInput{
			AwsAccountId: aws.String(arn.AccountID),
			Namespace:    aws.String(parts[1]),
			GroupName:    v.GroupName,
		}

		if _, err := conn.DeleteGroup(input); err != nil {
			return err
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_group" "default" {
  group_name = %[1]q
}
`, rName)
}

func testAccGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "default" {
  aws_account_id = data.aws_caller_identity.current.account_id
  group_name     = %[1]q
  description    = %[2]q
}
`, rName, description)
}
