package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EIPsDataSource_vpcDomain(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPsVPCDomainDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue("data.aws_eips.all", "allocation_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_eips.by_tags", "allocation_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_eips.by_tags", "public_ips.#", "1"),
					resource.TestCheckResourceAttr("data.aws_eips.none", "allocation_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_eips.none", "public_ips.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2EIPsDataSource_standardDomain(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPsStandardDomainDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue("data.aws_eips.all", "public_ips.#", "0"),
				),
			},
		},
	})
}

func testAccEIPsVPCDomainDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test1" {
  vpc = true

  tags = {
    Name = "%[1]s-1"
  }
}

resource "aws_eip" "test2" {
  vpc = true

  tags = {
    Name = "%[1]s-2"
  }
}

data "aws_eips" "all" {
  depends_on = [aws_eip.test1, aws_eip.test2]
}

data "aws_eips" "by_tags" {
  tags = {
    Name = "%[1]s-1"
  }

  depends_on = [aws_eip.test1, aws_eip.test2]
}

data "aws_eips" "none" {
  filter {
    name   = "tag-key"
    values = ["%[1]s-3"]
  }

  depends_on = [aws_eip.test1, aws_eip.test2]
}
`, rName)
}

func testAccEIPsStandardDomainDataSourceConfig() string {
	return acctest.ConfigCompose(acctest.ConfigEC2ClassicRegionProvider(), `
resource "aws_eip" "test" {}

data "aws_eips" "all" {
  depends_on = [aws_eip.test]
}
`)
}
