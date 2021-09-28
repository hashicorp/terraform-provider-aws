package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// func init() {
// 	resource.AddTestSweepers("aws_fsx_storage_virtual_machine", &resource.Sweeper{
// 		Name: "aws_fsx_storage_virtual_machine",
// 		F:    testSweepFSXOntapStorageVirtualMachines,
// 	})
// }

// func testSweepFSXOntapStorageVirtualMachines(region string) error {
// 	client, err := sharedClientForRegion(region)

// 	if err != nil {
// 		return fmt.Errorf("error getting client: %s", err)
// 	}

// 	conn := client.(*AWSClient).fsxconn
// 	sweepResources := make([]*testSweepResource, 0)
// 	var errs *multierror.Error
// 	input := &fsx.DescribeStorageVirtualMachinesInput{}

// 	err = conn.DescribeStorageVirtualMachinesPages(input, func(page *fsx.DescribeStorageVirtualMachinesOutput, lastPage bool) bool {
// 		if page == nil {
// 			return !lastPage
// 		}

// 		for _, svm := range page.StorageVirtualMachines {
// 			if aws.StringValue(svm.StorageVirtualMachineType) != fsx.StorageVirtualMachineTypeOntap {
// 				continue
// 			}

// 			r := resourceAwsFsxStorageVirtualMachine()
// 			d := r.Data(nil)
// 			d.SetId(aws.StringValue(svm.StorageVirtualMachineId))

// 			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
// 		}

// 		return !lastPage
// 	})

// 	if err != nil {
// 		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Ontap File Systems for %s: %w", region, err))
// 	}

// 	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
// 		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Ontap File Systems for %s: %w", region, err))
// 	}

// 	if testSweepSkipSweepError(errs.ErrorOrNil()) {
// 		log.Printf("[WARN] Skipping FSx Storage Virtual Machine sweep for %s: %s", region, errs)
// 		return nil
// 	}

// 	return errs.ErrorOrNil()
// }

func TestAccAWSFsxStorageVirtualMachine_basic(t *testing.T) {
	var svm fsx.StorageVirtualMachine
	resourceName := "aws_fsx_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxStorageVirtualMachineConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxStorageVirtualMachineExists(resourceName, &svm),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`storage-virtual-machine/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_id", "aws_fsx_ontap_file_system.test", "id"),
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

func TestAccAWSFsxStorageVirtualMachine_disappears(t *testing.T) {
	var svm fsx.StorageVirtualMachine
	resourceName := "aws_fsx_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxStorageVirtualMachineConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxStorageVirtualMachineExists(resourceName, &svm),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFsxStorageVirtualMachine(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFsxStorageVirtualMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSFsxStorageVirtualMachine_tags(t *testing.T) {
	var svm1, svm2, svm3 fsx.StorageVirtualMachine
	resourceName := "aws_fsx_storage_virtual_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(fsx.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, fsx.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxStorageVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxStorageVirtualMachineConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxStorageVirtualMachineExists(resourceName, &svm1),
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
				Config: testAccAwsFsxStorageVirtualMachineConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxStorageVirtualMachineExists(resourceName, &svm2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsFsxStorageVirtualMachineConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxStorageVirtualMachineExists(resourceName, &svm3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckFsxStorageVirtualMachineExists(resourceName string, svm *fsx.StorageVirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		resp, err := finder.StorageVirtualMachineByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("FSx Storage Virtual Machine (%s) not found", rs.Primary.ID)
		}

		*svm = *resp

		return nil
	}
}

func testAccCheckFsxStorageVirtualMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_storage_virtual_machine" {
			continue
		}

		svm, err := finder.StorageVirtualMachineByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if svm != nil {
			return fmt.Errorf("FSx Storage Virtual Machine (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccAwsFsxStorageVirtualMachineConfigBase() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
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

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
}
`)
}

func testAccAwsFsxStorageVirtualMachineConfigBasic(rName string) string {
	return composeConfig(testAccAwsFsxStorageVirtualMachineConfigBase(), fmt.Sprintf(`
resource "aws_fsx_storage_virtual_machine" "test" {
  name           = %[1]q
  file_system_id = aws_fsx_ontap_file_system.test.id
}
`, rName))
}

func testAccAwsFsxStorageVirtualMachineConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAwsFsxStorageVirtualMachineConfigBase(), fmt.Sprintf(`
resource "aws_fsx_storage_virtual_machine" "test" {
  name           = %[1]q
  file_system_id = aws_fsx_ontap_file_system.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAwsFsxStorageVirtualMachineConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAwsFsxStorageVirtualMachineConfigBase(), fmt.Sprintf(`
resource "aws_fsx_storage_virtual_machine" "test" {
  name           = %[1]q
  file_system_id = aws_fsx_ontap_file_system.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
