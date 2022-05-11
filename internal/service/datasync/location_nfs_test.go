package datasync_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDataSyncLocationNFS_basic(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^nfs://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
		},
	})
}

func TestAccDataSyncLocationNFS_mountOptions(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSMountOptionsConfig(rName, "NFS4_0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "NFS4_0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
			{
				Config: testAccLocationNFSMountOptionsConfig(rName, "NFS4_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "NFS4_1"),
				),
			},
		},
	})
}

func TestAccDataSyncLocationNFS_disappears(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					testAccCheckLocationNFSDisappears(&locationNfs1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationNFS_AgentARNs_multiple(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSAgentARNsMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
		},
	})
}

func TestAccDataSyncLocationNFS_subdirectory(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var locationNfs1 datasync.DescribeLocationNfsOutput
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSSubdirectoryConfig(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
			{
				Config: testAccLocationNFSSubdirectoryConfig(rName, "/subdirectory2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory2/"),
				),
			},
		},
	})
}

func TestAccDataSyncLocationNFS_tags(t *testing.T) {
	var locationNfs1, locationNfs2, locationNfs3 datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationNFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
			{
				Config: testAccLocationNFSTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs2),
					testAccCheckLocationNFSNotRecreated(&locationNfs1, &locationNfs2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationNFSTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(resourceName, &locationNfs3),
					testAccCheckLocationNFSNotRecreated(&locationNfs2, &locationNfs3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationNFSDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_nfs" {
			continue
		}

		input := &datasync.DescribeLocationNfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationNfs(input)

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationNFSExists(resourceName string, locationNfs *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationNfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationNfs(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationNfs = *output

		return nil
	}
}

func testAccCheckLocationNFSDisappears(location *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckLocationNFSNotRecreated(i, j *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location Nfs was recreated")
		}

		return nil
	}
}

func testAccLocationNFSBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_internet_gateway.test]

  ami                         = data.aws_ami.aws-thinstaller.id
  associate_public_ip_address = true

  # Default instance type from sync.sh
  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %q
}
`, rName)
}

func testAccLocationNFSConfig(rName string) string {
	return testAccLocationNFSBaseConfig(rName) + `
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}
`
}

func testAccLocationNFSMountOptionsConfig(rName, option string) string {
	return testAccLocationNFSBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  mount_options {
    version = %[1]q
  }
}
`, option)
}

func testAccLocationNFSAgentARNsMultipleConfig(rName string) string {
	return testAccLocationNFSBaseConfig(rName) + fmt.Sprintf(`
resource "aws_instance" "test2" {
  depends_on = [aws_internet_gateway.test]

  ami                         = data.aws_ami.aws-thinstaller.id
  associate_public_ip_address = true

  # Default instance type from sync.sh
  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_datasync_agent" "test2" {
  ip_address = aws_instance.test2.public_ip
  name       = "%s2"
}

resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [
      aws_datasync_agent.test.arn,
      aws_datasync_agent.test2.arn,
    ]
  }
}
`, rName)
}

func testAccLocationNFSSubdirectoryConfig(rName, subdirectory string) string {
	return testAccLocationNFSBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = %q

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}
`, subdirectory)
}

func testAccLocationNFSTags1Config(rName, key1, value1 string) string {
	return testAccLocationNFSBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccLocationNFSTags2Config(rName, key1, value1, key2, value2 string) string {
	return testAccLocationNFSBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
