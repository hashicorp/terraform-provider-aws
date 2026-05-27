// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceDiscoveryService_private(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_private(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dns_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.routing_policy", "MULTIVALUE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.0.failure_threshold", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "DNS_HTTP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccServiceConfig_privateUpdate(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.ttl", "10"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.1.type", "AAAA"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.1.ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.0.failure_threshold", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "DNS_HTTP"),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryService_public(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_service_discovery_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_public(rName, 5, "/path"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.routing_policy", "WEIGHTED"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.resource_path", "/path"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccServiceConfig_public(rName, 3, "/updated-path"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.failure_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.resource_path", "/updated-path"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccServiceConfig_publicUpdateNoHealthCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dns_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryService_private_http(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_service_discovery_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_private_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccServiceDiscoveryService_http(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_service_discovery_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`service/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccServiceDiscoveryService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_service_discovery_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicediscovery.ResourceService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryService_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_service_discovery_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServiceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceDiscoveryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_service_discovery_service" {
				continue
			}

			_, err := tfservicediscovery.FindServiceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Discovery Service %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceDiscoveryClient(ctx)

		_, err := tfservicediscovery.FindServiceByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccServiceConfig_private(rName string, th int) string {
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
    failure_threshold = %[2]d
  }
}
`, rName, th)
}

func testAccServiceConfig_privateUpdate(rName string, th int) string {
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

  description = "test"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    dns_records {
      ttl  = 5
      type = "AAAA"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = %[2]d
  }
}
`, rName, th)
}

func testAccServiceConfig_private_http(rName string) string {
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

  namespace_id = aws_service_discovery_private_dns_namespace.test.id

  type = "HTTP"
}
`, rName)
}

func testAccServiceConfig_public(rName string, th int, path string) string {
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
    failure_threshold = %[2]d
    resource_path     = %[3]q
    type              = "HTTP"
  }
}
`, rName, th, path)
}

func testAccServiceConfig_publicUpdateNoHealthCheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.test"
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }

    routing_policy = "WEIGHTED"
  }
}
`, rName)
}

func testAccServiceConfig_http(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id
}
`, rName)
}

func testAccServiceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
