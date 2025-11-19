// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	set "github.com/hashicorp/go-set/v3"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCVPCEncryptionControl_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.Null()),
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

func TestAccVPCVPCEncryptionControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCEncryptionControl, resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
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

func TestAccVPCVPCEncryptionControl_enforce_ImplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ImplicitExclusions("egress_only_internet_gateway_exclusion", "nat_gateway_exclusion", "vpc_peering_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
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

func TestAccVPCVPCEncryptionControl_enforce_ExplicitExclusions(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enforce_ExplicitExclusions("internet_gateway_exclusion", "virtual_private_gateway_exclusion"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("egress_only_internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("virtual_private_gateway_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_peering_exclusion"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateInputDisable)),

					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_exclusions"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"egress_only_internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"internet_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"nat_gateway_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"virtual_private_gateway": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateEnabled),
							"state_message": tfknownvalue.StringExact("succeeded"),
						}),
						"vpc_peering": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"state":         tfknownvalue.StringExact(awstypes.VpcEncryptionControlExclusionStateDisabled),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),
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

func TestAccVPCVPCEncryptionControl_update_enforceToMonitor(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.VpcEncryptionControl
	resourceName := "aws_vpc_encryption_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEncryptionControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeEnforce),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),
				},
			},
			{
				Config: testAccVPCEncryptionControlConfig_enable(awstypes.VpcEncryptionControlModeMonitor),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEncryptionControlExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mode"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state"), tfknownvalue.StringExact(awstypes.VpcEncryptionControlStateAvailable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("state_message"), tfknownvalue.StringExact("succeeded")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("vpc_id"), "aws_vpc.test", tfjsonpath.New("id"), compare.ValuesSame()),
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

func testAccCheckVPCEncryptionControlDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

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

func testAccCheckVPCEncryptionControlExists(ctx context.Context, name string, vpcencryptioncontrol *awstypes.VpcEncryptionControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVPCEncryptionControlByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCEncryptionControl, rs.Primary.ID, err)
		}

		*vpcencryptioncontrol = resp

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
	allKeys := set.From([]string{
		"egress_only_internet_gateway_exclusion",
		"internet_gateway_exclusion",
		"nat_gateway_exclusion",
		"virtual_private_gateway_exclusion",
		"vpc_peering_exclusion",
	})
	var buf strings.Builder
	for key := range allKeys.Items() {
		var v string
		if slices.Contains(keys, key) {
			v = "enable"
		} else {
			v = "disable"
		}
		fmt.Fprintf(&buf, `%[1]s = %[2]q`+"\n", key, v)
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
