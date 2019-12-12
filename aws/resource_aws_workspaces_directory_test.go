package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	testAccWorkspaceConfig = `
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "primary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[0]}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[1]}"
  cidr_block = "10.0.2.0/24"
}

resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
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
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "primary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[0]}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[1]}"
  cidr_block = "10.0.2.0/24"
}
resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
  }
}

data aws_iam_policy_document workspaces {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  tags = {
    Name = "test"
    Terraform = true
    Directory = "tf-acctest.example.com"
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
  subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
}
`

	testAccWorkspaceConfig_selfServicePermissionsA = `
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "primary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[0]}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[1]}"
  cidr_block = "10.0.2.0/24"
}
resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
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
  subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]

  self_service_permissions {
    change_compute_type = true
    increase_volume_size = true
    rebuild_workspace = true
    restart_workspace = true
    switch_running_mode = true
  }

  tags = {
    Purpose   = "test"
    Directory = "tf-acctest.example.com"
  }
}
`

	testAccWorkspaceConfig_selfServicePermissionsB = `
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "primary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[0]}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone_id = "${local.workspaces_az_ids[1]}"
  cidr_block = "10.0.2.0/24"
}
resource "aws_directory_service_directory" "test" {
  name = "tf-acctest.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
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
  subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]

  self_service_permissions {
    change_compute_type = false
    increase_volume_size = true
    rebuild_workspace = false
    restart_workspace = true
    switch_running_mode = false
  }
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
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.Name", "test"),
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.Terraform", "true"),
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.Directory", "tf-acctest.example.com"),
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
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.Directory", "tf-acctest.example.com"),
					resource.TestCheckResourceAttr("aws_workspaces_directory.test", "tags.Purpose", "test"),
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

func TestAccAwsWorkspacesDirectory_selfServicePermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_selfServicePermissionsA,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists("aws_workspaces_directory.test"),
				),
			},
			{
				ResourceName:      "aws_workspaces_directory.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_selfServicePermissionsB,
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

func TestExpandSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.SelfservicePermissions
	}{
		// Empty
		{
			input:    []interface{}{},
			expected: nil,
		},
		// Full
		{
			input: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
			expected: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
		},
	}

	for _, c := range cases {
		actual := expandSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    *workspaces.SelfservicePermissions
		expected []interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []interface{}{},
		},
		// Full
		{
			input: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
			expected: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}
