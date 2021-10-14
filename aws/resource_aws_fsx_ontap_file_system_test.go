package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_fsx_ontap_file_system", &resource.Sweeper{
		Name: "aws_fsx_ontap_file_system",
		F:    testSweepFSXOntapFileSystems,
	})
}

func testSweepFSXOntapFileSystems(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).fsxconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeOntap {
				continue
			}

			r := resourceAwsFsxOntapFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.FileSystemId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Ontap File Systems for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Ontap File Systems for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Ontap File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSFsxOntapFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1024"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test2", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", fsx.OntapDeploymentTypeMultiAz1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", fsx.StorageTypeSsd),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_ip_address_range"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_vpc.test", "default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "512"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_subnet_id", "aws_subnet.test1", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.intercluster.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.intercluster.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_fsxAdminPassword(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	pass := sdkacctest.RandomWithPrefix("tf-acc-test")
	pass2 := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigFsxAdminPassword(pass),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids", "fsx_admin_password"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigFsxAdminPassword(pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass2),
				),
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_endpointIpAddressRange(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigEndpointIpAddressRange(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "endpoint_ip_address_range", "198.19.255.0/24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_diskIopsConfiguration(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigDiskIopsConfiguration(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsFsxOntapFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_SecurityGroupIds(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigSecurityGroupIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigSecurityGroupIds2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_routeTableIds(t *testing.T) {
	var filesystem1 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigRouteTable(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem3),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigWeeklyMaintenanceStartTime("1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigWeeklyMaintenanceStartTime("2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_automaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigAutomaticBackupRetentionDays(90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigAutomaticBackupRetentionDays(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigAutomaticBackupRetentionDays(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "1"),
				),
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_kmsKeyId(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxOntapFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigKmsKeyId(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxOntapFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxOntapFileSystemConfigDailyAutomaticBackupStartTime("01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxOntapFileSystemConfigDailyAutomaticBackupStartTime("02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxOntapFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxOntapFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func testAccCheckFsxOntapFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		filesystem, err := finder.FileSystemByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if filesystem == nil {
			return fmt.Errorf("FSx Ontap File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckFsxOntapFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_ontap_file_system" {
			continue
		}

		filesystem, err := finder.FileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx Ontap File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckFsxOntapFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckFsxOntapFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccAwsFsxOntapFileSystemConfigBase() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}
`)
}

func testAccAwsFsxOntapFileSystemConfigBasic() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccAwsFsxOntapFileSystemConfigFsxAdminPassword(pass string) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
  fsx_admin_password  = %[1]q
}
`, pass))
}

func testAccAwsFsxOntapFileSystemConfigEndpointIpAddressRange() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity          = 1024
  subnet_ids                = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type           = "MULTI_AZ_1"
  throughput_capacity       = 512
  preferred_subnet_id       = aws_subnet.test1.id
  endpoint_ip_address_range = "198.19.255.0/24"
}
`)
}

func testAccAwsFsxOntapFileSystemConfigDiskIopsConfiguration() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id

  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = 3072
  }
}
`)
}

func testAccAwsFsxOntapFileSystemConfigRouteTable() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  lifecycle {
    ignore_changes = [tags, tags_all]
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
  route_table_ids     = [aws_route_table.test.id]
}
`)
}

func testAccAwsFsxOntapFileSystemConfigSecurityGroupIds1() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id]
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccAwsFsxOntapFileSystemConfigSecurityGroupIds2() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_security_group" "test2" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccAwsFsxOntapFileSystemConfigTags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAwsFsxOntapFileSystemConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAwsFsxOntapFileSystemConfigWeeklyMaintenanceStartTime(weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity              = 1024
  subnet_ids                    = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type               = "MULTI_AZ_1"
  throughput_capacity           = 512
  preferred_subnet_id           = aws_subnet.test1.id
  weekly_maintenance_start_time = %[1]q
}
`, weeklyMaintenanceStartTime))
}

func testAccAwsFsxOntapFileSystemConfigDailyAutomaticBackupStartTime(dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                  = 1024
  subnet_ids                        = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type                   = "MULTI_AZ_1"
  throughput_capacity               = 512
  preferred_subnet_id               = aws_subnet.test1.id
  daily_automatic_backup_start_time = %[1]q
  automatic_backup_retention_days   = 1
}
`, dailyAutomaticBackupStartTime))
}

func testAccAwsFsxOntapFileSystemConfigAutomaticBackupRetentionDays(retention int) string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 1024
  subnet_ids                      = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type                 = "MULTI_AZ_1"
  throughput_capacity             = 512
  preferred_subnet_id             = aws_subnet.test1.id
  automatic_backup_retention_days = %[1]d
}
`, retention))
}

func testAccAwsFsxOntapFileSystemConfigKmsKeyId() string {
	return acctest.ConfigCompose(testAccAwsFsxOntapFileSystemConfigBase(), `
resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
  kms_key_id          = aws_kms_key.test.arn
}
`)
}
