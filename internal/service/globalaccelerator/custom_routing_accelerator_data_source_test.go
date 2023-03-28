package globalaccelerator_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlobalAcceleratorCustomRoutingAcceleratorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	dataSourceName := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_arn"
	dataSourceName2 := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorDataSourceConfig_basic(resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
				),
			},
		},
	})

}

func testAccCustomRoutingAcceleratorDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

data "aws_globalaccelerator_custom_routing_accelerator" "test_by_arn" {
  arn = aws_globalaccelerator_custom_routing_accelerator.test.id
}
  
data "aws_globalaccelerator_custom_routing_accelerator" "test_by_name" {
  name = aws_globalaccelerator_custom_routing_accelerator.test.name
}
`, rName)
}
