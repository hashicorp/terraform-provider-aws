// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/xray"
	awstypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRayTraceSegmentDestination_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccTraceSegmentDestination_basic,
		"update":        testAccTraceSegmentDestination_update,
		"Identity":      testAccXRayTraceSegmentDestination_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccTraceSegmentDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v xray.GetTraceSegmentDestinationOutput
	resourceName := "aws_xray_trace_segment_destination.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TraceSegmentDestination/basic/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTraceSegmentDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDestination), tfknownvalue.StringExact(awstypes.TraceSegmentDestinationXRay)),
				},
			},
		},
	})
}

func testAccTraceSegmentDestination_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v xray.GetTraceSegmentDestinationOutput
	resourceName := "aws_xray_trace_segment_destination.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TraceSegmentDestination/update/"),
				ConfigVariables: config.Variables{
					"destination":   config.StringVariable(string(awstypes.TraceSegmentDestinationCloudWatchLogs)),
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTraceSegmentDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDestination), tfknownvalue.StringExact(awstypes.TraceSegmentDestinationCloudWatchLogs)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TraceSegmentDestination/update/"),
				ConfigVariables: config.Variables{
					"destination":   config.StringVariable(string(awstypes.TraceSegmentDestinationXRay)),
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTraceSegmentDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDestination), tfknownvalue.StringExact(awstypes.TraceSegmentDestinationXRay)),
				},
			},
		},
	})
}

func testAccCheckTraceSegmentDestinationExists(ctx context.Context, t *testing.T, n string, v *xray.GetTraceSegmentDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).XRayClient(ctx)

		output, err := tfxray.FindTraceSegmentDestination(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
