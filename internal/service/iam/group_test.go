package iam_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMGroup_basic(t *testing.T) {
	var conf iam.GetGroupOutput
	resourceName := "aws_iam_group.test"
	resourceName2 := "aws_iam_group.test2"
	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-basic-2-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName, "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroup2Config(groupName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName2, &conf),
					testAccCheckGroupAttributes(&conf, groupName2, "/funnypath/"),
				),
			},
		},
	})
}

func TestAccIAMGroup_nameChange(t *testing.T) {
	var conf iam.GetGroupOutput
	resourceName := "aws_iam_group.test"
	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-basic-2-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName, "/"),
				),
			},
			{
				Config: testAccGroupConfig(groupName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName2, "/"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group" {
			continue
		}

		// Try to get group
		_, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return errors.New("still exist.")
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckGroupExists(n string, res *iam.GetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Group name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckGroupAttributes(group *iam.GetGroupOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.Group.GroupName != name {
			return fmt.Errorf("Bad name: %s when %s was expected", *group.Group.GroupName, name)
		}

		if *group.Group.Path != path {
			return fmt.Errorf("Bad path: %s when %s was expected", *group.Group.Path, path)
		}

		return nil
	}
}

func testAccGroupConfig(groupName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "%s"
  path = "/"
}
`, groupName)
}

func testAccGroup2Config(groupName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test2" {
  name = "%s"
  path = "/funnypath/"
}
`, groupName)
}
