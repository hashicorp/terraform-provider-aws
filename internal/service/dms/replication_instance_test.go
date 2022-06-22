package dms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDMSReplicationInstance_basic(t *testing.T) {
	// NOTE: Using larger dms.c4.large here for AWS GovCloud (US) support
	replicationInstanceClass := "dms.c4.large"
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_class(rName, replicationInstanceClass),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "100"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_private_ips.#", "1"),
					// ARN resource is its own unique identifier
					resource.TestCheckResourceAttrSet(resourceName, "replication_instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccDMSReplicationInstance_allocatedStorage(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_allocatedStorage(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_allocatedStorage(rName, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "6"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_autoMinorVersionUpgrade(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				Config: testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_availabilityZone(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_availabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "names.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

/*
** Temporarily commented out: "replication_instance_test.go:186: Test validation error: TestStep 1/3 validation error: TestStep missing Config or ImportState".

func TestAccDMSReplicationInstance_engineVersion(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// This acceptance test is designed to test engine version upgrades.
	// Over time, DMS replication instance engine versions are deprecated
	// so they will eventually error on resource creation, e.g.
	//   InvalidParameterValueException: No replication engine found with version: 2.4.2
	// During the PreCheck, we will find candidate engine versions from the
	// orderable replication instances and generate the TestStep.
	// We prefer this method over creating a plural data source that
	// seems impractical for real world usage.
	testSteps := []resource.TestStep{
		{},
		{},
		{
			ResourceName:            resourceName,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"allow_major_version_upgrade", "apply_immediately"},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)

			engineVersions := testAccReplicationInstanceEngineVersionsPreCheck(t)

			testSteps[0] = resource.TestStep{
				Config: testAccReplicationInstanceConfig_engineVersion(rName, engineVersions[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersions[0]),
				),
			}
			testSteps[1] = resource.TestStep{
				Config: testAccReplicationInstanceConfig_engineVersion(rName, engineVersions[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersions[1]),
				),
			}
		},
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps:             testSteps,
	})
}

**
*/

func TestAccDMSReplicationInstance_kmsKeyARN(t *testing.T) {
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccDMSReplicationInstance_multiAz(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
				),
			},
			{
				Config: testAccReplicationInstanceConfig_multiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_preferredMaintenanceWindow(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, "sun:00:30-sun:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "sun:00:30-sun:02:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, "mon:00:30-mon:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "mon:00:30-mon:02:30"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_publiclyAccessible(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_public_ips.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccDMSReplicationInstance_replicationInstanceClass(t *testing.T) {
	// NOTE: Using larger dms.c4.(x)?large here for AWS GovCloud (US) support
	replicationInstanceClass1 := "dms.c4.large"
	replicationInstanceClass2 := "dms.c4.xlarge"
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_class(rName, replicationInstanceClass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_class(rName, replicationInstanceClass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", replicationInstanceClass2),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_replicationSubnetGroupID(t *testing.T) {
	dmsReplicationSubnetGroupResourceName := "aws_dms_replication_subnet_group.test"
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_subnetGroupID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "replication_subnet_group_id", dmsReplicationSubnetGroupResourceName, "replication_subnet_group_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccDMSReplicationInstance_tags(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_tagsOne(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationInstanceConfig_tagsTwo(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReplicationInstanceConfig_tagsOne(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDMSReplicationInstance_vpcSecurityGroupIDs(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func testAccCheckReplicationInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn
		resp, err := conn.DescribeReplicationInstances(&dms.DescribeReplicationInstancesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-instance-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("DMS replication instance error: %v", err)
		}
		if resp == nil || len(resp.ReplicationInstances) == 0 {
			return fmt.Errorf("DMS replication instance not found")
		}

		return nil
	}
}

func testAccCheckReplicationInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_replication_instance" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn

		resp, err := conn.DescribeReplicationInstances(&dms.DescribeReplicationInstancesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-instance-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			for _, replicationInstance := range resp.ReplicationInstances {
				if aws.StringValue(replicationInstance.ReplicationInstanceIdentifier) == rs.Primary.ID {
					return fmt.Errorf("DMS Replication Instance (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}

/*
**

// Ensure at least two engine versions of the replication instance class are available
func testAccReplicationInstanceEngineVersionsPreCheck(t *testing.T) []string {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn

	// Gather all orderable DMS replication instances of the instance class
	// used in the acceptance testing. Not currently available as an input
	// parameter to the describe API call.
	var orderableReplicationInstances []*dms.OrderableReplicationInstance
	input := &dms.DescribeOrderableReplicationInstancesInput{}
	// NOTE: Using larger dms.c4.large here for AWS GovCloud (US) support
	replicationInstanceClass := "dms.c4.large"

	err := conn.DescribeOrderableReplicationInstancesPages(input, func(output *dms.DescribeOrderableReplicationInstancesOutput, lastPage bool) bool {
		for _, orderableReplicationInstance := range output.OrderableReplicationInstances {
			if orderableReplicationInstance == nil {
				continue
			}

			if aws.StringValue(orderableReplicationInstance.ReplicationInstanceClass) == replicationInstanceClass {
				orderableReplicationInstances = append(orderableReplicationInstances, orderableReplicationInstance)
			}
		}

		return !lastPage
	})

	if err != nil {
		t.Fatalf("error describing DMS orderable replication instances: %s", err)
	}

	// Ensure we have enough
	if len(orderableReplicationInstances) < 2 {
		t.Fatalf("found (%d) DMS orderable replication instances for instance class (%s), need at least 2", len(orderableReplicationInstances), replicationInstanceClass)
	}

	// Sort them ascending
	sort.Slice(orderableReplicationInstances, func(i, j int) bool {
		return verify.SemVerLessThan(aws.StringValue(orderableReplicationInstances[i].EngineVersion), aws.StringValue(orderableReplicationInstances[j].EngineVersion))
	})

	engineVersions := make([]string, len(orderableReplicationInstances))

	for i, orderableReplicationInstance := range orderableReplicationInstances {
		engineVersions[i] = aws.StringValue(orderableReplicationInstance.EngineVersion)
	}

	return engineVersions
}

**
*/

func testAccReplicationInstanceConfig_allocatedStorage(rName string, allocatedStorage int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage          = %d
  apply_immediately          = true
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, allocatedStorage, rName)
}

func testAccReplicationInstanceConfig_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  auto_minor_version_upgrade = %t
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, autoMinorVersionUpgrade, rName)
}

func testAccReplicationInstanceConfig_availabilityZone(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  availability_zone          = data.aws_availability_zones.available.names[0]
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, rName)
}

/*
**

func testAccReplicationInstanceConfig_engineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  engine_version              = %q
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id     = %q
}
`, engineVersion, rName)
}

**
*/

func testAccReplicationInstanceConfig_kmsKeyARN(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  kms_key_arn                = aws_kms_key.test.arn
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, rName)
}

func testAccReplicationInstanceConfig_multiAz(rName string, multiAz bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  multi_az                   = %t
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, multiAz, rName)
}

func testAccReplicationInstanceConfig_preferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  preferred_maintenance_window = %q
  replication_instance_class   = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id      = %q
}
`, preferredMaintenanceWindow, rName)
}

func testAccReplicationInstanceConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  publicly_accessible        = %t
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q
}
`, publiclyAccessible, rName)
}

func testAccReplicationInstanceConfig_class(rName, replicationInstanceClass string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  replication_instance_class = %q
  replication_instance_id    = %q
}
`, replicationInstanceClass, rName)
}

func testAccReplicationInstanceConfig_subnetGroupID(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = aws_vpc.test.tags["Name"]
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = %q
  replication_subnet_group_id          = %q
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id     = %q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`, rName, rName, rName, rName)
}

func testAccReplicationInstanceConfig_tagsOne(rName, key1, value1 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q

  tags = {
    %q = %q
  }
}
`, rName, key1, value1)
}

func testAccReplicationInstanceConfig_tagsTwo(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  replication_instance_class = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id    = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, key1, value1, key2, value2)
}

func testAccReplicationInstanceConfig_vpcSecurityGroupIDs(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "test" {
  name   = aws_vpc.test.tags["Name"]
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = aws_vpc.test.tags["Name"]
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = %q
  replication_subnet_group_id          = %q
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id     = %q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids      = [aws_security_group.test.id]
}
`, rName, rName, rName, rName)
}
