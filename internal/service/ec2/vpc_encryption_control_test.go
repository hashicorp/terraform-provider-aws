// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	exclusionAttributes = []string{
		"egress_only_internet_gateway_exclusion",
		"elastic_file_system_exclusion",
		"internet_gateway_exclusion",
		"lambda_exclusion",
		"nat_gateway_exclusion",
		"virtual_private_gateway_exclusion",
		"vpc_lattice_exclusion",
		"vpc_peering_exclusion",
	}

	defaultMonitorExclusionsImportStateCheck = acctest.ComposeAggregateImportStateCheckFunc(
		acctest.ImportCheckNoResourceAttr("egress_only_internet_gateway_exclusion"),
		acctest.ImportCheckNoResourceAttr("elastic_file_system_exclusion"),
		acctest.ImportCheckNoResourceAttr("internet_gateway_exclusion"),
		acctest.ImportCheckNoResourceAttr("lambda_exclusion"),
		acctest.ImportCheckNoResourceAttr("nat_gateway_exclusion"),
		acctest.ImportCheckNoResourceAttr("virtual_private_gateway_exclusion"),
		acctest.ImportCheckNoResourceAttr("vpc_lattice_exclusion"),
		acctest.ImportCheckNoResourceAttr("vpc_peering_exclusion"),
	)
	defaultEnforceExclusionsImportStateCheck = acctest.ComposeAggregateImportStateCheckFunc(
		acctest.ImportCheckResourceAttr("egress_only_internet_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("elastic_file_system_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("internet_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("lambda_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("nat_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("virtual_private_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("vpc_lattice_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
		acctest.ImportCheckResourceAttr("vpc_peering_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
	)
)

func TestAccVPCVPCEncryptionControl_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrVPCID), "aws_vpc.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("elastic_file_system_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("lambda_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_lattice_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultMonitorExclusionsImportStateCheck,
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceVPCEncryptionControl, resourceName),
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

func TestAccVPCVPCEncryptionControl_enforce(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrVPCID), "aws_vpc.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("elastic_file_system_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("lambda_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_lattice_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultEnforceExclusionsImportStateCheck,
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_enforce_ImplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ImplicitExclusions("egress_only_internet_gateway_exclusion", "lambda_exclusion", "nat_gateway_exclusion", "vpc_peering_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"internet_gateway_exclusion",
					"virtual_private_gateway_exclusion",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr("internet_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					acctest.ImportCheckResourceAttr("virtual_private_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
				),
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_enforce_ExplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ExplicitExclusions("internet_gateway_exclusion", "lambda_exclusion", "virtual_private_gateway_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_update_monitorToEnforce(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrVPCID), "aws_vpc.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrVPCID), "aws_vpc.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultEnforceExclusionsImportStateCheck,
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_update_monitorToEnforce_ImplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrVPCID), "aws_vpc.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ImplicitExclusions("egress_only_internet_gateway_exclusion", "lambda_exclusion", "nat_gateway_exclusion", "vpc_peering_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_update_enforce_ExplictDisableExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ExplictDisableExclusion(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_update_enforce_ImplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ImplicitExclusions("egress_only_internet_gateway_exclusion", "lambda_exclusion", "nat_gateway_exclusion", "vpc_peering_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"internet_gateway_exclusion",
					"virtual_private_gateway_exclusion",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr("internet_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					acctest.ImportCheckResourceAttr("virtual_private_gateway_exclusion", string(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
				),
			},
		},
	})
}

func TestAccVPCVPCEncryptionControl_update_enforceToMonitor(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultMonitorExclusionsImportStateCheck,
			},
		},
	})
}

// Test for associated resources that can be excluded, such as Internet Gateway
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Excludable_monitor(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	var vpc awstypes.Vpc
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultMonitorExclusionsImportStateCheck,
			},
		},
	})
}

// Test for associated resources that can be excluded, such as Internet Gateway
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Excludable_enforceWithoutExclusion(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	var vpc awstypes.Vpc
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config:      testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_enable(awstypes.VpcEncryptionControlModeEnforce),
				ExpectError: regexache.MustCompile(`The following resources prevented enforcement:`),
			},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),

					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(awstypes.VpcEncryptionControlModeMonitor)),
				),
				// RefreshPlanChecks: resource.RefreshPlanChecks{
				// 	PostRefresh: []plancheck.PlanCheck{
				// 		plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				// 	},
				// },
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Test for associated resources that can be excluded, such as Internet Gateway
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Excludable_enforceWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)

	var vpc awstypes.Vpc
	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_enableWithExclusions(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"elastic_file_system": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"lambda": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_lattice": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrState: tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
					})),
				},
			},
		},
	})
}

// Test for associated resources that can be migrated to encrypted hardward, such as load balancer
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Migratable_monitor(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	var vpc awstypes.Vpc
	resourceName := "aws_vpc_encryption_control.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_enable(rName, awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultMonitorExclusionsImportStateCheck,
			},
		},
	})
}

// Test for associated resources that can be migrated to encrypted hardward, such as load balancer
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Migratable_enforce(t *testing.T) {
	ctx := acctest.Context(t)

	var vpc awstypes.Vpc
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_enable(rName, awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultEnforceExclusionsImportStateCheck,
			},
		},
	})
}

// Test for associated resources that are not supported, such as RDS
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Unsupported_monitor(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	var vpc awstypes.Vpc
	resourceName := "aws_vpc_encryption_control.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_enable(rName, awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultMonitorExclusionsImportStateCheck,
			},
		},
	})
}

