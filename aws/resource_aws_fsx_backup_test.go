package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// func init() {
// 	resource.AddTestSweepers("aws_fsx_lustre_file_system", &resource.Sweeper{
// 		Name: "aws_fsx_lustre_file_system",
// 		F:    testSweepFSXBackups,
// 	})
// }

// func testSweepFSXBackups(region string) error {
// 	client, err := sharedClientForRegion(region)

// 	if err != nil {
// 		return fmt.Errorf("error getting client: %s", err)
// 	}

// 	conn := client.(*AWSClient).fsxconn
// 	sweepResources := make([]*testSweepResource, 0)
// 	var errs *multierror.Error
// 	input := &fsx.DescribeBackupsInput{}

// 	err = conn.DescribeBackupsPages(input, func(page *fsx.DescribeBackupsOutput, lastPage bool) bool {
// 		if page == nil {
// 			return !lastPage
// 		}

// 		for _, fs := range page.Backups {
// 			if aws.StringValue(fs.BackupType) != fsx.BackupTypeLustre {
// 				continue
// 			}

// 			r := resourceAwsFsxBackup()
// 			d := r.Data(nil)
// 			d.SetId(aws.StringValue(fs.BackupId))

// 			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
// 		}

// 		return !lastPage
// 	})

// 	if err != nil {
// 		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Lustre Filesystems for %s: %w", region, err))
// 	}

// 	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
// 		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Lustre Filesystems for %s: %w", region, err))
// 	}

// 	if testSweepSkipSweepError(errs.ErrorOrNil()) {
// 		log.Printf("[WARN] Skipping FSx Lustre Filesystems sweep for %s: %s", region, errs)
// 		return nil
// 	}

// 	return errs.ErrorOrNil()
// }

func TestAccAWSFsxBackup_basic(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxBackupConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`backup/.+`)),
					// resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`fs-.+\.fsx\.`)),
					// resource.TestCheckResourceAttr(resourceName, "export_path", ""),
					// resource.TestCheckResourceAttr(resourceName, "import_path", ""),
					// resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "0"),
					// resource.TestCheckResourceAttrSet(resourceName, "mount_name"),
					// resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					// resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					// resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					// resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
					// resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					// resource.TestCheckResourceAttr(resourceName, "deployment_type", deploymentType),
					// resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					// resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					// resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					// resource.TestCheckResourceAttr(resourceName, "data_compression_type", fsx.DataCompressionTypeNone),
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

func TestAccAWSFsxBackup_disappears(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxBackupConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFsxBackup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSFsxBackup_Tags(t *testing.T) {
	var backup fsx.Backup
	resourceName := "aws_fsx_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxBackupConfigTags1("key1", "value1"),
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
				Config: testAccAwsFsxBackupConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsFsxBackupConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxBackupExists(resourceName, &backup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		output, err := finder.BackupByID(conn, rs.Primary.ID)
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
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_backup" {
			continue
		}

		_, err := finder.BackupByID(conn, rs.Primary.ID)
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

func testAccAwsFsxBackupConfigBase() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
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

func testAccAwsFsxBackupConfigBasic() string {
	return composeConfig(testAccAwsFsxBackupConfigBase(), `
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id
}
`)
}

func testAccAwsFsxBackupConfigTags1(tagKey1, tagValue1 string) string {
	return composeConfig(testAccAwsFsxBackupConfigBase(), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAwsFsxBackupConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAwsFsxBackupConfigBase(), fmt.Sprintf(`
resource "aws_fsx_backup" "test" {
  file_system_id = aws_fsx_lustre_file_system.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
