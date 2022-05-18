package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaPartitionDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartition("data.aws_partition.current"),
					testAccCheckDNSSuffix("data.aws_partition.current"),
					resource.TestCheckResourceAttr("data.aws_partition.current", "reverse_dns_prefix", acctest.PartitionReverseDNSPrefix()),
				),
			},
		},
	})
}

func testAccCheckPartition(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find resource: %s", n)
		}

		expected := acctest.Provider.Meta().(*conns.AWSClient).Partition
		if rs.Primary.Attributes["partition"] != expected {
			return fmt.Errorf("Incorrect Partition: expected %q, got %q", expected, rs.Primary.Attributes["partition"])
		}

		return nil
	}
}

func testAccCheckDNSSuffix(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find resource: %s", n)
		}

		expected := acctest.Provider.Meta().(*conns.AWSClient).DNSSuffix
		if rs.Primary.Attributes["dns_suffix"] != expected {
			return fmt.Errorf("Incorrect DNS Suffix: expected %q, got %q", expected, rs.Primary.Attributes["dns_suffix"])
		}

		if rs.Primary.Attributes["dns_suffix"] == "" {
			return fmt.Errorf("DNS Suffix expected to not be nil")
		}

		return nil
	}
}

const testAccPartitionConfig_basic = `
data "aws_partition" "current" {}
`
