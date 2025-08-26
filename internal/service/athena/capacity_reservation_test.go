// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaCapacityReservation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_dpus", "24"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", regexache.MustCompile(`capacity-reservation/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCapacityReservationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccAthenaCapacityReservation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfathena.ResourceCapacityReservation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaCapacityReservation_targetDPUs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_targetDPUs(rName, 24),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_dpus", "24"),
				),
			},
			{
				Config: testAccCapacityReservationConfig_targetDPUs(rName, 32),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_dpus", "32"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Target DPUs can only be updated once within the first hour the reservation
			// is active, so do not attempt to change back to the original value.
		},
	})
}

func TestAccAthenaCapacityReservation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCapacityReservationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccCapacityReservationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCapacityReservationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCapacityReservationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_capacity_reservation" {
				continue
			}

			name := rs.Primary.Attributes[names.AttrName]
			_, err := tfathena.FindCapacityReservationByName(ctx, conn, name)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Athena, create.ErrActionCheckingDestroyed, tfathena.ResNameCapacityReservation, name, err)
			}

			return create.Error(names.Athena, create.ErrActionCheckingDestroyed, tfathena.ResNameCapacityReservation, name, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCapacityReservationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNameCapacityReservation, resourceName, errors.New("not found"))
		}

		name := rs.Primary.Attributes[names.AttrName]
		if name == "" {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNameCapacityReservation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)
		_, err := tfathena.FindCapacityReservationByName(ctx, conn, name)
		if err != nil {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNameCapacityReservation, name, err)
		}

		return nil
	}
}

func testAccCapacityReservationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccCapacityReservationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_capacity_reservation" "test" {
  name        = %[1]q
  target_dpus = 24
}
`, rName)
}

func testAccCapacityReservationConfig_targetDPUs(rName string, targetDPUs int) string {
	return fmt.Sprintf(`
resource "aws_athena_capacity_reservation" "test" {
  name        = %[1]q
  target_dpus = %[2]d
}
`, rName, targetDPUs)
}

func testAccCapacityReservationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_athena_capacity_reservation" "test" {
  name        = %[1]q
  target_dpus = 24

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCapacityReservationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_athena_capacity_reservation" "test" {
  name        = %[1]q
  target_dpus = 24

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
