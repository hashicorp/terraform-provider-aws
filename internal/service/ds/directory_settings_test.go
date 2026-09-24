// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSDirectorySettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DSServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectorySettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectorySettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectorySettingsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "directory_id"),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "setting.0.name", "TLS_1_0"),
					resource.TestCheckResourceAttr(resourceName, "setting.0.value", "Disable"),
					resource.TestCheckResourceAttrSet(resourceName, "setting.0.type"),
					resource.TestCheckResourceAttrSet(resourceName, "setting.0.request_status"),
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

func TestAccDSDirectorySettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DSServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectorySettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectorySettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectorySettingsExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfds.ResourceDirectorySettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDirectorySettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_directory_settings" {
				continue
			}

			_, err := tfds.FindDirectorySettingsByDirectoryID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameDirectorySettingsExported, rs.Primary.ID, err)
			}

			return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameDirectorySettingsExported, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDirectorySettingsExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		_, err := tfds.FindDirectorySettingsByDirectoryID(ctx, conn, rs.Primary.ID)
		return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, rs.Primary.ID, err)
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

	input := &directoryservice.DescribeDirectoriesInput{}
	_, err := conn.DescribeDirectories(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDirectorySettingsConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = "corp.%[1]s.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory_settings" "test" {
  directory_id = aws_directory_service_directory.test.id

  setting {
    name  = "TLS_1_0"
    value = "Disable"
  }
}
`, rName),
	)
}
