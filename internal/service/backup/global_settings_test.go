// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupGlobalSettings_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccGlobalSettings_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccGlobalSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var settings map[string]string
	resourceName := "aws_backup_global_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSettingsConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "global_settings.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "global_settings.isCrossAccountBackupEnabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalSettingsConfig_basic(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "global_settings.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "global_settings.isCrossAccountBackupEnabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccGlobalSettingsConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "global_settings.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "global_settings.isCrossAccountBackupEnabled", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckGlobalSettingsExists(ctx context.Context, n string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindGlobalSettings(ctx, conn)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccGlobalSettingsConfig_basic(setting string) string {
	return fmt.Sprintf(`
resource "aws_backup_global_settings" "test" {
  global_settings = {
    "isCrossAccountBackupEnabled" = %[1]q
  }
}
`, setting)
}
