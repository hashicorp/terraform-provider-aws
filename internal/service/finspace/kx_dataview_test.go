// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxDataview_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxDataviewStatusActive)),
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

func TestAccFinSpaceKxDataview_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxDataview(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxDataview_readWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_readWrite(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccFinSpaceKxDataview_onDemand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_onDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccFinSpaceKxDataview_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKxDataviewConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKxDataviewConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFinSpaceKxDataview_withKxVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dataview finspace.GetKxDataviewOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_dataview.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDataviewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDataviewConfig_withKxVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDataviewExists(ctx, resourceName, &dataview),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxDataviewStatusActive)),
				),
			},
		},
	})
}

func testAccCheckKxDataviewExists(ctx context.Context, name string, dataview *finspace.GetKxDataviewOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDataview, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDataview, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

		resp, err := tffinspace.FindKxDataviewById(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDataview, rs.Primary.ID, err)
		}

		*dataview = *resp

		return nil
	}
}

func testAccCheckKxDataviewDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_finspace_kx_dataview" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

			_, err := tffinspace.FindKxDataviewById(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDataview, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccKxDataviewConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
}
`, rName)
}

func testAccKxDataviewConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = true
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
}
`, rName))
}

func testAccKxDataviewConfig_readWrite(rName string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_volume" "test" {
  name               = %[1]q
  environment_id     = aws_finspace_kx_environment.test.id
  availability_zones = [aws_finspace_kx_environment.test.availability_zones[0]]
  az_mode            = "SINGLE"
  type               = "NAS_1"

  nas1_configuration {
    size = 1200
    type = "SSD_250"
  }
}

resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = false
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  read_write           = true

  segment_configurations {
    db_paths    = ["/*"]
    volume_name = aws_finspace_kx_volume.test.name
    on_demand   = false
  }
}
`, rName))
}

func testAccKxDataviewConfig_onDemand(rName string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_volume" "test" {
  name               = %[1]q
  environment_id     = aws_finspace_kx_environment.test.id
  availability_zones = [aws_finspace_kx_environment.test.availability_zones[0]]
  az_mode            = "SINGLE"
  type               = "NAS_1"

  nas1_configuration {
    size = 1200
    type = "SSD_250"
  }
}

resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = false
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  segment_configurations {
    db_paths    = ["/*"]
    volume_name = aws_finspace_kx_volume.test.name
    on_demand   = true
  }
}
`, rName))
}

func testAccKxDataviewConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = true
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccKxDataviewConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = true
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}

func testAccKxDataviewConfig_withKxVolume(rName string) string {
	return acctest.ConfigCompose(
		testAccKxDataviewConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_volume" "test" {
  name               = %[1]q
  environment_id     = aws_finspace_kx_environment.test.id
  availability_zones = [aws_finspace_kx_environment.test.availability_zones[0]]
  az_mode            = "SINGLE"
  type               = "NAS_1"
  nas1_configuration {
    size = 1200
    type = "SSD_250"
  }
}

resource "aws_finspace_kx_dataview" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  database_name        = aws_finspace_kx_database.test.name
  auto_update          = true
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  segment_configurations {
    db_paths    = ["/*"]
    volume_name = aws_finspace_kx_volume.test.name
  }
}
`, rName))
}
