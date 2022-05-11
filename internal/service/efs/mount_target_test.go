package efs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
)

func TestAccEFSMountTarget_basic(t *testing.T) {
	var mount efs.MountTargetDescription
	ct := fmt.Sprintf("createtoken-%d", sdkacctest.RandInt())
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMountTargetModifiedConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTarget(resourceName, &mount),
					testAccCheckMountTarget(resourceName2, &mount),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalHostname(resourceName2, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
				),
			},
		},
	})
}

func TestAccEFSMountTarget_disappears(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"
	ct := fmt.Sprintf("createtoken-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTarget(resourceName, &mount),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceMountTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSMountTarget_ipAddress(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetIPAddressConfig(rName, "10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.100"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13845
func TestAccEFSMountTarget_IPAddress_emptyString(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetIPAddressConfigNullIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTarget(resourceName, &mount),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func TestMountTarget_hasEmptyMountTargets(t *testing.T) {
	mto := &efs.DescribeMountTargetsOutput{
		MountTargets: []*efs.MountTargetDescription{},
	}

	actual := tfefs.HasEmptyMountTargets(mto)
	if !actual {
		t.Fatalf("Expected return value to be true, got %t", actual)
	}

	// Add an empty mount target.
	mto.MountTargets = append(mto.MountTargets, &efs.MountTargetDescription{})

	actual = tfefs.HasEmptyMountTargets(mto)
	if actual {
		t.Fatalf("Expected return value to be false, got %t", actual)
	}
}

func testAccCheckMountTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_mount_target" {
			continue
		}

		resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, efs.ErrCodeMountTargetNotFound) {
				// gone
				return nil
			}
			return fmt.Errorf("Error describing EFS Mount in tests: %s", err)
		}
		if len(resp.MountTargets) > 0 {
			return fmt.Errorf("EFS Mount target %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckMountTarget(resourceID string, mount *efs.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		fs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
		mt, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(fs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if aws.StringValue(mt.MountTargets[0].MountTargetId) != fs.Primary.ID {
			return fmt.Errorf("Mount target ID mismatch: %q != %q",
				*mt.MountTargets[0].MountTargetId, fs.Primary.ID)
		}

		*mount = *mt.MountTargets[0]

		return nil
	}
}

func testAccMountTargetConfig(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}
`, ct)
}

func testAccMountTargetModifiedConfig(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_efs_mount_target" "test2" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test2.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test2"
  }
}
`, ct)
}

func testAccMountTargetIPAddressConfig(rName, ipAddress string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

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
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = %[2]q
  subnet_id      = aws_subnet.test.id
}
`, rName, ipAddress)
}

func testAccMountTargetIPAddressConfigNullIP(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

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
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = null
  subnet_id      = aws_subnet.test.id
}
`, rName)
}

func testAccCheckVpnGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{
			VpnGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			var v *ec2.VpnGateway
			for _, g := range resp.VpnGateways {
				if *g.VpnGatewayId == rs.Primary.ID {
					v = g
				}
			}

			if v == nil {
				// wasn't found
				return nil
			}

			if *v.State != "deleted" {
				return fmt.Errorf("Expected VPN Gateway to be in deleted state, but was not: %s", v)
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpnGatewayID.NotFound" {
			return err
		}
	}

	return nil
}
