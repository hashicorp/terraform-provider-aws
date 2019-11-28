package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	testAccWorkspaceConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test-a" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-east-1a"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test-c" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-east-1c"
  cidr_block = "10.0.2.0/24"
}

resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test-a.id}","${aws_subnet.test-c.id}"]
  }
}

data aws_iam_policy_document workspaces {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  }
}

resource aws_iam_role workspaces-default {
  name               = "workspaces_DefaultRole"
  assume_role_policy = data.aws_iam_policy_document.workspaces.json
}

resource aws_iam_role_policy_attachment workspaces-default-service-access {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
}

resource aws_iam_role_policy_attachment workspaces-default-self-service-access {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
}

resource "aws_workspaces_directory" "test" {
  directory_id = "${aws_directory_service_directory.test.id}"
}
`

	testAccWorkspaceConfig_subnetIds = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test-a" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-east-1a"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test-c" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-east-1c"
  cidr_block = "10.0.2.0/24"
}

resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test-a.id}","${aws_subnet.test-c.id}"]
  }
}

data aws_iam_policy_document workspaces {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  }
}

resource aws_iam_role workspaces-default {
  name               = "workspaces_DefaultRole"
  assume_role_policy = data.aws_iam_policy_document.workspaces.json
}

resource aws_iam_role_policy_attachment workspaces-default-service-access {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
}

resource aws_iam_role_policy_attachment workspaces-default-self-service-access {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
}

resource "aws_workspaces_directory" "test" {
  directory_id = "${aws_directory_service_directory.test.id}"
  subnet_ids = ["${aws_subnet.test-a.id}","${aws_subnet.test-c.id}"]
}
`
)

func TestAccAwsWorkspacesDirectory_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists("aws_workspaces_directory.test"),
				),
			},
			{
				ResourceName:      "aws_workspaces_directory.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsWorkspacesDirectory_subnetIds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_subnetIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists("aws_workspaces_directory.test"),
				),
			},
			{
				ResourceName:      "aws_workspaces_directory.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsWorkspacesDirectoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).workspacesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_directory" {
			continue
		}

		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.Directories) == 0 {
			return nil
		}

		dir := resp.Directories[0]
		if *dir.State != workspaces.WorkspaceDirectoryStateDeregistering && *dir.State != workspaces.WorkspaceDirectoryStateDeregistered {
			return fmt.Errorf("directory %q was not deregistered", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesDirectoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("workspaces directory resource is not found: %q", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("workspaces directory resource ID is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if *resp.Directories[0].DirectoryId == rs.Primary.ID {
			return nil
		}

		return fmt.Errorf("workspaces directory %q is not found", rs.Primary.ID)
	}
}
