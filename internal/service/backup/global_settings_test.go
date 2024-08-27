// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupGlobalSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var settings backup.DescribeGlobalSettingsOutput

	resourceName := "aws_backup_global_settings.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSettingsConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalSettingsExists(ctx, &settings),
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
					testAccCheckGlobalSettingsExists(ctx, &settings),
					resource.TestCheckResourceAttr(resourceName, "global_settings.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "global_settings.isCrossAccountBackupEnabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccGlobalSettingsConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalSettingsExists(ctx, &settings),
					resource.TestCheckResourceAttr(resourceName, "global_settings.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "global_settings.isCrossAccountBackupEnabled", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckGlobalSettingsExists(ctx context.Context, settings *backup.DescribeGlobalSettingsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
		resp, err := conn.DescribeGlobalSettings(ctx, &backup.DescribeGlobalSettingsInput{})
		if err != nil {
			return err
		}

		*settings = *resp

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
