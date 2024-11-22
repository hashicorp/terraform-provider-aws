// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53DelegationSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_delegation_set.dset"
	resourceName := "aws_route53_delegation_set.dset"

	zoneName := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetDataSourceConfig_basic(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "name_servers.#", resourceName, "name_servers.#"),
					resource.TestMatchResourceAttr("data.aws_route53_delegation_set.dset", "caller_reference", regexache.MustCompile("DynDNS(.*)")),
				),
			},
		},
	})
}

func testAccDelegationSetDataSourceConfig_basic(zoneName string) string {
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
