// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var checkDeploymentARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`deployment:[0-9a-z]+$`))

func testAccDeploymentUpdateVariables(rName string, description any, isPublished any, policyIndexes []int, scopeIndex int) config.Variables {
	indexes := make([]config.Variable, 0, len(policyIndexes))
	for _, i := range policyIndexes {
		indexes = append(indexes, config.IntegerVariable(i))
	}

	variables := config.Variables{
		acctest.CtRName:  config.StringVariable(rName),
		"policy_indexes": config.ListVariable(indexes...),
		"scope_index":    config.IntegerVariable(scopeIndex),
	}
	if v, ok := description.(string); ok {
		variables[names.AttrDescription] = config.StringVariable(v)
	}
	if v, ok := isPublished.(bool); ok {
		variables["is_published"] = config.BoolVariable(v)
	}

	return variables
}

func TestAccNetworkSecurityManagerDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"
	policyResourceName := "aws_networksecuritymanager_policy.test"
	scopeResourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDeployment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkDeploymentARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"enable_cross_account_visibility": knownvalue.Bool(false),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arns"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValueCollection(resourceName, []tfjsonpath.Path{tfjsonpath.New("policy_arns")}, policyResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("scope_arn"), scopeResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNetworkSecurityManagerDeployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDeployment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceDeployment, resourceName),
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

// Updating description, policy_arns and scope_arn publishes a new version in place.
func TestAccNetworkSecurityManagerDeployment_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"
	policy1ResourceName := "aws_networksecuritymanager_policy.test[0]"
	scope1ResourceName := "aws_networksecuritymanager_scope.test[0]"
	scope2ResourceName := "aws_networksecuritymanager_scope.test[1]"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDeployment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables: testAccDeploymentUpdateVariables(rName, "initial", nil, []int{0}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arns"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValueCollection(resourceName, []tfjsonpath.Path{tfjsonpath.New("policy_arns")}, policy1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("scope_arn"), scope1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			// Add a second policy and move to another scope.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables: testAccDeploymentUpdateVariables(rName, "updated", nil, []int{0, 1}, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arns"), knownvalue.SetSizeExact(2)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("scope_arn"), scope2ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			// Remove a policy without changing anything else.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables: testAccDeploymentUpdateVariables(rName, "updated", nil, []int{0}, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arns"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValueCollection(resourceName, []tfjsonpath.Path{tfjsonpath.New("policy_arns")}, policy1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccNetworkSecurityManagerDeployment_descriptionRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDeployment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables: testAccDeploymentUpdateVariables(rName, "initial", nil, []int{0}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables: testAccDeploymentUpdateVariables(rName, nil, nil, []int{0}, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle. The deployment is created as a draft,
// published in place, and finally given a pending draft on top of the published
// version, so that destroy removes both.
func TestAccNetworkSecurityManagerDeployment_isPublished(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	step := func(isPublished bool, policyIndexes []int, action plancheck.ResourceActionType) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Deployment/update/"),
			ConfigVariables: testAccDeploymentUpdateVariables(rName, nil, isPublished, policyIndexes, 0),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckDeploymentExists(ctx, t, resourceName),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkDeploymentARN),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(isPublished)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arns"), knownvalue.SetSizeExact(len(policyIndexes))),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDeployment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Created as a draft.
			step(false, []int{0}, plancheck.ResourceActionCreate),
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Deployment/update/"),
				ConfigVariables:                      testAccDeploymentUpdateVariables(rName, nil, false, []int{0}, 0),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// DRAFT -> ACTIVE: the draft itself is published.
			step(true, []int{0}, plancheck.ResourceActionUpdate),
			// ACTIVE -> DRAFT: a draft is saved on top of the published version.
			step(false, []int{0, 1}, plancheck.ResourceActionUpdate),
		},
	})
}

func testAccCheckDeploymentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		_, err := tfnetworksecuritymanager.FindDeploymentByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckDeploymentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_deployment" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
			_, err := tfnetworksecuritymanager.FindDeploymentByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Deployment %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccPreCheckDeployment(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	var input networksecuritymanager.ListDeploymentsInput
	_, err := conn.ListDeployments(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
