package backup_test

import (
	"fmt"
	"regexp"
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

func TestAccBackupVaultPolicy_basic(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Id\":\"default\".+"))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Id\":\"default\".+")),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("backup:ListRecoveryPointsByBackupVault")),
				),
			},
		},
	})
}

func TestAccBackupVaultPolicy_disappears(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVaultPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVaultPolicy_Disappears_vault(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"
	vaultResourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVault(), vaultResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVaultPolicy_ignoreEquivalent(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("\"Version\":\"2012-10-17\""))),
			},
			{
				Config:   testAccVaultPolicyConfig_newOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVaultPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_policy" {
			continue
		}

		_, err := tfbackup.FindVaultAccessPolicyByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Backup Vault Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVaultPolicyExists(name string, vault *backup.GetBackupVaultAccessPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Vault Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

		output, err := tfbackup.FindVaultAccessPolicyByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*vault = *output

		return nil
	}
}

func testAccVaultPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "backup:DescribeBackupVault",
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:GetBackupVaultAccessPolicy",
        "backup:StartBackupJob",
        "backup:GetBackupVaultNotifications",
        "backup:PutBackupVaultNotifications",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}

func testAccVaultPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "backup:DescribeBackupVault",
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:GetBackupVaultAccessPolicy",
        "backup:StartBackupJob",
        "backup:GetBackupVaultNotifications",
        "backup:PutBackupVaultNotifications",
        "backup:ListRecoveryPointsByBackupVault",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}

func testAccVaultPolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultNotifications",
        "backup:GetBackupVaultNotifications",
        "backup:GetBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:DescribeBackupVault",
        "backup:StartBackupJob",
        "backup:PutBackupVaultAccessPolicy",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}
