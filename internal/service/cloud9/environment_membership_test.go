package cloud9_test

import (
	"fmt"
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

func TestAccEnvironmentMembership_basic(t *testing.T) {
	var conf cloud9.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "permissions", "read-only"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "aws_cloud9_environment_ec2.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnvironmentMembershipConfig(rName, "read-write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "permissions", "read-write"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "aws_cloud9_environment_ec2.test", "id"),
				),
			},
		},
	})
}

func TestAccEnvironmentMembership_disappears(t *testing.T) {
	var conf cloud9.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentMembership(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEnvironmentMembership_disappears_env(t *testing.T) {
	var conf cloud9.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloud9.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloud9.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), "aws_cloud9_environment_ec2.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloud9.ResourceEnvironmentMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnvironmentMemberExists(n string, res *cloud9.EnvironmentMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cloud9 Environment Member ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Conn

		envId, userArn, err := tfcloud9.DecodeEnviornmentMemberId(rs.Primary.ID)
		if err != nil {
			return err
		}

		out, err := tfcloud9.FindEnvironmentMembershipByID(conn, envId, userArn)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckEnvironmentMemberDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloud9_environment_membership" {
			continue
		}

		envId, userArn, err := tfcloud9.DecodeEnviornmentMemberId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfcloud9.FindEnvironmentMembershipByID(conn, envId, userArn)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Cloud9 Environment Membership %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccEnvironmentMemberBaseConfig(rName string) string {
	return fmt.Sprintf(`
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

resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id
}
`, rName)
}

func testAccEnvironmentMembershipConfig(rName, permissions string) string {
	return testAccEnvironmentMemberBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_cloud9_environment_membership" "test" {
  environment_id = aws_cloud9_environment_ec2.test.id
  permissions    = %[2]q
  user_arn       = aws_iam_user.test.arn
}
`, rName, permissions)
}
