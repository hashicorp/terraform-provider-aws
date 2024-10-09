// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceDiscoveryPrivateDNSNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateDNSNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "servicediscovery", regexache.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPrivateDNSNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateDNSNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicediscovery.ResourcePrivateDNSNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateDNSNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_description(rName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc1"),
				),
			},
			{
				Config: testAccPrivateDNSNamespaceConfig_description(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc2"),
				),
			},
		},
	})
}

// This acceptance test ensures we properly send back error messaging. References:
//   - https://github.com/hashicorp/terraform-provider-aws/issues/2830
//   - https://github.com/hashicorp/terraform-provider-aws/issues/5532
func TestAccServiceDiscoveryPrivateDNSNamespace_Error_overlap(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateDNSNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPrivateDNSNamespaceConfig_overlapping(rName),
				ExpectError: regexache.MustCompile(`ConflictingDomainExists`),
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateDNSNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPrivateDNSNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccPrivateDNSNamespaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPrivateDNSNamespaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckPrivateDNSNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_service_discovery_private_dns_namespace" {
				continue
			}

			_, err := tfservicediscovery.FindNamespaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Discovery Private DNS Namespace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPrivateDNSNamespaceExists(ctx context.Context, n string) resource.TestCheckFunc {
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	input := &servicediscovery.ListNamespacesInput{}

	_, err := conn.ListNamespaces(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPrivateDNSNamespaceConfig_basic(rName string) string {
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
`, rName)
}

func testAccPrivateDNSNamespaceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[2]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  description = %[1]q
  name        = "%[2]s.test"
  vpc         = aws_vpc.test.id
}
`, description, rName)
}

func testAccPrivateDNSNamespaceConfig_overlapping(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "top" {
  name = "%[1]s.test"
  vpc  = aws_vpc.test.id
}

# Ensure ordering after first namespace
resource "aws_service_discovery_private_dns_namespace" "subdomain" {
  name = aws_service_discovery_private_dns_namespace.top.name
  vpc  = aws_service_discovery_private_dns_namespace.top.vpc
}
`, rName)
}

func testAccPrivateDNSNamespaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPrivateDNSNamespaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPrivateDNSNamespaceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.ID, rs.Primary.Attributes["vpc"]), nil
	}
}
