package backup_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func TestAccBackupVault_basic(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
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

func TestAccBackupVault_withKMSKey(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccBackupVault_withTags(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "down"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "right"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupVaultWithUpdateTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "downdown"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "rightright"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccBackupVaultWithRemoveTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
		},
	})
}

func TestAccBackupVault_disappears(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault" {
			continue
		}

		input := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeBackupVault(input)

		if err == nil {
			if *resp.BackupVaultName == rs.Primary.ID {
				return fmt.Errorf("Vault '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckVaultExists(name string, vault *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
		params := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeBackupVault(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	input := &backup.ListBackupVaultsInput{}

	_, err := conn.ListBackupVaults(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBackupVaultConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}
`, randInt)
}

func testAccBackupVaultWithKmsKey(randInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Test KMS Key for AWS Backup Vault"
  deletion_window_in_days = 10
}

resource "aws_backup_vault" "test" {
  name        = "tf_acc_test_backup_vault_%d"
  kms_key_arn = aws_kms_key.test.arn
}
`, randInt)
}

func testAccBackupVaultWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "down"
    left = "right"
  }
}
`, randInt)
}

func testAccBackupVaultWithUpdateTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "downdown"
    left = "rightright"
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}

func testAccBackupVaultWithRemoveTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}
