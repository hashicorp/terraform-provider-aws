package datasync_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataSyncLocationHdfs_basic(t *testing.T) {
	var locationHdfs1 datasync.DescribeLocationHdfsOutput

	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationHdfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHdfsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHdfsExists(resourceName, &locationHdfs1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name_node.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "name_node.*", map[string]string{
						"port": "80",
					}),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "SIMPLE"),
					resource.TestCheckResourceAttr(resourceName, "simple_user", rName),
					resource.TestCheckResourceAttr(resourceName, "block_size", "134217728"),
					resource.TestCheckResourceAttr(resourceName, "replication_factor", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^hdfs://.+/`)),
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

func TestAccDataSyncLocationHdfs_disappears(t *testing.T) {
	var locationHdfs1 datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationHdfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHdfsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHdfsExists(resourceName, &locationHdfs1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationHdfs(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationHdfs(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationHdfs_tags(t *testing.T) {
	var locationHdfs1, locationHdfs2, locationHdfs3 datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationHdfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHdfsTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHdfsExists(resourceName, &locationHdfs1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationHdfsTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHdfsExists(resourceName, &locationHdfs2),
					testAccCheckLocationHdfsNotRecreated(&locationHdfs1, &locationHdfs2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationHdfsTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHdfsExists(resourceName, &locationHdfs3),
					testAccCheckLocationHdfsNotRecreated(&locationHdfs2, &locationHdfs3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationHdfsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_hdfs" {
			continue
		}

		_, err := tfdatasync.FindLocationHdfsByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataSync Task %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckLocationHdfsExists(resourceName string, locationHdfs *datasync.DescribeLocationHdfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		output, err := tfdatasync.FindLocationHdfsByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationHdfs = *output

		return nil
	}
}

func testAccCheckLocationHdfsNotRecreated(i, j *datasync.DescribeLocationHdfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location Hdfs was recreated")
		}

		return nil
	}
}

func testAccLocationHdfsBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-hdfs"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-hdfs"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-hdfs"
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "tf-acc-test-datasync-location-hdfs"
  }
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
    Name = "tf-acc-test-datasync-hdfs"
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_default_route_table.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = "tf-acc-test-datasync-hdfs"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationHdfsConfig(rName string) string {
	return testAccLocationHdfsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }
}
`, rName)
}

func testAccLocationHdfsTags1Config(rName, key1, value1 string) string {
	return testAccLocationHdfsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1)
}

func testAccLocationHdfsTags2Config(rName, key1, value1, key2, value2 string) string {
	return testAccLocationHdfsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2)
}
