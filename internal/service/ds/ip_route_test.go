// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIPRouteImportIDParse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		id      string
		want    map[string]any
		wantErr bool
	}{
		"IPv4": {
			id: "d-1234567890,192.0.2.0/24",
			want: map[string]any{
				"directory_id": "d-1234567890",
				"cidr_ip":      "192.0.2.0/24",
			},
		},
		"IPv6": {
			id: "d-1234567890,2001:db8::/64",
			want: map[string]any{
				"directory_id": "d-1234567890",
				"cidr_ipv6":    "2001:db8::/64",
			},
		},
		"no separator": {
			id:      "d-1234567890",
			wantErr: true,
		},
		"empty CIDR": {
			id:      "d-1234567890,",
			wantErr: true,
		},
		"empty directory": {
			id:      ",192.0.2.0/24",
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, got, err := tfds.IPRouteImportID{}.Parse(testCase.id)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !maps.Equal(got, testCase.want) {
				t.Errorf("Parse(%q) = %v, want %v", testCase.id, got, testCase.want)
			}
		})
	}
}

func TestIsIPRoutesUpdateRetryable(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err  error
		want bool
	}{
		"nil": {
			err: nil,
		},
		"update in progress": {
			err:  &awstypes.ClientException{Message: aws.String("An update is already in progress for directory d-1234567890. Please wait until this update is complete before making another change.")},
			want: true,
		},
		"under maintenance": {
			err:  &awstypes.ClientException{Message: aws.String("The directory is currently unavailable for updates. The directory d-1234567890 is under maintenance. Please retry the operation after a few minutes.")},
			want: true,
		},
		"other ClientException": {
			err: &awstypes.ClientException{Message: aws.String("Invalid Directory network configuration for the requested CIDR IPs")},
		},
		"other error type": {
			err: &awstypes.EntityDoesNotExistException{Message: aws.String("update is already in progress")},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := tfds.IsIPRoutesUpdateRetryable(testCase.err)
			if got != testCase.want {
				t.Errorf("IsIPRoutesUpdateRetryable() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAccDSIPRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRouteConfig_basic(rName, domainName, "example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRouteExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", "aws_directory_service_directory.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cidr_ip", "192.0.2.0/24"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "example"),
					resource.TestCheckResourceAttr(resourceName, "update_security_group_for_directory_controllers", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "directory_id", "cidr_ip"),
				ImportStateVerifyIdentifierAttribute: "directory_id",
			},
		},
	})
}

func TestAccDSIPRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRouteConfig_basic(rName, domainName, "example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRouteExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfds.ResourceIPRoute, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// Every argument forces replacement because there is no API to update a route.
func TestAccDSIPRoute_replace(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.IpRouteInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ip_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPRouteConfig_basic(rName, domainName, "example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRouteExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccIPRouteConfig_basic(rName, domainName, "updated"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRouteExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
				),
			},
			{
				// Exercises the mapping to AddIpRoutesInput.UpdateSecurityGroupForDirectoryControllers.
				Config: testAccIPRouteConfig_updateSecurityGroup(rName, domainName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPRouteExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "update_security_group_for_directory_controllers", acctest.CtTrue),
				),
			},
		},
	})
}

// IPv6 routes (cidr_ipv6) can only be added to a directory whose network type
// has been updated to dual-stack. That update is one-way and not exposed by
// aws_directory_service_directory, so it cannot be provisioned in a standard
// acceptance test (AWS returns "Invalid Directory network configuration for
// the requested CIDR IPs"). IPv6 import ID parsing is covered by
// TestIPRouteImportIDParse.

func testAccCheckIPRouteDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_ip_route" {
				continue
			}

			directoryID, cidr := rs.Primary.Attributes["directory_id"], testAccIPRouteKey(rs)
			_, err := tfds.FindIPRouteByTwoPartKey(ctx, conn, directoryID, cidr)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service IP Route %s,%s still exists", directoryID, cidr)
		}

		return nil
	}
}

func testAccCheckIPRouteExists(ctx context.Context, t *testing.T, n string, v *awstypes.IpRouteInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		output, err := tfds.FindIPRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], testAccIPRouteKey(rs))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIPRouteKey(rs *terraform.ResourceState) string {
	if v := rs.Primary.Attributes["cidr_ip"]; v != "" {
		return inttypes.CanonicalCIDRBlock(v)
	}
	return inttypes.CanonicalCIDRBlock(rs.Primary.Attributes["cidr_ipv6"])
}

func testAccIPRouteConfig_base(rName, domainName string) string {
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

func testAccIPRouteConfig_basic(rName, domainName, description string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), fmt.Sprintf(`
resource "aws_directory_service_ip_route" "test" {
  directory_id = aws_directory_service_directory.test.id
  cidr_ip      = "192.0.2.0/24"
  description  = %[1]q
}
`, description))
}

func testAccIPRouteConfig_updateSecurityGroup(rName, domainName string) string {
	return acctest.ConfigCompose(testAccIPRouteConfig_base(rName, domainName), `
resource "aws_directory_service_ip_route" "test" {
  directory_id = aws_directory_service_directory.test.id
  cidr_ip      = "192.0.2.0/24"
  description  = "updated"

  update_security_group_for_directory_controllers = true
}
`)
}
