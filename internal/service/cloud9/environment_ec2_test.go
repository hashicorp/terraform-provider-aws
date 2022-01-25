package cloud9_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloud9"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloud9 "github.com/hashicorp/terraform-provider-aws/internal/service/cloud9"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloud9EnvironmentEC2_basic(t *testing.T) {
	var conf cloud9.Environment

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloud9.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "data.aws_caller_identity.current", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id"},
			},
			{
				Config: testAccEnvironmentEC2Config(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "data.aws_caller_identity.current", "arn"),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_allFields(t *testing.T) {
	var conf cloud9.Environment

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	description := sdkacctest.RandomWithPrefix("Tf Acc Test")
	uDescription := sdkacctest.RandomWithPrefix("Tf Acc Test Updated")
	userName := sdkacctest.RandomWithPrefix("tf_acc_cloud9_env")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloud9.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2AllFieldsConfig(rName, description, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "automatic_stop_time_minutes", "subnet_id"},
			},
			{
				Config: testAccEnvironmentEC2AllFieldsConfig(rNameUpdated, uDescription, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_tags(t *testing.T) {
	var conf cloud9.Environment

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloud9.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Tags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id"},
			},
			{
				Config: testAccEnvironmentEC2Tags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEnvironmentEC2Tags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_disappears(t *testing.T) {
	var conf cloud9.Environment

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloud9.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnvironmentEC2Exists(n string, res *cloud9.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cloud9 Environment EC2 ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Conn

		out, err := tfcloud9.FindEnvironmentByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckEnvironmentEC2Destroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloud9_environment_ec2" {
			continue
		}

		_, err := tfcloud9.FindEnvironmentByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Cloud9 Environment EC2 %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccEnvironmentEC2BaseConfig() string {
	return `
data "aws_availability_zones" "available" {
  # t2.micro instance type is not available in these Availability Zones
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}
`
}

func testAccEnvironmentEC2Config(name string) string {
	return testAccEnvironmentEC2BaseConfig() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id
}

# By default, the Cloud9 environment EC2 is owned by the creator
data "aws_caller_identity" "current" {}
`, name)
}

func testAccEnvironmentEC2AllFieldsConfig(name, description, userName string) string {
	return testAccEnvironmentEC2BaseConfig() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  automatic_stop_time_minutes = 60
  description                 = %[2]q
  instance_type               = "t2.micro"
  name                        = %[1]q
  owner_arn                   = aws_iam_user.test.arn
  subnet_id                   = aws_subnet.test.id
}

resource "aws_iam_user" "test" {
  name = %[3]q
}
`, name, description, userName)
}

func testAccEnvironmentEC2Tags1Config(name, tagKey1, tagValue1 string) string {
	return testAccEnvironmentEC2BaseConfig() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccEnvironmentEC2Tags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccEnvironmentEC2BaseConfig() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
