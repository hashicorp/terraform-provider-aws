package fsx_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_fsx_backup", &resource.Sweeper{
		Name: "aws_fsx_backup",
		F:    sweepFSXBackups,
	})
}

func sweepFSXBackups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).FSxConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeBackupsInput{}

	err = conn.DescribeBackupsPages(input, func(page *fsx.DescribeBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.Backups {
			r := tffsx.ResourceBackup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.BackupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Backups for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Backups for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Backups sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccFSxBackup_basic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccFSxBackup_disappears(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceBackup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_Disappears_filesystem(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceLustreFileSystem(), "aws_fsx_lustre_file_system.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxBackup_tags(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
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
				Config: testAccBackupTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBackupTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxBackup_implicitTags(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupImplictTagsConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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

func testAccCheckFsxBackupExists(resourceName string, fs *fsx.Backup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		output, err := tffsx.FindBackupByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("FSx Backup (%s) not found", rs.Primary.ID)
		}

		*fs = *output

		return nil
	}
}

func testAccCheckFsxBackupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_backup" {
			continue
		}

		_, err := tffsx.FindBackupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("FSx Backup %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccBackupBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
}
`)
}

func testAccBackupBasicConfig() string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), `
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id
}
`)
}

func testAccBackupTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccBackupTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccBackupBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccBackupImplictTagsConfig(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
  copy_tags_to_backups        = true

  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id
}
`, tagKey1, tagValue1))
}
