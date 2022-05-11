package backup_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBackupVaultLockConfiguration_basic(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_lock_configuration.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultLockConfigurationConfigAll(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockConfigurationExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "changeable_for_days", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_retention_days", "1200"),
					resource.TestCheckResourceAttr(resourceName, "min_retention_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"changeable_for_days"},
			},
		},
	})
}

func TestAccBackupVaultLockConfiguration_disappears(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_lock_configuration.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultLockConfigurationConfigAll(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultLockConfigurationExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVaultLockConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultLockConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_lock_configuration" {
			continue
		}

		_, err := tfbackup.FindBackupVaultByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Backup Vault Lock Configuration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVaultLockConfigurationExists(name string, vault *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Vault Lock Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

		output, err := tfbackup.FindBackupVaultByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*vault = *output

		return nil
	}
}

func testAccBackupVaultLockConfigurationConfigAll(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_lock_configuration" "test" {
  backup_vault_name   = aws_backup_vault.test.name
  changeable_for_days = 3
  max_retention_days  = 1200
  min_retention_days  = 7
}
`, rName)
}
