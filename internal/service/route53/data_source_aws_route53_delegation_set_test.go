package route53_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSRoute53DelegationSetDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_delegation_set.dset"
	resourceName := "aws_route53_delegation_set.dset"

	zoneName := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceAWSRoute53DelegationSetConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name_servers.#", resourceName, "name_servers.#"),
					resource.TestMatchResourceAttr("data.aws_route53_delegation_set.dset", "caller_reference", regexp.MustCompile("DynDNS(.*)")),
				),
			},
		},
	})
}

func testAccAWSDataSourceAWSRoute53DelegationSetConfig_basic(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "dset" {
  reference_name = "DynDNS"
}

resource "aws_route53_zone" "primary" {
  name              = %[1]q
  delegation_set_id = aws_route53_delegation_set.dset.id
}

data "aws_route53_delegation_set" "dset" {
  id = aws_route53_delegation_set.dset.id
}
`, zoneName)
}
