package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsWorkspacesWorkspace_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesWorkspaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists("aws_workspaces_workspace.test"),
				),
			},
		},
	})
}

func TestAccAwsWorkspacesWorkspace_import(t *testing.T) {
	resourceName := "aws_workspaces_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccWorkspacesWorkspaceConfig(acctest.RandString(5)),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsWorkspacesWorkspaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).workspacesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_workspace" {
			continue
		}

		input := &workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeWorkspaces(input)
		if err != nil {
			if isAWSErr(err, workspaces.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		ws := resp.Workspaces[0]
		if *ws.State != workspaces.WorkspaceStateTerminating ||
			*ws.State != workspaces.WorkspaceStateTerminated {
			return fmt.Errorf("Workspace (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesWorkspaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn

		input := &workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []*string{aws.String(rs.Primary.ID)},
		}

		_, err := conn.DescribeWorkspaces(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccWorkspacesWorkspaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test2" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
}

resource "aws_directory_service_directory" "test" {
  name = "tf-test-%s.ikinari-steak.net"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test1.id}","${aws_subnet.test2.id}"]
  }
}

data "aws_workspaces_bundle" "test" {
  bundle_id = "wsb-b0s22j3d7"
}

resource "aws_workspaces_workspace" "test" {
  bundle_id = "${data.aws_workspaces_bundle.test.bundle_id}"
  directory_id = "${aws_directory_service_directory.test.id}"
  user_name = "tf-test-%s"
}
`, rName, rName)
}
