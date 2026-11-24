// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesResourceAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53profiles_resource_associations.test"
	emptyDataSourceName := "data.aws_route53profiles_resource_associations.empty"
	resourceName := "aws_route53profiles_resource_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_associations.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.owner_id", resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.profile_id", resourceName, "profile_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.resource_arn", resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_associations.0.resource_type", resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttr(dataSourceName, "resource_associations.0.status", string(awstypes.ProfileStatusComplete)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(emptyDataSourceName, tfjsonpath.New("resource_associations"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccRoute53ProfilesResourceAssociationsDataSource_resourceType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	allDataSourceName := "data.aws_route53profiles_resource_associations.all"
	filteredDataSourceName := "data.aws_route53profiles_resource_associations.firewall"
	zoneResourceName := "aws_route53profiles_resource_association.zone"
	firewallResourceName := "aws_route53profiles_resource_association.firewall"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationsDataSourceConfig_resourceType(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(allDataSourceName, "resource_associations.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(allDataSourceName, "resource_associations.*.id", zoneResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(allDataSourceName, "resource_associations.*.id", firewallResourceName, names.AttrID),
					resource.TestCheckResourceAttr(filteredDataSourceName, "resource_associations.#", "1"),
					resource.TestCheckResourceAttrPair(filteredDataSourceName, "resource_associations.0.id", firewallResourceName, names.AttrID),
					resource.TestCheckResourceAttr(filteredDataSourceName, "resource_associations.0.resource_type", "FIREWALL_RULE_GROUP"),
					resource.TestCheckResourceAttrPair(filteredDataSourceName, "resource_associations.0.resource_properties", firewallResourceName, "resource_properties"),
				),
			},
		},
	})
}

func testAccResourceAssociationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceAssociationConfig_basic(rName), fmt.Sprintf(`
resource "aws_route53profiles_profile" "empty" {
  name = "%[1]s-empty"
}

data "aws_route53profiles_resource_associations" "test" {
  profile_id = aws_route53profiles_resource_association.test.profile_id
}

data "aws_route53profiles_resource_associations" "empty" {
  profile_id = aws_route53profiles_profile.empty.id
}
`, rName))
}

func testAccResourceAssociationsDataSourceConfig_resourceType(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53_zone" "test" {
  name = %[2]q

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53profiles_resource_association" "zone" {
  name         = "%[1]s-zone"
  profile_id   = aws_route53profiles_profile.test.id
  resource_arn = aws_route53_zone.test.arn
}

resource "aws_route53profiles_resource_association" "firewall" {
  name                = "%[1]s-firewall"
  profile_id          = aws_route53profiles_profile.test.id
  resource_arn        = aws_route53_resolver_firewall_rule_group.test.arn
  resource_properties = jsonencode({ "priority" = 102 })
}

data "aws_route53profiles_resource_associations" "all" {
  profile_id = aws_route53profiles_profile.test.id

  depends_on = [
    aws_route53profiles_resource_association.zone,
    aws_route53profiles_resource_association.firewall,
  ]
}

data "aws_route53profiles_resource_associations" "firewall" {
  profile_id    = aws_route53profiles_profile.test.id
  resource_type = "FIREWALL_RULE_GROUP"

  depends_on = [
    aws_route53profiles_resource_association.zone,
    aws_route53profiles_resource_association.firewall,
  ]
}
`, rName, domainName)
}
