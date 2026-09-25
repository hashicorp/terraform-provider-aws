// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIPRoutesSemanticEquals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	route := func(cidrIP, cidrIPv6, description string) *tfds.IPRouteModel {
		m := &tfds.IPRouteModel{
			CidrIP:      types.StringNull(),
			CidrIPv6:    types.StringNull(),
			Description: types.StringValue(description),
		}
		if cidrIP != "" {
			m.CidrIP = types.StringValue(cidrIP)
		}
		if cidrIPv6 != "" {
			m.CidrIPv6 = types.StringValue(cidrIPv6)
		}
		return m
	}

	set := func(routes ...*tfds.IPRouteModel) fwtypes.SetNestedObjectValueOf[tfds.IPRouteModel] {
		return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, routes)
	}

	testCases := map[string]struct {
		a, b fwtypes.SetNestedObjectValueOf[tfds.IPRouteModel]
		want bool
	}{
		"equivalent IPv6 spellings": {
			a:    set(route("", "2001:0db8::/64", "example")),
			b:    set(route("", "2001:db8::/64", "example")),
			want: true,
		},
		"identical IPv4": {
			a:    set(route("192.0.2.0/24", "", "example")),
			b:    set(route("192.0.2.0/24", "", "example")),
			want: true,
		},
		"order independent": {
			a:    set(route("192.0.2.0/24", "", "a"), route("", "2001:db8::/64", "b")),
			b:    set(route("", "2001:0db8::/64", "b"), route("192.0.2.0/24", "", "a")),
			want: true,
		},
		"different description": {
			a:    set(route("", "2001:db8::/64", "one")),
			b:    set(route("", "2001:db8::/64", "two")),
			want: false,
		},
		"different CIDR": {
			a:    set(route("", "2001:db8::/64", "example")),
			b:    set(route("", "2001:db8:1::/64", "example")),
			want: false,
		},
		"different count": {
			a:    set(route("192.0.2.0/24", "", "a")),
			b:    set(route("192.0.2.0/24", "", "a"), route("", "2001:db8::/64", "b")),
			want: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := tfds.IPRoutesSemanticEquals(ctx, testCase.a, testCase.b)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != testCase.want {
				t.Errorf("IPRoutesSemanticEquals = %v, want %v", got, testCase.want)
			}
		})
	}
}

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

// IPv6 IP routes (cidr_ipv6) are supported by the AddIpRoutes API and this
// resource, but they can only be added to a directory whose network type has
// been updated to dual-stack (IPv4 and IPv6). That update is a one-way,
// irreversible operation that is not exposed by aws_directory_service_directory,
// so an IPv6 route cannot be provisioned in a standard acceptance test (AWS
// returns "Invalid Directory network configuration for the requested CIDR IPs").
// IPv6 handling is covered by the TestIPRoutesSemanticEquals unit test instead.

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

		var cidrIPs, cidrIPv6s []string
		for _, route := range routes {
			if route.CidrIp != nil {
				cidrIPs = append(cidrIPs, aws.ToString(route.CidrIp))
			} else if route.CidrIpv6 != nil {
				cidrIPv6s = append(cidrIPv6s, aws.ToString(route.CidrIpv6))
			}
		}

		if _, err = conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     cidrIPs,
			CidrIpv6s:   cidrIPv6s,
		}); err != nil {
			return err
		}

		// RemoveIpRoutes is asynchronous. Wait for the routes to finish removing
		// so the backing directory does not begin teardown mid-removal.
		return tfds.WaitIPRoutesRemoved(ctx, conn, directoryID, slices.Concat(cidrIPs, cidrIPv6s), 30*time.Minute)
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
