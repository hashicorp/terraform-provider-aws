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
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "CONNECT_SSH"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id"},
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_allFields(t *testing.T) {
	var conf cloud9.Environment

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentEC2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2AllFieldsConfig(rName, name1, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "automatic_stop_time_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "CONNECT_SSH"),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
					resource.TestCheckResourceAttr(resourceName, "image_id", "amazonlinux-2-x86_64"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"automatic_stop_time_minutes", "image_id", "instance_type", "subnet_id"},
			},
			{
				Config: testAccEnvironmentEC2AllFieldsConfig(rName, name2, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
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
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentEC2Destroy,
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
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentEC2Destroy,
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

func testAccCheckEnvironmentEC2Exists(n string, v *cloud9.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cloud9 Environment EC2 ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Conn

		output, err := tfcloud9.FindEnvironmentByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

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

		return fmt.Errorf("Cloud9 Environment EC2 %s still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccEnvironmentEC2BaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}
`, rName))
}

func testAccEnvironmentEC2Config(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2BaseConfig(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id
}
`, rName))
}

func testAccEnvironmentEC2AllFieldsConfig(rName, name, description string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2BaseConfig(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  automatic_stop_time_minutes = 60
  description                 = %[2]q
  instance_type               = "t2.micro"
  name                        = %[1]q
  owner_arn                   = aws_iam_user.test.arn
  subnet_id                   = aws_subnet.test.id
  connection_type             = "CONNECT_SSH"
  image_id                    = "amazonlinux-2-x86_64"
}

resource "aws_iam_user" "test" {
  name = %[3]q
}
`, name, description, rName))
}

func testAccEnvironmentEC2Tags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2BaseConfig(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEnvironmentEC2Tags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2BaseConfig(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
