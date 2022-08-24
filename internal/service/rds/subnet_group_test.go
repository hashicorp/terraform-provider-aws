package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const ipv4SupportedNetworkType = "IPV4"
const dualSupportedNetworkType = "DUAL"

func TestAccRDSSubnetGroup_basic(t *testing.T) {
	var v rds.DBSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("subgrp:%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "supported_network_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "supported_network_types.0", "IPV4"),
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

func TestAccRDSSubnetGroup_nameGenerated(t *testing.T) {
	var v rds.DBSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
				),
			},
		},
	})
}

func TestAccRDSSubnetGroup_namePrefix(t *testing.T) {
	var v rds.DBSubnetGroup
	resourceName := "aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
		},
	})
}

func TestAccRDSSubnetGroup_dualStack(t *testing.T) {
	var v rds.DBSubnetGroup

	resourceName := "aws_db_subnet_group.test"
	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_dualStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "supported_network_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "supported_network_types.0", dualSupportedNetworkType),
					resource.TestCheckResourceAttr(resourceName, "supported_network_types.1", ipv4SupportedNetworkType),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/2603 and
// https://github.com/hashicorp/terraform/issues/2664
func TestAccRDSSubnetGroup_withUndocumentedCharacters(t *testing.T) {
	var v rds.DBSubnetGroup

	testCheck := func(*terraform.State) error {
		return nil
	}
	resourceName := "aws_db_subnet_group.underscores"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.underscores", &v),
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.periods", &v),
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.spaces", &v),
					testCheck,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})
}

func TestAccRDSSubnetGroup_updateDescription(t *testing.T) {
	var v rds.DBSubnetGroup
	resourceName := "aws_db_subnet_group.test"
	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccDBSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test description updated"),
				),
			},
		},
	})
}

func testAccCheckDBSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_subnet_group" {
			continue
		}

		_, err := tfrds.FindDBSubnetGroupByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("RDS DB Subnet Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDBSubnetGroupExists(n string, v *rds.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS DB Subnet Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		output, err := tfrds.FindDBSubnetGroupByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubnetGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccSubnetGroupConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), `
resource "aws_db_subnet_group" "test" {
  subnet_ids = aws_subnet.test[*].id
}`)
}

func testAccSubnetGroupConfig_namePrefix(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name_prefix = %[1]q
  subnet_ids  = aws_subnet.test[*].id
}`, rName))
}

func testAccSubnetGroupConfig_dualStack(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true	

  tags = {
    Name = "terraform-testacc-db-subnet-group"
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = "10.1.1.0/24"
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0)
  assign_ipv6_address_on_creation = true
  availability_zone               = data.aws_availability_zones.available.names[0]
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block                      = "10.1.2.0/24"
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  availability_zone               = data.aws_availability_zones.available.names[1]
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%s"
  subnet_ids = [aws_subnet.test.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}
`, rName)
}

func testAccDBSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-subnet-group-updated-description"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name        = "%s"
  description = "test description updated"
  subnet_ids  = [aws_subnet.test.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}
`, rName)
}

const testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "terraform-testacc-db-subnet-group-w-underscores-etc"
  }
}

resource "aws_subnet" "frontend" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.1.0/24"
  tags = {
    Name = "tf-acc-db-subnet-group-w-underscores-etc-front"
  }
}

resource "aws_subnet" "backend" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.2.0/24"
  tags = {
    Name = "tf-acc-db-subnet-group-w-underscores-etc-back"
  }
}

resource "aws_db_subnet_group" "underscores" {
  name        = "with_underscores"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}

resource "aws_db_subnet_group" "periods" {
  name        = "with.periods"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}

resource "aws_db_subnet_group" "spaces" {
  name        = "with spaces"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}
`
