// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLakeFormationDataCellsFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datacellsfilter lakeformation.GetDataCellsFilterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationServiceID)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lakeformation", regexache.MustCompile(`datacellsfilter:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccLakeFormationDataCellsFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datacellsfilter lakeformation.GetDataCellsFilterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationServiceID)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceDataCellsFilter, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataCellsFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_data_cells_filter" {
				continue
			}

			//input := &lakeformation.DescribeDataCellsFilterInput{
			//	DataCellsFilterId: aws.String(rs.Primary.ID),
			//}
			//_, err := conn.DescribeDataCellsFilter(ctx, &lakeformation.DescribeDataCellsFilterInput{
			//	DataCellsFilterId: aws.String(rs.Primary.ID),
			//})
			//if errs.IsA[*types.ResourceNotFoundException](err) {
			//	return nil
			//}
			//if err != nil {
			//	return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, err)
			//}

			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataCellsFilterExists(ctx context.Context, name string, datacellsfilter *lakeformation.GetDataCellsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, name, errors.New("not set"))
		}

		//conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)
		//resp, err := conn.DescribeDataCellsFilter(ctx, &lakeformation.DescribeDataCellsFilterInput{
		//	DataCellsFilterId: aws.String(rs.Primary.ID),
		//})
		//
		//if err != nil {
		//	return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, err)
		//}
		//
		//*datacellsfilter = *resp

		return nil
	}
}

func testAccDataCellsFilterPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.ListDataCellsFilterInput{}
	_, err := conn.ListDataCellsFilter(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDataCellsFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_lakeformation_data_cells_filter" "test" {
  data_cells_filter_name             = %[1]q
  engine_type             = "ActiveLakeFormation"
  engine_version          = %[2]q
  host_instance_type      = "lakeformation.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}
