package docdb_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
)

func TestAccDocDBSubnetGroup_basic(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDocDBSubnetGroup_disappears(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					testAccCheckSubnetGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBSubnetGroup_namePrefix(t *testing.T) {
	var v docdb.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_subnet_group.test", "name", regexp.MustCompile("^tf_test-")),
				),
			},
			{
				ResourceName:            "aws_docdb_subnet_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccDocDBSubnetGroup_generatedName(t *testing.T) {
	var v docdb.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDocDBSubnetGroup_updateDescription(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},

			{
				Config: testAccSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "foo description updated"),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_subnet_group" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDBSubnetGroups(
			&docdb.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})

		if err == nil {
			if len(resp.DBSubnetGroups) != 0 &&
				aws.StringValue(resp.DBSubnetGroups[0].DBSubnetGroupName) == rs.Primary.ID {
				return fmt.Errorf("DocDB Subnet Group %s still exists", rs.Primary.ID)
			}
		}

		if err != nil {
			if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckSubnetGroupDisappears(group *docdb.DBSubnetGroup) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		params := &docdb.DeleteDBSubnetGroupInput{
			DBSubnetGroupName: group.DBSubnetGroupName,
		}

		_, err := conn.DeleteDBSubnetGroup(params)
		if err != nil {
			return err
		}

		return tfdocdb.WaitForSubnetGroupDeletion(conn, *group.DBSubnetGroupName)
	}
}

func testAccCheckSubnetGroupExists(n string, v *docdb.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn
		resp, err := conn.DescribeDBSubnetGroups(
			&docdb.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if len(resp.DBSubnetGroups) == 0 {
			return fmt.Errorf("DbSubnetGroup not found")
		}

		*v = *resp.DBSubnetGroups[0]

		return nil
	}
}

func testAccSubnetGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-docdb-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-docdb-subnet-group-2"
  }
}

resource "aws_docdb_subnet_group" "foo" {
  name       = "%s"
  subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-docdb-subnet-group-test"
  }
}
`, rName))
}

func testAccSubnetGroupConfig_updatedDescription(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-docdb-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-docdb-subnet-group-2"
  }
}

resource "aws_docdb_subnet_group" "foo" {
  name        = "%s"
  description = "foo description updated"
  subnet_ids  = [aws_subnet.foo.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-docdb-subnet-group-test"
  }
}
`, rName))
}

func testAccSubnetGroupConfig_namePrefix() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-subnet-group-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-docdb-subnet-group-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-docdb-subnet-group-name-prefix-b"
  }
}

resource "aws_docdb_subnet_group" "test" {
  name_prefix = "tf_test-"
  subnet_ids  = [aws_subnet.a.id, aws_subnet.b.id]
}`)
}

func testAccSubnetGroupConfig_generatedName() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-subnet-group-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-docdb-subnet-group-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-docdb-subnet-group-generated-name-a"
  }
}

resource "aws_docdb_subnet_group" "test" {
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}`)
}
