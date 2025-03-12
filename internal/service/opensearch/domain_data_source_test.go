// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchDomainDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := testAccRandomDomainName()
	datasourceName := "data.aws_opensearch_domain.test"
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "auto_tune_options.#", resourceName, "auto_tune_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.#", resourceName, "cluster_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.#", resourceName, "ebs_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(datasourceName, "log_publishing_options.#", resourceName, "log_publishing_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "off_peak_window_options.#", resourceName, "off_peak_window_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.#", resourceName, "snapshot_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_options.#", resourceName, "vpc_options.#"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainDataSource_complex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := testAccRandomDomainName()
	datasourceName := "data.aws_opensearch_domain.test"
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig_complex(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "auto_tune_options.#", resourceName, "auto_tune_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.#", resourceName, "cluster_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_type", resourceName, "cluster_config.0.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_count", resourceName, "cluster_config.0.instance_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.dedicated_master_enabled", resourceName, "cluster_config.0.dedicated_master_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.zone_awareness_enabled", resourceName, "cluster_config.0.zone_awareness_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.#", resourceName, "ebs_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.ebs_enabled", resourceName, "ebs_options.0.ebs_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_size", resourceName, "ebs_options.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_type", resourceName, "ebs_options.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(datasourceName, "log_publishing_options.#", resourceName, "log_publishing_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "off_peak_window_options.#", resourceName, "off_peak_window_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.#", resourceName, "snapshot_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.0.automated_snapshot_start_hour", resourceName, "snapshot_options.0.automated_snapshot_start_hour"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_options.#", resourceName, "vpc_options.#"),
				),
			},
		},
	})
}

func testAccDomainDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), `
data "aws_opensearch_domain" "test" {
  domain_name = aws_opensearch_domain.test.domain_name
}
`)
}

func testAccDomainDataSourceConfig_complex(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_complex(rName), `
data "aws_opensearch_domain" "test" {
  domain_name = aws_opensearch_domain.test.domain_name
}
`)
}
