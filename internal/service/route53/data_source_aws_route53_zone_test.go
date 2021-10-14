package route53_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSRoute53ZoneDataSource_id(t *testing.T) {
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigId(fqdn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ZoneDataSource_name(t *testing.T) {
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigName(fqdn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ZoneDataSource_tags(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	fqdn := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigTagsPrivate(fqdn, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ZoneDataSource_vpc(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigVpc(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers.#", dataSourceName, "name_servers.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ZoneDataSource_serviceDiscovery(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("servicediscovery", t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigServiceDiscovery(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "linked_service_principal", "servicediscovery.amazonaws.com"),
					resource.TestCheckResourceAttrPair(dataSourceName, "linked_service_description", resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRoute53ZoneConfigId(fqdn string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

data "aws_route53_zone" "test" {
  zone_id = aws_route53_zone.test.zone_id
}
`, fqdn)
}

func testAccDataSourceAwsRoute53ZoneConfigName(fqdn string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}

data "aws_route53_zone" "test" {
  name = aws_route53_zone.test.name
}
`, fqdn)
}

func testAccDataSourceAwsRoute53ZoneConfigTagsPrivate(fqdn string, rInt int) string {
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

func testAccDataSourceAwsRoute53ZoneConfigVpc(rInt int) string {
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

func testAccDataSourceAwsRoute53ZoneConfigServiceDiscovery(rInt int) string {
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
