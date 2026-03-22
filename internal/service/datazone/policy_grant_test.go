// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZonePolicyGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policygrant awstypes.PolicyGrantMember
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_policy_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGrantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyGrantExists(ctx, t, resourceName, &policygrant),
					resource.TestCheckResourceAttrSet(resourceName, "grant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, "entity_type", "DOMAIN_UNIT"),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "CREATE_DOMAIN_UNIT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyGrantImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataZonePolicyGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policygrant awstypes.PolicyGrantMember
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_policy_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGrantConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyGrantExists(ctx, t, resourceName, &policygrant),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourcePolicyGrant, resourceName),
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

func TestAccDataZonePolicyGrant_domainUnitPrincipal(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policygrant awstypes.PolicyGrantMember
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_policy_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGrantConfig_domainUnitPrincipal(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyGrantExists(ctx, t, resourceName, &policygrant),
					resource.TestCheckResourceAttrSet(resourceName, "grant_id"),
					resource.TestCheckResourceAttr(resourceName, "entity_type", "DOMAIN_UNIT"),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "CREATE_DOMAIN_UNIT"),
					resource.TestCheckResourceAttr(resourceName, "detail.0.create_domain_unit.0.include_child_domain_units", "true"),
					resource.TestCheckResourceAttr(resourceName, "principal.0.domain_unit.0.domain_unit_designation", "OWNER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyGrantImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPolicyGrantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_policy_grant" {
				continue
			}

			_, err := tfdatazone.FindPolicyGrantByID(ctx, conn,
				rs.Primary.Attributes["domain_identifier"],
				rs.Primary.Attributes["entity_type"],
				rs.Primary.Attributes["entity_identifier"],
				rs.Primary.Attributes["policy_type"],
				rs.Primary.Attributes["grant_id"],
			)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNamePolicyGrant, rs.Primary.Attributes["grant_id"], err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNamePolicyGrant, rs.Primary.Attributes["grant_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPolicyGrantExists(ctx context.Context, t *testing.T, name string, policygrant *awstypes.PolicyGrantMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNamePolicyGrant, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["grant_id"] == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNamePolicyGrant, name, errors.New("grant_id not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		resp, err := tfdatazone.FindPolicyGrantByID(ctx, conn,
			rs.Primary.Attributes["domain_identifier"],
			rs.Primary.Attributes["entity_type"],
			rs.Primary.Attributes["entity_identifier"],
			rs.Primary.Attributes["policy_type"],
			rs.Primary.Attributes["grant_id"],
		)
		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNamePolicyGrant, rs.Primary.Attributes["grant_id"], err)
		}

		*policygrant = *resp

		return nil
	}
}

func testAccPolicyGrantImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s,%s,%s",
			rs.Primary.Attributes["domain_identifier"],
			rs.Primary.Attributes["entity_type"],
			rs.Primary.Attributes["entity_identifier"],
			rs.Primary.Attributes["policy_type"],
			rs.Primary.Attributes["grant_id"],
		), nil
	}
}

func testAccPolicyGrantConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), `
resource "aws_datazone_policy_grant" "test" {
  domain_identifier = aws_datazone_domain.test.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.test.root_domain_unit_id
  policy_type       = "CREATE_DOMAIN_UNIT"

  detail {
    create_domain_unit {}
  }

  principal {
    domain_unit {
      domain_unit_designation = "OWNER"
      domain_unit_identifier  = aws_datazone_domain.test.root_domain_unit_id
    }
  }
}
`)
}

func testAccPolicyGrantConfig_domainUnitPrincipal(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), `
resource "aws_datazone_policy_grant" "test" {
  domain_identifier = aws_datazone_domain.test.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.test.root_domain_unit_id
  policy_type       = "CREATE_DOMAIN_UNIT"

  detail {
    create_domain_unit {
      include_child_domain_units = true
    }
  }

  principal {
    domain_unit {
      domain_unit_designation = "OWNER"
      domain_unit_identifier  = aws_datazone_domain.test.root_domain_unit_id
    }
  }
}
`)
}
