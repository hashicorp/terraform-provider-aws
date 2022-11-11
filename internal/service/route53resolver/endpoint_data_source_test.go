package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverEndpointDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_endpoint.test"
	datasourceName := "data.aws_route53_resolver_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "direction", resourceName, "direction"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_addresses.#", resourceName, "ip_address.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "host_vpc_id"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverEndpointDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_endpoint.test"
	datasourceName := "data.aws_route53_resolver_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				Config: testAccEndpointDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "direction", resourceName, "direction"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_addresses.#", resourceName, "ip_address.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "host_vpc_id"),
				),
			},
		},
	})
}

func testAccEndpointDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_basic(rName), `
data "aws_route53_resolver_endpoint" "test" {
  resolver_endpoint_id = aws_route53_resolver_endpoint.test.id
}
`)
}

func testAccEndpointDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_outbound(rName, rName), `
data "aws_route53_resolver_endpoint" "test" {
  filter {
    name   = "Name"
    values = [aws_route53_resolver_endpoint.test.name]
  }

  filter {
    name   = "SecurityGroupIds"
    values = aws_security_group.test[*].id
  }
}
`)
}
