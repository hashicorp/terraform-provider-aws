// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
		"both empty": {
			a:    set(),
			b:    set(),
			want: true,
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

func TestAccDSIPRoutesExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes_exclusive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveExists(ctx, t, resourceName),
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
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

func TestAccDSIPRoutesExclusive_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes_exclusive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
			{
				Config: testAccIPRoutesExclusiveConfig_updated(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 2),
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
				// The flag only affects routes added later, so toggling it is an
				// in-place update rather than a replacement.
				Config: testAccIPRoutesExclusiveConfig_updateSecurityGroup(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "update_security_group_for_directory_controllers", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
		},
	})
}

// An empty configuration removes all routes, including one managed by
// aws_directory_service_ip_route, which then shows as needing re-creation.
func TestAccDSIPRoutesExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes_exclusive.test"
	routeResourceName := "aws_directory_service_ip_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesExclusiveConfig_empty(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 0),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_route"), knownvalue.SetExact([]knownvalue.Check{})),
				},
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccIPRoutesExclusiveConfig_empty(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(routeResourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSIPRoutesExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes_exclusive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					testAccCheckIPRoutesExclusiveAddOutOfBand(ctx, t, resourceName, "198.51.100.0/24"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
		},
	})
}

func TestAccDSIPRoutesExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_routes_exclusive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					testAccCheckIPRoutesExclusiveRemoveOutOfBand(ctx, t, resourceName, "192.0.2.0/24"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccIPRoutesExclusiveConfig_basic(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRoutesExclusiveCount(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "ip_route.#", "1"),
				),
			},
		},
	})
}

func testAccCheckIPRoutesExclusiveExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		_, err := tfds.FindDirectoryByID(ctx, conn, rs.Primary.Attributes["directory_id"])

		return err
	}
}

func testAccCheckIPRoutesExclusiveCount(ctx context.Context, t *testing.T, n string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		routes, err := tfds.FindIPRoutesByDirectoryID(ctx, conn, rs.Primary.Attributes["directory_id"])
		if err != nil {
			return err
		}

		if got := len(routes); got != want {
			return fmt.Errorf("Directory Service IP Routes count = %d, want %d", got, want)
		}

		return nil
	}
}

func testAccCheckIPRoutesExclusiveAddOutOfBand(ctx context.Context, t *testing.T, n, cidr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)
		directoryID := rs.Primary.Attributes["directory_id"]

		input := directoryservice.AddIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			IpRoutes:    []awstypes.IpRoute{{CidrIp: aws.String(cidr)}},
		}
		if _, err := conn.AddIpRoutes(ctx, &input); err != nil {
			return err
		}

		return tfds.WaitIPRoutesAdded(ctx, conn, directoryID, []string{cidr}, 30*time.Minute)
	}
}

func testAccCheckIPRoutesExclusiveRemoveOutOfBand(ctx context.Context, t *testing.T, n, cidr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)
		directoryID := rs.Primary.Attributes["directory_id"]

		input := directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     []string{cidr},
		}
		if _, err := conn.RemoveIpRoutes(ctx, &input); err != nil {
			return err
		}

		return tfds.WaitIPRoutesRemoved(ctx, conn, directoryID, []string{cidr}, 30*time.Minute)
	}
}

func testAccIPRoutesExclusiveConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes_exclusive" "test" {
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "example"
  }
}
`)
}

func testAccIPRoutesExclusiveConfig_updated(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes_exclusive" "test" {
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

func testAccIPRoutesExclusiveConfig_updateSecurityGroup(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), `
resource "aws_directory_service_ip_routes_exclusive" "test" {
  directory_id = aws_directory_service_directory.test.id

  update_security_group_for_directory_controllers = true

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = "updated"
  }
}
`)
}

func testAccIPRoutesExclusiveConfig_empty(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), `
# Created first so the exclusive resource removes it.
resource "aws_directory_service_ip_route" "test" {
  directory_id = aws_directory_service_directory.test.id
  cidr_ip      = "192.0.2.0/24"
}

resource "aws_directory_service_ip_routes_exclusive" "test" {
  directory_id = aws_directory_service_directory.test.id

  depends_on = [aws_directory_service_ip_route.test]
}
`)
}
