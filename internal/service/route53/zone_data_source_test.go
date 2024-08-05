// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ZoneDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_id(fqdn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_name_server", dataSourceName, "primary_name_server"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

func TestAccRoute53ZoneDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_name(fqdn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_name_server", dataSourceName, "primary_name_server"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

// Verifies the data source works when name is set and zone_id is an empty string.
//
// This may be behavior we want to disable in the future with an ExactlyOneOf
// constraint, but because we've historically allowed it we need to continue
// doing so until a major version bump.
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/37683
func TestAccRoute53ZoneDataSource_name_idEmptyString(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_name_idEmptyString(fqdn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_name_server", dataSourceName, "primary_name_server"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

func TestAccRoute53ZoneDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_tagsPrivate(fqdn, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_name_server", dataSourceName, "primary_name_server"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

func TestAccRoute53ZoneDataSource_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_vpc(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_name_server", dataSourceName, "primary_name_server"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTags, dataSourceName, names.AttrTags),
				),
			},
		},
	})
}

func TestAccRoute53ZoneDataSource_serviceDiscovery(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandInt()
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "servicediscovery") },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_serviceDiscovery(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "linked_service_principal", "servicediscovery.amazonaws.com"),
					resource.TestCheckResourceAttrPair(dataSourceName, "linked_service_description", resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig_id(fqdn string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

data "aws_route53_zone" "test" {
  zone_id = aws_route53_zone.test.zone_id
}
`, fqdn)
}

func testAccZoneDataSourceConfig_name(fqdn string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

data "aws_route53_zone" "test" {
  name = aws_route53_zone.test.name
}
`, fqdn)
}

func testAccZoneDataSourceConfig_name_idEmptyString(fqdn string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

data "aws_route53_zone" "test" {
  zone_id = ""
  name    = aws_route53_zone.test.name
}
`, fqdn)
}

func testAccZoneDataSourceConfig_tagsPrivate(fqdn string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53_zone" "test" {
  name = %[1]q

  vpc {
    vpc_id = aws_vpc.test.id
  }

  tags = {
    Environment = "tf-acc-test-%[2]d"
    Name        = "tf-acc-test-%[2]d"
  }
}

data "aws_route53_zone" "test" {
  name         = aws_route53_zone.test.name
  private_zone = true
  vpc_id       = aws_vpc.test.id

  tags = {
    Environment = "tf-acc-test-%[2]d"
  }
}
`, fqdn, rInt)
}

func testAccZoneDataSourceConfig_vpc(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source-%[1]d"
  }
}

resource "aws_route53_zone" "test" {
  name = "test.acc-%[1]d."

  vpc {
    vpc_id = aws_vpc.test.id
  }

  tags = {
    Environment = "dev-%[1]d"
  }
}

data "aws_route53_zone" "test" {
  name         = aws_route53_zone.test.name
  private_zone = true
  vpc_id       = aws_vpc.test.id
}
`, rInt)
}

func testAccZoneDataSourceConfig_serviceDiscovery(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source-%[1]d"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "test.acc-sd-%[1]d"
  vpc  = aws_vpc.test.id
}

data "aws_route53_zone" "test" {
  name   = aws_service_discovery_private_dns_namespace.test.name
  vpc_id = aws_vpc.test.id
}
`, rInt)
}
