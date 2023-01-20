package datasync_test

import (
	"context"
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

func TestAccDataSyncLocationObjectStorage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage1 datasync.DescribeLocationObjectStorageOutput

	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", rName),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^object-storage://.+/`)),
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

func TestAccDataSyncLocationObjectStorage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage1 datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationObjectStorage(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationObjectStorage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationObjectStorage_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage1, locationObjectStorage2, locationObjectStorage3 datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage1),
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
				Config: testAccLocationObjectStorageConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage2),
					testAccCheckLocationObjectStorageNotRecreated(&locationObjectStorage1, &locationObjectStorage2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationObjectStorageConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage3),
					testAccCheckLocationObjectStorageNotRecreated(&locationObjectStorage2, &locationObjectStorage3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationObjectStorageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_object_storage" {
				continue
			}

			_, err := tfdatasync.FindLocationObjectStorageByARN(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckLocationObjectStorageExists(ctx context.Context, resourceName string, locationObjectStorage *datasync.DescribeLocationObjectStorageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn()
		output, err := tfdatasync.FindLocationObjectStorageByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationObjectStorage = *output

		return nil
	}
}

func testAccCheckLocationObjectStorageNotRecreated(i, j *datasync.DescribeLocationObjectStorageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location Object Storage was recreated")
		}

		return nil
	}
}

func testAccLocationObjectStorageBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
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

func testAccLocationObjectStorageConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[1]q
  bucket_name     = %[1]q
}
`, rName))
}

func testAccLocationObjectStorageConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[1]q
  bucket_name     = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccLocationObjectStorageConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[1]q
  bucket_name     = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
