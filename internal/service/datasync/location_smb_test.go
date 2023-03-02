package datasync_test

import (
	"context"
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

func TestAccDataSyncLocationSMB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var locationSmb1 datasync.DescribeLocationSmbOutput

	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^smb://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^smb://.+/`)),
				),
			},
		},
	})
}

func TestAccDataSyncLocationSMB_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var locationSmb1 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb1),
					testAccCheckLocationSMBDisappears(ctx, &locationSmb1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationSMB_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var locationSmb1, locationSmb2, locationSmb3 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
			{
				Config: testAccLocationSMBConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb2),
					testAccCheckLocationSMBNotRecreated(&locationSmb1, &locationSmb2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationSMBConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &locationSmb3),
					testAccCheckLocationSMBNotRecreated(&locationSmb2, &locationSmb3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationSMBDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_smb" {
				continue
			}

			input := &datasync.DescribeLocationSmbInput{
				LocationArn: aws.String(rs.Primary.ID),
			}

			_, err := conn.DescribeLocationSmbWithContext(ctx, input)

			if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
				return nil
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckLocationSMBExists(ctx context.Context, resourceName string, locationSmb *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn()
		input := &datasync.DescribeLocationSmbInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationSmbWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationSmb = *output

		return nil
	}
}

func testAccCheckLocationSMBDisappears(ctx context.Context, location *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn()

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocationWithContext(ctx, input)

		return err
	}
}

func testAccCheckLocationSMBNotRecreated(i, j *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location SMB was recreated")
		}

		return nil
	}
}

func testAccLocationSMBConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
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
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_default_route_table.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationSMBConfig_basic(rName, dir string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = %[1]q
  user            = "Guest"
}
`, dir))
}

func testAccLocationSMBConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationSMBConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}
