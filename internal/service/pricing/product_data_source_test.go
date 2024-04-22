// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pricing_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPricingProductDataSource_ec2(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_pricing_product.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApSouth1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_ec2,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrIsJSONString(dataSourceName, "result"),
				),
			},
		},
	})
}

func TestAccPricingProductDataSource_redshift(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_pricing_product.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApSouth1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PricingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_redshift,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrIsJSONString(dataSourceName, "result"),
				),
			},
		},
	})
}

const testAccProductDataSourceConfig_ec2 = `
data "aws_ec2_instance_type_offering" "available" {
  preferred_instance_types = ["c5.large", "c4.large"]
}

data "aws_region" "current" {}

data "aws_pricing_product" "test" {
  service_code = "AmazonEC2"

  filters {
    field = "instanceType"
    value = data.aws_ec2_instance_type_offering.available.instance_type
  }

  filters {
    field = "operatingSystem"
    value = "Linux"
  }

  filters {
    field = "location"
    value = data.aws_region.current.description
  }

  filters {
    field = "preInstalledSw"
    value = "NA"
  }

  filters {
    field = "licenseModel"
    value = "No License required"
  }

  filters {
    field = "tenancy"
    value = "Shared"
  }

  filters {
    field = "capacitystatus"
    value = "Used"
  }
}
`

const testAccProductDataSourceConfig_redshift = `
data "aws_redshift_orderable_cluster" "test" {
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}

data "aws_region" "current" {}

data "aws_pricing_product" "test" {
  service_code = "AmazonRedshift"

  filters {
    field = "instanceType"
    value = data.aws_redshift_orderable_cluster.test.node_type
  }

  filters {
    field = "location"
    value = data.aws_region.current.description
  }

  filters {
    field = "productFamily"
    value = "Compute Instance"
  }
}
`
