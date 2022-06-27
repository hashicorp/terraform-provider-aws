package backup_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBackupVaultDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_vault.test"
	resourceName := "aws_backup_vault.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVaultDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Vault`),
			},
			{
				Config: testAccVaultDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key_arn", resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "recovery_points", resourceName, "recovery_points"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

const testAccVaultDataSourceConfig_nonExistent = `
data "aws_backup_vault" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccVaultDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "down"
    left = "right"
  }
}

data "aws_backup_vault" "test" {
  name = aws_backup_vault.test.name
}
`, rInt)
}
