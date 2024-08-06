// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceDiscoveryServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_service_discovery_service.test"
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_config.#", resourceName, "dns_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_config.#", resourceName, "health_check_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_custom_config.#", resourceName, "health_check_custom_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_id", resourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryServiceDataSource_private(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_service_discovery_service.test"
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_private(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_config.#", resourceName, "dns_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_config.#", resourceName, "health_check_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_custom_config.#", resourceName, "health_check_custom_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_id", resourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryServiceDataSource_public(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_service_discovery_service.test"
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_public(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_config.#", resourceName, "dns_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_config.#", resourceName, "health_check_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_check_custom_config.#", resourceName, "health_check_custom_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_id", resourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_service_discovery_service" "test" {
  name         = aws_service_discovery_service.test.name
  namespace_id = aws_service_discovery_http_namespace.test.id
}
`, rName)
}

func testAccServiceDataSourceConfig_private(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.test"
  vpc  = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

data "aws_service_discovery_service" "test" {
  name         = aws_service_discovery_service.test.name
  namespace_id = aws_service_discovery_private_dns_namespace.test.id
}
`, rName)
}

func testAccServiceDataSourceConfig_public(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.test"
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  description = "test"

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }

    routing_policy = "WEIGHTED"
  }

  health_check_config {
    failure_threshold = 5
    resource_path     = "/path"
    type              = "HTTP"
  }
}

data "aws_service_discovery_service" "test" {
  name         = aws_service_discovery_service.test.name
  namespace_id = aws_service_discovery_public_dns_namespace.test.id
}
`, rName)
}
