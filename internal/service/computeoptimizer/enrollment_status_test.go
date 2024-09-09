// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomputeoptimizer "github.com/hashicorp/terraform-provider-aws/internal/service/computeoptimizer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccComputeOptimizerEnrollmentStatus_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccEnrollmentStatus_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEnrollmentStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_computeoptimizer_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_basic("Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnrollmentStatusConfig_basic("Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
		},
	})
}

func testAccCheckEnrollmentStatusExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

		_, err := tfcomputeoptimizer.FindEnrollmentStatus(ctx, conn)

		return err
	}
}

func testAccEnrollmentStatusConfig_basic(status string) string {
	return fmt.Sprintf(`
resource "aws_computeoptimizer_enrollment_status" "test" {
  status = %[1]q
}
`, status)
}
