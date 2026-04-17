// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53DelegationSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	refName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_basic(refName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(ctx, t, resourceName),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "route53", regexache.MustCompile("delegationset/.+")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func TestAccRoute53DelegationSet_withZones(t *testing.T) {
	ctx := acctest.Context(t)
	var zone route53.GetHostedZoneOutput

	refName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"
	primaryZoneResourceName := "aws_route53_zone.primary"
	secondaryZoneResourceName := "aws_route53_zone.secondary"

	domain := acctest.RandomDomainName()
	zoneName1 := fmt.Sprintf("primary.%s", domain)
	zoneName2 := fmt.Sprintf("secondary.%s", domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_zones(refName, zoneName1, zoneName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(ctx, t, resourceName),
					testAccCheckZoneExists(ctx, t, primaryZoneResourceName, &zone),
					testAccCheckZoneExists(ctx, t, secondaryZoneResourceName, &zone),
					testAccCheckNameServersMatch(ctx, t, resourceName, primaryZoneResourceName),
					testAccCheckNameServersMatch(ctx, t, resourceName, secondaryZoneResourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func TestAccRoute53DelegationSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	refName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_delegation_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSetConfig_basic(refName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSetExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53.ResourceDelegationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegationSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_delegation_set" {
				continue
			}

			_, err := tfroute53.FindDelegationSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Reusable Delegation Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDelegationSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		_, err := tfroute53.FindDelegationSetByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckNameServersMatch(ctx context.Context, t *testing.T, delegationSetResourceName, hostedZoneResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsDelegationSet, ok := s.RootModule().Resources[delegationSetResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", delegationSetResourceName)
		}
		rsHostedZone, ok := s.RootModule().Resources[hostedZoneResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", hostedZoneResourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		delegationSet, err := tfroute53.FindDelegationSetByID(ctx, conn, rsDelegationSet.Primary.ID)

		if err != nil {
			return err
		}

		hostedZone, err := tfroute53.FindHostedZoneByID(ctx, conn, rsHostedZone.Primary.ID)

		if err != nil {
			return err
		}

		if !reflect.DeepEqual(delegationSet.NameServers, hostedZone.DelegationSet.NameServers) {
			return fmt.Errorf("Name servers do not match:\nDelegation Set: %#v\nHosted Zone:%#v",
				delegationSet.NameServers, hostedZone.DelegationSet.NameServers)
		}

		return nil
	}
}

func testAccDelegationSetConfig_basic(refName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
  reference_name = %[1]q
}
`, refName)
}

func testAccDelegationSetConfig_zones(refName, zoneName1, zoneName2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
  reference_name = %[1]q
}

resource "aws_route53_zone" "primary" {
  name              = %[2]q
  delegation_set_id = aws_route53_delegation_set.test.id
}

resource "aws_route53_zone" "secondary" {
  name              = %[3]q
  delegation_set_id = aws_route53_delegation_set.test.id
}
`, refName, zoneName1, zoneName2)
}