// Test for associated resources that are not supported, such as RDS
func TestAccVPCVPCEncryptionControl_WithAssociatedResources_Unsupported_enforce(t *testing.T) {
	ctx := acctest.Context(t)

	var vpc awstypes.Vpc
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, "aws_vpc.test", &vpc),
					testAccCheckVPCEncryptionControlDoesNotExist(ctx, t, resourceName),
				),
			},
			{
				Config:      testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_enable(rName, awstypes.VpcEncryptionControlModeEnforce),
				ExpectError: regexache.MustCompile(`The following resources prevented enforcement:`),
			},
		},
	})
}

// When in `enforce` mode, import behavior is different
func TestAccVPCVPCEncryptionControl_Identity_Enforce(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Import command
			{
				Config:                  testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				ImportStateKind:         resource.ImportCommandWithID,
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: exclusionAttributes,
				ImportStateCheck:        defaultEnforceExclusionsImportStateCheck,
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func testAccCheckVPCEncryptionControlDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_encryption_control" {
				continue
			}

			_, err := tfec2.FindVPCEncryptionControlByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCEncryptionControl, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCEncryptionControl, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVPCEncryptionControlExists(ctx context.Context, t *testing.T, name string, vpcencryptioncontrol *awstypes.VpcEncryptionControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		resp, err := tfec2.FindVPCEncryptionControlByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, rs.Primary.ID, err)
		}

		*vpcencryptioncontrol = *resp

		return nil
	}
}

func testAccCheckVPCEncryptionControlDoesNotExist(_ context.Context, _ *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, name, errors.New("found"))
		}

		return nil
	}
}

func testAccVPCEncryptionControlConfig_enable(mode awstypes.VpcEncryptionControlMode) string {
	return fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, mode)
}

func testAccVPCEncryptionControlConfig_enforce_ImplicitExclusions(keys ...string) string {
	var buf strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&buf, `%[1]s = "enable"`+"\n", key)
	}
	return fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  %[2]s
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, awstypes.VpcEncryptionControlModeEnforce, buf.String())
}

func testAccVPCEncryptionControlConfig_enforce_ExplicitExclusions(keys ...string) string {
	return fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  %[2]s
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, awstypes.VpcEncryptionControlModeEnforce, testAccVPCEncryptionControlConfig_explicitExclusions(keys...))
}

func testAccVPCEncryptionControlConfig_enforce_ExplictDisableExclusion() string {
	return testAccVPCEncryptionControlConfig_enforce_ExplicitExclusions()
}

func testAccVPCEncryptionControlConfig_explicitExclusions(keys ...string) string {
	var buf strings.Builder
	for _, key := range exclusionAttributes {
		var v string
		if slices.Contains(keys, key) {
			v = "enable"
		} else {
			v = "disable"
		}
		fmt.Fprintf(&buf, `%[1]s = %[2]q`+"\n", key, v)
	}
	return buf.String()
}

// Test for associated resources that can be excluded, such as Internet Gateway
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`
}

// Test for associated resources that can be excluded, such as Internet Gateway
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_enable(mode awstypes.VpcEncryptionControlMode) string {
	return acctest.ConfigCompose(
		testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup(),
		fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  depends_on = [
    aws_internet_gateway.test,
  ]
}
`, mode))
}

// Test for associated resources that can be excluded, such as Internet Gateway
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_enableWithExclusions(mode awstypes.VpcEncryptionControlMode) string {
	return acctest.ConfigCompose(
		testAccVPCEncryptionControlConfig_WithAssociatedResources_Excludable_setup(),
		fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  internet_gateway_exclusion = "enable"

  depends_on = [
    aws_internet_gateway.test,
  ]
}
`, mode))
}

// Test for associated resources that can be migrated to encrypted hardward, such as load balancer
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_setup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}
`, rName))
}

// Test for associated resources that can be migrated to encrypted hardward, such as load balancer
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_enable(rName string, mode awstypes.VpcEncryptionControlMode) string {
	return acctest.ConfigCompose(
		testAccVPCEncryptionControlConfig_WithAssociatedResources_Migratable_setup(rName),
		fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  depends_on = [
    aws_lb.test,
  ]
}
`, mode))
}

// Test for associated resources that are not supported, such as RDS
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_setup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigSubnets(rName, 2),
		acctest.ConfigRandomPassword(),
		testAccInstanceConfig_orderableClassPostgres(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_db_instance" "test" {
  identifier = %[1]q

  engine         = data.aws_rds_orderable_db_instance.test.engine
  engine_version = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class = data.aws_rds_orderable_db_instance.test.instance_class

  password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version = 1
  username            = "tfacctest"

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  db_subnet_group_name = aws_db_subnet_group.test.name

  allocated_storage   = 10
  skip_final_snapshot = true

  backup_retention_period = 0
  apply_immediately       = true
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

// Test for associated resources that are not supported, such as RDS
func testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_enable(rName string, mode awstypes.VpcEncryptionControlMode) string {
	return acctest.ConfigCompose(
		testAccVPCEncryptionControlConfig_WithAssociatedResources_Unsupported_setup(rName),
		fmt.Sprintf(`
resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = %[1]q

  depends_on = [
    aws_db_instance.test,
  ]
}
`, mode))
}

func testAccInstanceConfig_orderableClassPostgres() string {
	return testAccInstanceConfig_orderableClass(rds.InstanceEnginePostgres, "postgresql-license", "standard")
}

func testAccInstanceConfig_orderableClass(engine, license, storage string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = ["db.t4g.small", "db.t4g.medium", "db.t3.small", "db.t3.medium"]
}
`, engine, license, storage)
}
