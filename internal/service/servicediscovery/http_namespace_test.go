// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceDiscoveryHTTPNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_http_namespace.test"
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHTTPNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHTTPNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "http_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccServiceDiscoveryHTTPNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_http_namespace.test"
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHTTPNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHTTPNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicediscovery.ResourceHTTPNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryHTTPNamespace_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_http_namespace.test"
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHTTPNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHTTPNamespaceConfig_description(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
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

func TestAccServiceDiscoveryHTTPNamespace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_http_namespace.test"
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHTTPNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHTTPNamespaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHTTPNamespaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccHTTPNamespaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHTTPNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckHTTPNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_service_discovery_http_namespace" {
				continue
			}

			_, err := tfservicediscovery.FindNamespaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Discovery HTTP Namespace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHTTPNamespaceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

		_, err := tfservicediscovery.FindNamespaceByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccHTTPNamespaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}
`, rName)
}

func testAccHTTPNamespaceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  description = %[1]q
  name        = %[2]q
}
`, description, rName)
}

func testAccHTTPNamespaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHTTPNamespaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
