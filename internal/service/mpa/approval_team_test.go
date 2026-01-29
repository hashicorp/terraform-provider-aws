// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mpa_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mpa"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmpa "github.com/hashicorp/terraform-provider-aws/internal/service/mpa"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameApprovalTeam = "Approval Team"
)

func TestAccMPAApprovalTeam_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var approvalteam mpa.GetApprovalTeamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mpa_approval_team.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MPAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalTeamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalTeamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovalTeamExists(ctx, t, resourceName, &approvalteam),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test approval team"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mpa", regexache.MustCompile(`approval-team/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "approval_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_strategy.0.m_of_n.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_strategy.0.m_of_n.0.min_approvals_required", "1"),
					resource.TestCheckResourceAttr(resourceName, "approver.#", "1"),
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

func TestAccMPAApprovalTeam_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var approvalteam mpa.GetApprovalTeamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mpa_approval_team.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MPAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalTeamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalTeamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalTeamExists(ctx, t, resourceName, &approvalteam),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmpa.ResourceApprovalTeam, resourceName),
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

func TestAccMPAApprovalTeam_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var approvalteam mpa.GetApprovalTeamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mpa_approval_team.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MPAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovalTeamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalTeamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovalTeamExists(ctx, t, resourceName, &approvalteam),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test approval team"),
					resource.TestCheckResourceAttr(resourceName, "approval_strategy.0.m_of_n.0.min_approvals_required", "1"),
				),
			},
			{
				Config: testAccApprovalTeamConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovalTeamExists(ctx, t, resourceName, &approvalteam),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated approval team"),
					resource.TestCheckResourceAttr(resourceName, "approval_strategy.0.m_of_n.0.min_approvals_required", "2"),
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

func testAccCheckApprovalTeamDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MPAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mpa_approval_team" {
				continue
			}

			_, err := tfmpa.FindApprovalTeamByARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.MPA, create.ErrActionCheckingDestroyed, ResNameApprovalTeam, rs.Primary.ID, err)
			}

			return create.Error(names.MPA, create.ErrActionCheckingDestroyed, ResNameApprovalTeam, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApprovalTeamExists(ctx context.Context, t *testing.T, name string, approvalteam *mpa.GetApprovalTeamOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MPA, create.ErrActionCheckingExistence, ResNameApprovalTeam, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MPA, create.ErrActionCheckingExistence, ResNameApprovalTeam, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MPAClient(ctx)

		resp, err := tfmpa.FindApprovalTeamByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MPA, create.ErrActionCheckingExistence, ResNameApprovalTeam, rs.Primary.ID, err)
		}

		*approvalteam = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MPAClient(ctx)

	var input mpa.ListApprovalTeamsInput

	_, err := conn.ListApprovalTeams(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccApprovalTeamConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_mpa_approval_team" "test" {
  name        = %[1]q
  description = "Test approval team"

  approval_strategy {
    m_of_n {
      min_approvals_required = 1
    }
  }

  approver {
    primary_identity_id         = data.aws_caller_identity.current.user_id
    primary_identity_source_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
  }

  policy {
    policy_arn = "arn:${data.aws_partition.current.partition}:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:policy/example"
  }
}
`, rName)
}

func testAccApprovalTeamConfig_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_mpa_approval_team" "test" {
  name        = %[1]q
  description = "Updated approval team"

  approval_strategy {
    m_of_n {
      min_approvals_required = 2
    }
  }

  approver {
    primary_identity_id         = data.aws_caller_identity.current.user_id
    primary_identity_source_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
  }

  approver {
    primary_identity_id         = "second-approver"
    primary_identity_source_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
  }

  policy {
    policy_arn = "arn:${data.aws_partition.current.partition}:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:policy/example"
  }
}
`, rName)
}
