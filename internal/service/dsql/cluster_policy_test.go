// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdsql "github.com/hashicorp/terraform-provider-aws/internal/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLClusterPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterPolicy dsql.GetClusterPolicyOutput
	var initialPolicyVersion string
	resourceName := "aws_dsql_cluster_policy.test"
	clusterResourceName := "aws_dsql_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterPolicy),
					testAccCheckClusterPolicyRemoteAction(ctx, t, resourceName, "dsql:DbConnect", true),
					testAccCheckClusterPolicyRemoteAction(ctx, t, resourceName, "dsql:DbConnectAdmin", true),
					testAccCheckClusterPolicyVersionSet(t, resourceName, &initialPolicyVersion),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIdentifier, clusterResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_version"),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "identifier",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "identifier"),
				ImportStateVerifyIgnore:              []string{"bypass_policy_lockout_safety_check", names.AttrPolicy},
			},
			{
				Config: testAccClusterPolicyConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterPolicy),
					testAccCheckClusterPolicyRemoteAction(ctx, t, resourceName, "dsql:DbConnect", true),
					testAccCheckClusterPolicyRemoteAction(ctx, t, resourceName, "dsql:DbConnectAdmin", false),
					testAccCheckClusterPolicyVersionChanged(t, resourceName, &initialPolicyVersion),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDSQLClusterPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterPolicy dsql.GetClusterPolicyOutput
	resourceName := "aws_dsql_cluster_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterPolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdsql.ResourceClusterPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDSQLClusterPolicy_bypassPolicyLockoutSafetyCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterPolicy dsql.GetClusterPolicyOutput
	resourceName := "aws_dsql_cluster_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_bypassPolicyLockoutSafetyCheck(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterPolicy),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterPolicyConfig_bypassPolicyLockoutSafetyCheck(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterPolicy),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckClusterPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dsql_cluster_policy" {
				continue
			}
			id := rs.Primary.Attributes["identifier"]

			_, err := tfdsql.FindClusterPolicyByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Aurora DSQL Cluster Policy %s still exists", id)
		}

		return nil
	}
}

func testAccCheckClusterPolicyExists(ctx context.Context, t *testing.T, n string, v *dsql.GetClusterPolicyOutput) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		id := rs.Primary.Attributes["identifier"]

		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)
		output, err := tfdsql.FindClusterPolicyByID(ctx, conn, id)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterPolicyRemoteAction(ctx context.Context, t *testing.T, n, action string, expected bool) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		id := rs.Primary.Attributes["identifier"]

		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)
		output, err := tfdsql.FindClusterPolicyByID(ctx, conn, id)
		if err != nil {
			return err
		}

		contains, err := testAccClusterPolicyContainsAction(aws.ToString(output.Policy), action)
		if err != nil {
			return err
		}

		if contains != expected {
			return fmt.Errorf("expected Aurora DSQL Cluster Policy %s action %q presence to be %t, got %t", n, action, expected, contains)
		}

		return nil
	}
}

func testAccCheckClusterPolicyVersionSet(t *testing.T, n string, policyVersion *string) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		v := rs.Primary.Attributes["policy_version"]
		if v == "" {
			return fmt.Errorf("expected Aurora DSQL Cluster Policy %s policy version to be set", n)
		}

		*policyVersion = v

		return nil
	}
}

func testAccCheckClusterPolicyVersionChanged(t *testing.T, n string, previousPolicyVersion *string) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		v := rs.Primary.Attributes["policy_version"]
		if v == "" {
			return fmt.Errorf("expected Aurora DSQL Cluster Policy %s policy version to be set", n)
		}

		if v == *previousPolicyVersion {
			return fmt.Errorf("expected Aurora DSQL Cluster Policy %s policy version to change from %q", n, *previousPolicyVersion)
		}

		return nil
	}
}

func testAccClusterPolicyContainsAction(policy string, action string) (bool, error) {
	var doc struct {
		Statement []struct {
			Action any `json:"Action"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(policy), &doc); err != nil {
		return false, fmt.Errorf("unmarshaling Aurora DSQL Cluster Policy: %w", err)
	}

	for _, statement := range doc.Statement {
		switch actions := statement.Action.(type) {
		case string:
			if actions == action {
				return true, nil
			}
		case []any:
			for _, v := range actions {
				if v == action {
					return true, nil
				}
			}
		default:
			return false, fmt.Errorf("unexpected Aurora DSQL Cluster Policy action type %T", statement.Action)
		}
	}

	return false, nil
}

func testAccClusterPolicyConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_dsql_cluster" "test" {}

resource "aws_dsql_cluster_policy" "test" {
  identifier = aws_dsql_cluster.test.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCurrentAccountConnect"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.account_id
        }
        Action = [
          "dsql:DbConnect",
          "dsql:DbConnectAdmin",
        ]
        Resource = aws_dsql_cluster.test.arn
      }
    ]
  })
}
`
}

func testAccClusterPolicyConfig_updated() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_dsql_cluster" "test" {}

resource "aws_dsql_cluster_policy" "test" {
  identifier = aws_dsql_cluster.test.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCurrentAccountConnect"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.account_id
        }
        Action = [
          "dsql:DbConnect",
        ]
        Resource = aws_dsql_cluster.test.arn
      }
    ]
  })
}
`
}

func testAccClusterPolicyConfig_bypassPolicyLockoutSafetyCheck(bypass bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_dsql_cluster" "test" {}

resource "aws_dsql_cluster_policy" "test" {
  bypass_policy_lockout_safety_check = %[1]t
  identifier                         = aws_dsql_cluster.test.identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCurrentAccountConnect"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.account_id
        }
        Action = [
          "dsql:DbConnect",
        ]
        Resource = aws_dsql_cluster.test.arn
      }
    ]
  })
}
`, bypass)
}
