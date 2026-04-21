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

func TestAccServiceDiscoveryPublicDNSNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicDNSNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPublicDNSNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicDNSNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicediscovery.ResourcePublicDNSNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPublicDNSNamespace_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicDNSNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicDNSNamespaceConfig_description(rName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc1"),
				),
			},
			{
				Config: testAccPublicDNSNamespaceConfig_description(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc2"),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryPublicDNSNamespace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicDNSNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicDNSNamespaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPublicDNSNamespaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPublicDNSNamespaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckPublicDNSNamespaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceDiscoveryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_service_discovery_public_dns_namespace" {
				continue
			}

			_, err := tfservicediscovery.FindNamespaceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Discovery Public DNS Namespace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPublicDNSNamespaceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceDiscoveryClient(ctx)

		_, err := tfservicediscovery.FindNamespaceByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPublicDNSNamespaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.test"
}
`, rName)
}

func testAccPublicDNSNamespaceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  description = %[1]q
  name        = "%[2]s.test"
}
`, description, rName)
}

func testAccPublicDNSNamespaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.test"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPublicDNSNamespaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.test"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
