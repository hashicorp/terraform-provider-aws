// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
)

func TestAccXRayTraceSegmentDestination_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"Identity": testAccXRayTraceSegmentDestination_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
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
