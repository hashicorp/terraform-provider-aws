package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSBackupVaultDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_vault.test"
	resourceName := "aws_backup_vault.test"
	rInt := acctest.RandInt()
	vaultName := fmt.Sprintf("tf_acc_test_backup_vault_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupVaultDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Vault`),
			},
			{
				Config: testAccAwsBackupVaultDataSourceConfig_basic(vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key_arn", resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "recovery_points", resourceName, "recovery_points"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSBackupVaultDataSource_withTags(t *testing.T) {
	resourceName := "aws_backup_vault.test"
	datasourceName := "data.aws_backup_vault.test"
	rInt := acctest.RandInt()
	vaultName := fmt.Sprintf("tf_acc_test_backup_vault_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsBackupVaultDataSourceConfig_tags(vaultName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupVaultDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.up", resourceName, "tags.up"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.left", resourceName, "tags.left"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupVaultDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Backup Vault data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Backup Vault data source ID not set")
		}
		return nil
	}
}

const testAccAwsBackupVaultDataSourceConfig_nonExistent = `
data "aws_backup_vault" "test" {
	name = "tf-acc-test-does-not-exist"
}
`

func testAccAwsBackupVaultDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%s"
}
`, name) + testAwsBackupVaultDataSourceConfig
}

func testAccAwsBackupVaultDataSourceConfig_tags(name string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "%s"

  tags = {
    up   = "down"
    left = "right"
  }
}
`, name) + testAwsBackupVaultDataSourceConfig
}

const testAwsBackupVaultDataSourceConfig = `
data "aws_backup_vault" "test" {
	name = aws_backup_vault.test.name
}
`
