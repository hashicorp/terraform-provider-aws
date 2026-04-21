// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaPartitionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_partition.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPartitionDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(dataSourceName, "partition", func(value string) error {
						expected := acctest.ProviderMeta(ctx, t).Partition(ctx)
						if value != expected {
							return fmt.Errorf("Incorrect Partition: expected %q, got %q", expected, value)
						}

						return nil
					}),
					resource.TestCheckResourceAttrWith(dataSourceName, "dns_suffix", func(value string) error {
						expected := acctest.ProviderMeta(ctx, t).DNSSuffix(ctx)
						if value != expected {
							return fmt.Errorf("Incorrect DNS Suffix: expected %q, got %q", expected, value)
						}

						if value == "" {
							return fmt.Errorf("DNS Suffix expected to not be nil")
						}

						return nil
					}),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", acctest.PartitionReverseDNSPrefix()),
				),
			},
		},
	})
}

const testAccPartitionDataSourceConfig_basic = `
data "aws_partition" "test" {}
`
