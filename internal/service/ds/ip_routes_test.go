// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSIPRoutes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRoutesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", "aws_directory_service_directory.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_route.*", map[string]string{
						"cidr_ip":             "192.0.2.0/24",
						names.AttrDescription: "example",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "directory_id"),
				ImportStateVerifyIdentifierAttribute: "directory_id",
				ImportStateVerifyIgnore:              []string{"update_security_group_for_directory_controllers"},
			},
		},
	})
}

func TestAccDSIPRoutes_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRoutesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
			{
				Config: testAccIPRoutesConfig_updated(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_route.*", map[string]string{
						"cidr_ip":             "192.0.2.0/24",
						names.AttrDescription: "updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_route.*", map[string]string{
						"cidr_ip":             "198.51.100.0/24",
						names.AttrDescription: "second",
					}),
				),
			},
			{
				// Toggling the flag forces replacement (it is only consumed by
				// AddIpRoutes) and exercises the mapping to
				// AddIpRoutesInput.UpdateSecurityGroupForDirectoryControllers.
				Config: testAccIPRoutesConfig_updateSecurityGroup(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "update_security_group_for_directory_controllers", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
		},
	})
}

func TestAccDSIPRoutes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRoutesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					testAccCheckIPRoutesDisappears(ctx, t, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSIPRoutes_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRoutesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesConfig_ipv6(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_route.*", map[string]string{
						"cidr_ip":             "192.0.2.0/24",
						names.AttrDescription: "ipv4",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_route.*", map[string]string{
						"cidr_ipv6":           "2001:db8::/64",
						names.AttrDescription: "ipv6",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "directory_id"),
				ImportStateVerifyIdentifierAttribute: "directory_id",
				ImportStateVerifyIgnore:              []string{"update_security_group_for_directory_controllers"},
			},
		},
	})
}

func testAccCheckIPRoutesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_ip_routes" {
				continue
			}

			_, err := tfds.FindIPRoutesByDirectoryID(ctx, conn, rs.Primary.Attributes["directory_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service IP Routes %s still exist", rs.Primary.Attributes["directory_id"])
		}

		return nil
	}
}

func testAccCheckIPRoutesExists(ctx context.Context, t *testing.T, n string, v *awstypes.IpRouteInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		output, err := tfds.FindIPRoutesByDirectoryID(ctx, conn, rs.Primary.Attributes["directory_id"])

		if err != nil {
			return err
		}

		*v = output[0]

		return nil
	}
}

func testAccCheckIPRoutesDisappears(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		directoryID := rs.Primary.Attributes["directory_id"]
		routes, err := tfds.FindIPRoutesByDirectoryID(ctx, conn, directoryID)
		if err != nil {
			return err
		}

		cidrs := make([]string, 0, len(routes))
		for _, route := range routes {
			cidrs = append(cidrs, aws.ToString(route.CidrIp))
		}

		_, err = conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     cidrs,
		})

		return err
	}
}

func testAccIPRoutesConfig_base(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domainName))
}

func testAccIPRoutesConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRoutesConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes" "test" {
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "example"
  }
}
`)
}

func testAccIPRoutesConfig_updated(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRoutesConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes" "test" {
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "updated"
  }

  ip_route {
    cidr_ip     = "198.51.100.0/24"
    description = "second"
  }
}
`)
}

func testAccIPRoutesConfig_updateSecurityGroup(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRoutesConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes" "test" {
  directory_id = aws_directory_service_directory.test.id

  update_security_group_for_directory_controllers = true

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "example"
  }
}
`)
}

func testAccIPRoutesConfig_ipv6(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_ip_routes" "test" {
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "ipv4"
  }

  ip_route {
    cidr_ipv6   = "2001:db8::/64"
    description = "ipv6"
  }
}
`, rName, domainName))
}
