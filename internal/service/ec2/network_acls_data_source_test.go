package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsNetworkAcls_basic(t *testing.T) {
	rName := sdkacctest.RandString(5)
	dataSourceName := "data.aws_network_acls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				// Ensure at least 1 network ACL exists. We cannot use depends_on.
				Config: testAccNetworkACLsDataSourceConfig_Base(rName),
			},
			{
				Config: testAccNetworkACLsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					// At least 1
					resource.TestMatchResourceAttr(dataSourceName, "ids.#", regexp.MustCompile(`^[1-9][0-9]*`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_Filter(t *testing.T) {
	rName := sdkacctest.RandString(5)
	dataSourceName := "data.aws_network_acls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLsDataSourceConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_Tags(t *testing.T) {
	rName := sdkacctest.RandString(5)
	dataSourceName := "data.aws_network_acls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLsDataSourceConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_VpcID(t *testing.T) {
	rName := sdkacctest.RandString(5)
	dataSourceName := "data.aws_network_acls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLsDataSourceConfig_VPCID(rName),
				Check: resource.ComposeTestCheckFunc(
					// The VPC will have a default network ACL
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "3"),
				),
			},
		},
	})
}

func testAccNetworkACLsDataSourceConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "testacc-acl-%[1]s"
  }
}

resource "aws_network_acl" "acl" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = "testacc-acl-%[1]s"
  }
}
`, rName)
}

func testAccNetworkACLsDataSourceConfig_basic(rName string) string {
	return testAccNetworkACLsDataSourceConfig_Base(rName) + `
data "aws_network_acls" "test" {}
`
}

func testAccNetworkACLsDataSourceConfig_Filter(rName string) string {
	return testAccNetworkACLsDataSourceConfig_Base(rName) + `
data "aws_network_acls" "test" {
  filter {
    name   = "network-acl-id"
    values = [aws_network_acl.acl[0].id]
  }
}
`
}

func testAccNetworkACLsDataSourceConfig_Tags(rName string) string {
	return testAccNetworkACLsDataSourceConfig_Base(rName) + `
data "aws_network_acls" "test" {
  tags = {
    Name = aws_network_acl.acl[0].tags.Name
  }
}
`
}

func testAccNetworkACLsDataSourceConfig_VPCID(rName string) string {
	return testAccNetworkACLsDataSourceConfig_Base(rName) + `
data "aws_network_acls" "test" {
  vpc_id = aws_network_acl.acl[0].vpc_id
}
`
}
