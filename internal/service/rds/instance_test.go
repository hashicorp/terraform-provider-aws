package rds_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSInstance_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					testAccCheckInstanceAttributes(&dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_major_version_upgrade"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_name", "baz"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttr(resourceName, "engine", "mysql"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "license_model", "general-public-license"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
					resource.TestMatchResourceAttr(resourceName, "option_group_name", regexp.MustCompile(`^default:mysql-\d`)),
					resource.TestMatchResourceAttr(resourceName, "parameter_group_name", regexp.MustCompile(`^default\.mysql\d`)),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "username", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_NameDeprecated_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_nameDeprecated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					testAccCheckInstanceAttributes(&dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_major_version_upgrade"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_name", "baz"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttr(resourceName, "engine", "mysql"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "license_model", "general-public-license"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", "baz"),
					resource.TestCheckResourceAttr(resourceName, "option_group_name", "default:mysql-8-0"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.mysql8.0"),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "username", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_onlyMajorVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance
	resourceName := "aws_db_instance.test"
	engine := "mysql"
	engineVersion := "8.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_majorVersionOnly(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "engine", engine),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersion),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"engine_version",
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_namePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	const identifierPrefix = "tf-acc-test-prefix-"
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_namePrefix(identifierPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "identifier", identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_nameGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "identifier"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", resource.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_kmsKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_db_instance.bar"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_subnetGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_subnetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", rName),
				),
			},
			{
				Config: testAccInstanceConfig_subnetGroupUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", fmt.Sprintf("%s-2", rName)),
				),
			},
		},
	})
}

func TestAccRDSInstance_optionGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_optionGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "option_group_name", rName),
				),
			},
		},
	})
}

func TestAccRDSInstance_iamAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_iamAuth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.bar", &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_allowMajorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_allowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_allowMajorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_dbSubnetGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_DBSubnetGroupName_ramShared(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_DBSubnetGroupName_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_deletionProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccInstanceConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_finalSnapshotIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// testAccCheckInstanceSnapshot verifies a database snapshot is
		// created, and subsequently deletes it
		CheckDestroy: testAccCheckInstanceSnapshot,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_finalSnapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.snapshot", &snap),
				),
			},
		},
	})
}

func TestAccRDSInstance_FinalSnapshotIdentifier_skipFinalSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var snap rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceNoSnapshot,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_FinalSnapshotID_skipFinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.snapshot", &snap),
				),
			},
		},
	})
}

func TestAccRDSInstance_isAlreadyBeingDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mariaDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(rName),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccInstanceConfig_mariaDB(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccRDSInstance_maxAllocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_maxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
			{
				Config: testAccInstanceConfig_maxAllocatedStorage(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
				),
			},
			{
				Config: testAccInstanceConfig_maxAllocatedStorage(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "15"),
				),
			},
			{
				Config: testAccInstanceConfig_maxAllocatedStorage(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
				),
			},
		},
	})
}

func TestAccRDSInstance_password(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Password should not be shown in error message
			{
				Config:      testAccInstanceConfig_password(rName, "invalid"),
				ExpectError: regexp.MustCompile(`MasterUserPassword: "\*{8}",`),
			},
			{
				Config: testAccInstanceConfig_password(rName, "valid-password"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "password", "valid-password"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "name", sourceResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "username", sourceResourceName, "username"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_namePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const identifierPrefix = "tf-acc-test-prefix-"
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_namePrefix(identifierPrefix, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "identifier", identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_nameGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_nameGenerated(sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "identifier"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", resource.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_addLater(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_addLaterSetup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
				),
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_addLater(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_iops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_iops(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iops", "1000"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allocatedStorageAndIops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allocatedStorageAndIOPS(rName, 220, 2200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "220"),
					resource.TestCheckResourceAttr(resourceName, "iops", "2200"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allowMajorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_autoMinorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_availabilityZone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_availabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_backupRetentionPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_backupRetentionPeriod(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "1"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_backupWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_backupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_dbSubnetGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_dbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDBDBSubnetGroupName_ramShared(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternateAccountAndAlternateRegion(&providers),
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDBDBSubnetGroupName_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_deletionProtection(t *testing.T) {
	acctest.Skip(t, "CreateDBInstanceReadReplica API currently ignores DeletionProtection=true with SourceDBInstanceIdentifier set")
	// --- FAIL: TestAccRDSInstance_ReplicateSourceDB_deletionProtection (1624.88s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'deletion_protection' expected "true", got "false"
	//
	// Action=CreateDBInstanceReadReplica&AutoMinorVersionUpgrade=true&CopyTagsToSnapshot=false&DBInstanceClass=db.t2.micro&DBInstanceIdentifier=tf-acc-test-6591588621809891413&DeletionProtection=true&PubliclyAccessible=false&SourceDBInstanceIdentifier=tf-acc-test-6591588621809891413-source&Tags=&Version=2014-10-31
	// <RestoreDBInstanceFromDBSnapshotResponse xmlns="http://rds.amazonaws.com/doc/2014-10-31/">
	//   <RestoreDBInstanceFromDBSnapshotResult>
	//     <DBInstance>
	//       <DeletionProtection>false</DeletionProtection>
	//
	// AWS Support has confirmed this issue and noted that it will be fixed in the future.

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_maintenanceWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_maintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_maxAllocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_maxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_monitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_monitoring(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_multiAZ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_multiAZ(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameSameSetOnBoth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_sameSetOnBoth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", sourceResourceName, "parameter_group_name"),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameDifferentSetOnBoth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_differentSetOnBoth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, "parameter_group_name", fmt.Sprintf("%s-source", rName)),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameReplicaCopiesValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_replicaCopiesValue(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", sourceResourceName, "parameter_group_name"),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameSetOnReplica(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_setOnReplica(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_port(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_port(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "port", "9999"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_caCertificateIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"
	certifiateDataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_caCertificateID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttrPair(sourceResourceName, "ca_cert_identifier", certifiateDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", certifiateDataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_replicaMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_replicaMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "replica_mode", "mounted"),
				),
			},
		},
	})
}

// When an RDS Instance is added in a separate apply from the creation of the source instance, and the
// parameter group is changed on the replica, it can sometimes lead to the API trying to reboot the instance
//  whenanother "management operation" is in progress:
// InvalidDBInstanceState: Instance cannot currently reboot due to an in-progress management operation
// https://github.com/hashicorp/terraform-provider-aws/issues/11905
func TestAccRDSInstance_ReplicateSourceDB_parameterGroupTwoStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	sourceResourceName := "aws_db_instance.source"
	parameterGroupResourceName := "aws_db_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, "parameter_group_name", "default.oracle-ee-19"),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_parameterGroupTwoStep(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, "parameter_group_name", "default.oracle-ee-19"),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "replica_mode", "open-read-only"),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", parameterGroupResourceName, "id"),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_S3Import_basic(t *testing.T) {
	var snap rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPrefix := sdkacctest.RandString(5)

	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_S3Import_basic(rName, bucketPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &snap),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_S3Import_nameDeprecated(t *testing.T) {
	var snap rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPrefix := sdkacctest.RandString(5)

	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_S3Import_NameDeprecated_basic(rName, bucketPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &snap),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_S3Import_namePrefix(t *testing.T) {
	var snap rds.DBInstance
	const identifierPrefix = "tf-acc-test-prefix-"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPrefix := sdkacctest.RandString(5)

	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_S3Import_namePrefix(rName, bucketPrefix, identifierPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &snap),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "identifier", identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_S3Import_nameGenerated(t *testing.T) {
	var snap rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPrefix := sdkacctest.RandString(5)

	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_S3Import_nameGenerated(rName, bucketPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &snap),
					create.TestCheckResourceAttrNameGenerated(resourceName, "identifier"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", resource.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", sourceDbResourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(resourceName, "allocated_storage", sourceDbResourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", sourceDbResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", sourceDbResourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(resourceName, "username", sourceDbResourceName, "username"),
					resource.TestCheckResourceAttrPair(resourceName, "name", sourceDbResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "maintenance_window", sourceDbResourceName, "maintenance_window"),
					resource.TestCheckResourceAttrPair(resourceName, "option_group_name", sourceDbResourceName, "option_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", sourceDbResourceName, "parameter_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "port", sourceDbResourceName, "port"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_namePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const identifierPrefix = "tf-acc-test-prefix-"
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotIdentifier_namePrefix(identifierPrefix, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "identifier", identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_nameGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotIdentifier_nameGenerated(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "identifier"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", resource.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_AssociationRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1, dbInstance2 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
				),
			},
			{
				Config: testAccInstanceConfig_SnapshotID_associationRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance2),
					testAccCheckInstanceNotRecreated(&dbInstance1, &dbInstance2),
					resource.TestCheckResourceAttrPair(resourceName, "allocated_storage", sourceDbResourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", sourceDbResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "username", sourceDbResourceName, "username"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_allocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_allocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_io1Storage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_io1Storage(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iops", "1000"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_allowMajorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_allowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_autoMinorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_availabilityZone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_availabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupRetentionPeriodOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_backupRetentionPeriod(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "1"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupRetentionPeriodUnset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_BackupRetentionPeriod_unset(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "0"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_backupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_dbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupNameRAMShared(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupNameVPCSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_deletionProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccInstanceConfig_SnapshotID_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_iamDatabaseAuthenticationEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_iamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_maintenanceWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_maintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_maxAllocatedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_maxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_monitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_monitoring(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_multiAZ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_multiAZ(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_multiAZSQLServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_MultiAZ_sqlServer(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_parameterGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_parameterGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_port(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_port(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "port", "9999"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_tagsRemove(t *testing.T) {
	acctest.Skip(t, "To be fixed: https://github.com/hashicorp/terraform-provider-aws/issues/5959")
	// --- FAIL: TestAccRDSInstance_SnapshotIdentifierTags_unset (1086.15s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'tags.%' expected "0", got "1"

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_Tags_clear(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

// Regression reference: https://github.com/hashicorp/terraform-provider-aws/issues/5360
// This acceptance test explicitly tests when snapshot_identifier is set,
// vpc_security_group_ids is set (which triggered the resource update function),
// and tags is set which was missing its ARN used for tagging
func TestAccRDSInstance_SnapshotIdentifier_vpcSecurityGroupIDsTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_VPCSecurityGroupIDs_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccRDSInstance_monitoringInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_enabledToDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_enabledToRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_removedToEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3760 .
// We apply a plan, then change just the iops. If the apply succeeds, we
// consider this a pass, as before in 3760 the request would fail
func TestAccRDSInstance_separateIopsUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.bar", &v),
					testAccCheckInstanceAttributes(&v),
				),
			},

			{
				Config: testAccInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.bar", &v),
					testAccCheckInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccRDSInstance_portUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotInstanceConfig_mySQLPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
				),
			},

			{
				Config: testAccInstanceConfig_SnapshotInstanceConfig_updateMySQLPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "port", "3305"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_tz(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_MSSQL_timezone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckInstanceAttributes_MSSQL(&v, ""),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "20"),
					resource.TestCheckResourceAttr(resourceName, "engine", "sqlserver-ex"),
				),
			},

			{
				Config: testAccInstanceConfig_MSSQL_timezone_AKST(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckInstanceAttributes_MSSQL(&v, "Alaskan Standard Time"),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "20"),
					resource.TestCheckResourceAttr(resourceName, "engine", "sqlserver-ex"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_domain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vBefore, vAfter rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomain()
	directory1 := domain.RandomSubdomain().String()
	directory2 := domain.RandomSubdomain().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlDomain(rName, directory1, directory2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &vBefore),
					testAccCheckInstanceDomainAttributes(directory1, &vBefore),
					resource.TestCheckResourceAttrSet(resourceName, "domain"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
			{
				Config: testAccInstanceConfig_mssqlUpdateDomain(rName, directory1, directory2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &vAfter),
					testAccCheckInstanceDomainAttributes(directory2, &vAfter),
					resource.TestCheckResourceAttrSet(resourceName, "domain"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_domainSnapshotRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v, vRestoredInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	originResourceName := "aws_db_instance.origin"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlDomainSnapshotRestore(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &vRestoredInstance),
					testAccCheckInstanceExists(originResourceName, &v),
					testAccCheckInstanceDomainAttributes(domain, &vRestoredInstance),
					resource.TestCheckResourceAttrSet(resourceName, "domain"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MySQL_snapshotRestoreWithEngineVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v, vRestoredInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	restoreResourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mySQLSnapshotRestoreEngineVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(restoreResourceName, &vRestoredInstance),
					testAccCheckInstanceExists(resourceName, &v),
					// Hardcoded older version. Will to update when no longer compatible to upgrade from this to the default version.
					resource.TestCheckResourceAttr(resourceName, "engine_version", "8.0.25"),
					resource.TestCheckResourceAttrPair(restoreResourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
				),
			},
		},
	})
}

func TestAccRDSInstance_minorVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_autoMinorVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.bar", &v),
				),
			},
		},
	})
}

func TestAccRDSInstance_ec2Classic(t *testing.T) {
	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceEC2ClassicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ec2Classic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEC2ClassicExists(resourceName, &v),
				),
			},
		},
	})
}

func TestAccRDSInstance_cloudWatchLogsExport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_db_instance.bar", &v),
				),
			},
			{
				ResourceName:      "aws_db_instance.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_EnabledCloudWatchLogsExports_mySQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.bar"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportConfigurationAdd(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportConfigurationModify(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportConfigurationDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
				),
			},
		},
	})
}

func TestAccRDSInstance_EnabledCloudWatchLogsExports_msSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_EnabledCloudWatchLogsExports_mssql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_EnabledCloudWatchLogsExports_oracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_EnabledCloudWatchLogsExports_oracle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_EnabledCloudWatchLogsExports_postgresql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_EnabledCloudWatchLogsExports_postgreSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_noDeleteAutomatedBackups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceAutomatedBackups,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noDeleteAutomatedBackups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func testAccCheckInstanceAutomatedBackups(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		log.Printf("[INFO] Trying to locate the DBInstance Automated Backup")
		describeOutput, err := conn.DescribeDBInstanceAutomatedBackups(
			&rds.DescribeDBInstanceAutomatedBackupsInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})
		if err != nil {
			return err
		}

		if describeOutput == nil || len(describeOutput.DBInstanceAutomatedBackups) == 0 {
			return fmt.Errorf("Automated backup for %s not found", rs.Primary.ID)
		}

		log.Printf("[INFO] Deleting automated backup for %s", rs.Primary.ID)
		_, err = conn.DeleteDBInstanceAutomatedBackup(
			&rds.DeleteDBInstanceAutomatedBackupInput{
				DbiResourceId: describeOutput.DBInstanceAutomatedBackups[0].DbiResourceId,
			})
		if err != nil {
			return err
		}
	}

	return testAccCheckInstanceDestroy(s)
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		_, err := tfrds.FindDBInstanceByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DB Instance %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != "mysql" {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		if *v.EngineVersion == "" {
			return fmt.Errorf("bad engine_version: %#v", *v.EngineVersion)
		}

		if *v.BackupRetentionPeriod != 0 {
			return fmt.Errorf("bad backup_retention_period: %#v", *v.BackupRetentionPeriod)
		}

		return nil
	}
}

func testAccCheckInstanceAttributes_MSSQL(v *rds.DBInstance, tz string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != "sqlserver-ex" {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		rtz := ""
		if v.Timezone != nil {
			rtz = *v.Timezone
		}

		if tz != rtz {
			return fmt.Errorf("Expected (%s) Timezone for MSSQL test, got (%s)", tz, rtz)
		}

		return nil
	}
}

func testAccCheckInstanceDomainAttributes(domain string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dm := range v.DomainMemberships {
			if *dm.FQDN != domain {
				continue
			}

			return nil
		}

		return fmt.Errorf("Domain %s not found in domain memberships", domain)
	}
}

func testAccCheckInstanceParameterApplyStatusInSync(dbInstance *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dbParameterGroup := range dbInstance.DBParameterGroups {
			parameterApplyStatus := aws.StringValue(dbParameterGroup.ParameterApplyStatus)
			if parameterApplyStatus != "in-sync" {
				id := aws.StringValue(dbInstance.DBInstanceIdentifier)
				parameterGroupName := aws.StringValue(dbParameterGroup.DBParameterGroupName)
				return fmt.Errorf("expected DB Instance (%s) Parameter Group (%s) apply status to be: \"in-sync\", got: %q", id, parameterGroupName, parameterApplyStatus)
			}
		}

		return nil
	}
}

func testAccCheckInstanceReplicaAttributes(source, replica *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if replica.ReadReplicaSourceDBInstanceIdentifier != nil && *replica.ReadReplicaSourceDBInstanceIdentifier != *source.DBInstanceIdentifier {
			return fmt.Errorf("bad source identifier for replica, expected: '%s', got: '%s'", *source.DBInstanceIdentifier, *replica.ReadReplicaSourceDBInstanceIdentifier)
		}

		return nil
	}
}

func testAccCheckInstanceSnapshot(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		awsClient := acctest.Provider.Meta().(*conns.AWSClient)
		conn := awsClient.RDSConn

		log.Printf("[INFO] Trying to locate the DBInstance Final Snapshot")
		snapOutput, err := conn.DescribeDBSnapshots(
			&rds.DescribeDBSnapshotsInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})

		if err != nil {
			return err
		}

		if snapOutput == nil || len(snapOutput.DBSnapshots) == 0 {
			return fmt.Errorf("Snapshot %s not found", rs.Primary.Attributes["final_snapshot_identifier"])
		}

		// verify we have the tags copied to the snapshot
		tagsARN := aws.StringValue(snapOutput.DBSnapshots[0].DBSnapshotArn)
		listTagsOutput, err := conn.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: aws.String(tagsARN),
		})
		if err != nil {
			return fmt.Errorf("Error retrieving tags for ARN (%s): %s", tagsARN, err)
		}

		if listTagsOutput.TagList == nil || len(listTagsOutput.TagList) == 0 {
			return fmt.Errorf("Tag list is nil or zero: %s", listTagsOutput.TagList)
		}

		var found bool
		for _, t := range listTagsOutput.TagList {
			if aws.StringValue(t.Key) == "Name" && strings.HasPrefix(aws.StringValue(t.Value), acctest.ResourcePrefix) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Expected to find tag Name with prefix \"%s\", but wasn't found. Tags: %s", acctest.ResourcePrefix, listTagsOutput.TagList)
		}
		// end tag search

		log.Printf("[INFO] Deleting the Snapshot %s", rs.Primary.Attributes["final_snapshot_identifier"])
		_, err = conn.DeleteDBSnapshot(
			&rds.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})
		if err != nil {
			return err
		}

		resp, err := conn.DescribeDBInstances(
			&rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
				continue
			}
			return err

		}

		if len(resp.DBInstances) != 0 && aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) == rs.Primary.ID {
			return fmt.Errorf("DB Instance still exists")
		}
	}

	return nil
}

func testAccCheckInstanceNoSnapshot(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		resp, err := conn.DescribeDBInstances(
			&rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
			return err
		}

		if len(resp.DBInstances) != 0 && aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) == rs.Primary.ID {
			return fmt.Errorf("DB Instance still exists")
		}

		_, err = conn.DescribeDBSnapshots(
			&rds.DescribeDBSnapshotsInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSnapshotNotFoundFault) {
			return err
		}
	}

	return nil
}

func testAccCheckInstanceNotRecreated(instance1, instance2 *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(instance1.InstanceCreateTime).Equal(aws.TimeValue(instance2.InstanceCreateTime)) {
			return fmt.Errorf("database instance was recreated. expected: %s, got: %s", instance1.InstanceCreateTime, instance2.InstanceCreateTime)
		}
		return nil
	}
}

func testAccCheckInstanceExists(n string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		output, err := tfrds.FindDBInstanceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckInstanceEC2ClassicDestroy(s *terraform.State) error {
	conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		_, err := tfrds.FindDBInstanceByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DB Instance %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckInstanceEC2ClassicExists(resourceName string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource (%s) state not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).RDSConn

		output, err := tfrds.FindDBInstanceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8792
func TestAccRDSInstance_PerformanceInsightsEnabled_disabledToEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_PerformanceInsightsEnabled_enabledToDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_performanceInsightsKMSKeyID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyIdDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccRDSInstance_performanceInsightsRetentionPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsRetentionPeriod(rName, 731),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "731"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsRetentionPeriod(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_performanceInsightsEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_performanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_performanceInsightsEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_performanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSInstance_caCertificateIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	resourceName := "aws_db_instance.bar"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_caCertificateID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_sourceIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_sourceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"latest_restorable_time", // dynamic value of a DBInstance
					"password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_sourceResourceID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_sourceResourceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"latest_restorable_time", // dynamic value of a DBInstance
					"password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_NationalCharacterSet_oracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_NationalCharacterSet_oracle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "UTF8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_NoNationalCharacterSet_oracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_NoNationalCharacterSet_oracle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "AL16UTF16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_coIPEnabled(t *testing.T) {
	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						resourceName, "customer_owned_ip_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_CoIPEnabled_disabledToEnabled(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_CoIPEnabled_enabledToDisabled(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "false"),
				),
			},
		},
	})
}

func TestAccRDSInstance_CoIPEnabled_restoreToPointInTime(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CoIPEnabled_restorePointInTime(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceName, &sourceDbInstance),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"latest_restorable_time", // dynamic value of a DBInstance
					"password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_CoIPEnabled_snapshotIdentifier(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.test"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CoIPEnabled_snapshotID(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", "true"),
				),
			},
		},
	})
}

func TestAccRDSInstance_license(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_license(rName, "license-included"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "license_model", "license-included"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccInstanceConfig_license(rName, "bring-your-own-license"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "license_model", "bring-your-own-license"),
				),
			},
		},
	})
}

func testAccInstanceConfig_orderableClass(engine, license, storage, classes string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}
`, engine, license, storage, classes)
}

func testAccInstanceConfig_orderableClassMySQL() string {
	return testAccInstanceConfig_orderableClass("mysql", "general-public-license", "standard", mySQLPreferredInstanceClasses)
}

func testAccInstanceConfig_orderableClassMariadb() string {
	return testAccInstanceConfig_orderableClass("mariadb", "general-public-license", "standard", mariaDBPreferredInstanceClasses)
}

func testAccInstanceConfig_orderableClassSQLServerEx() string {
	return testAccInstanceConfig_orderableClass("sqlserver-ex", "license-included", "standard", sqlServerPreferredInstanceClasses)
}

func testAccInstanceConfig_orderableClassSQLServerSe() string {
	return testAccInstanceConfig_orderableClass("sqlserver-se", "license-included", "standard", sqlServerSEPreferredInstanceClasses)
}

func testAccInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "test"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_nameDeprecated(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz" # deprecated
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "test"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_majorVersionOnly() string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`)
}

func testAccInstanceConfig_namePrefix(identifierPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier_prefix   = %[1]q
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  publicly_accessible = true
  skip_final_snapshot = true
  username            = "root"
}
`, identifierPrefix))
}

func testAccInstanceConfig_nameGenerated() string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(), `
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  publicly_accessible = true
  skip_final_snapshot = true
  username            = "root"
}
`)
}

func testAccInstanceConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[2]s]

  supports_storage_encryption = true
}

resource "aws_db_instance" "bar" {
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  kms_key_id              = aws_kms_key.foo.arn
  db_name                 = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  storage_encrypted       = true
  username                = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName, mySQLPreferredInstanceClasses)
}

func testAccInstanceConfig_caCertificateID() string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), `
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "bar" {
  allocated_storage   = 10
  apply_immediately   = true
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`)
}

func testAccInstanceConfig_optionGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name              = data.aws_rds_orderable_db_instance.test.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = aws_db_option_group.test.engine_name
  engine_version      = aws_db_option_group.test.major_engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  option_group_name   = aws_db_option_group.test.name
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName))
}

func testAccInstanceConfig_iamAuth(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[1]s]

  supports_iam_database_authentication = true
}

resource "aws_db_instance" "bar" {
  identifier                          = %[2]q
  allocated_storage                   = 10
  engine                              = data.aws_rds_engine_version.default.engine
  engine_version                      = data.aws_rds_engine_version.default.version
  instance_class                      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                             = "baz"
  password                            = "barbarbarbar"
  username                            = "foo"
  backup_retention_period             = 0
  skip_final_snapshot                 = true
  parameter_group_name                = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  iam_database_authentication_enabled = true
}
`, mySQLPreferredInstanceClasses, rName)
}

func testAccInstanceConfig_FinalSnapshotID_skipFinalSnapshot(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "snapshot" {
  identifier = %[1]q

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  password                = "barbarbarbar"
  username                = "foo"
  backup_retention_period = 1

  publicly_accessible = true

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"

  skip_final_snapshot       = true
  final_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccInstanceConfig_S3Import_Base(rName, bucketPrefix string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_s3_bucket" "xtrabackup" {
  bucket = %[1]q
}

resource "aws_s3_object" "xtrabackup_db" {
  bucket = aws_s3_bucket.xtrabackup.id
  key    = "%[2]s/mysql-5-6-xtrabackup.tar.gz"
  source = "./testdata/mysql-5-6-xtrabackup.tar.gz"
  etag   = filemd5("./testdata/mysql-5-6-xtrabackup.tar.gz")
}

data "aws_partition" "current" {}

resource "aws_iam_role" "rds_s3_access_role" {
  name = "%[1]s-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = "%[1]s-policy"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.xtrabackup.arn}",
        "${aws_s3_bucket.xtrabackup.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "test-attach" {
  name = "%[1]s-policy-attachment"

  roles = [
    aws_iam_role.rds_s3_access_role.name,
  ]

  policy_arn = aws_iam_policy.test.arn
}

resource "aws_db_subnet_group" "foo" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  # instance class db.t2.micro is not supported for restoring from S3
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium", "db.t3.medium"]
}
`, rName, bucketPrefix))
}

func testAccInstanceConfig_S3Import_basic(rName, bucketPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_S3Import_Base(rName, bucketPrefix),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier = %[1]q

  allocated_storage          = 5
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                    = "baz"
  password                   = "barbarbarbar"
  publicly_accessible        = false
  username                   = "foo"
  backup_retention_period    = 0

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.foo.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6" # leave at 5.6 until someone makes a new testdata restore file

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = %[2]q
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, rName, bucketPrefix))
}

func testAccInstanceConfig_S3Import_NameDeprecated_basic(rName, bucketPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_S3Import_Base(rName, bucketPrefix),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier = %[1]q

  allocated_storage          = 5
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  name                       = "baz" # deprecated
  password                   = "barbarbarbar"
  publicly_accessible        = false
  username                   = "foo"
  backup_retention_period    = 0

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.foo.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6" # leave at 5.6 until someone makes a new testdata restore file

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = %[2]q
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, rName, bucketPrefix))
}

func testAccInstanceConfig_S3Import_namePrefix(rName, bucketPrefix, identifierPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_S3Import_Base(rName, bucketPrefix),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier_prefix = %[1]q

  allocated_storage          = 5
  engine                     = data.aws_rds_orderable_db_instance.test.engine
  engine_version             = data.aws_rds_engine_version.default.version
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                    = "baz"
  password                   = "barbarbarbar"
  publicly_accessible        = false
  username                   = "foo"
  backup_retention_period    = 0

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.foo.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6" # leave at 5.6 until someone makes a new testdata restore file

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = %[2]q
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, identifierPrefix, bucketPrefix))
}

func testAccInstanceConfig_S3Import_nameGenerated(rName, bucketPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_S3Import_Base(rName, bucketPrefix),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage          = 5
  engine                     = data.aws_rds_orderable_db_instance.test.engine
  engine_version             = data.aws_rds_engine_version.default.version
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                    = "baz"
  password                   = "barbarbarbar"
  publicly_accessible        = false
  username                   = "foo"
  backup_retention_period    = 0

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.foo.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6" # leave at 5.6 until someone makes a new testdata restore file

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = %[1]q
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, bucketPrefix))
}

func testAccInstanceConfig_finalSnapshotID(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "snapshot" {
  identifier = %[1]q

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  password                = "barbarbarbar"
  publicly_accessible     = true
  username                = "foo"
  backup_retention_period = 1

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"

  copy_tags_to_snapshot     = true
  final_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_monitoringInterval(rName string, monitoringInterval int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName, monitoringInterval)
}

func testAccInstanceConfig_monitoringRoleARNRemoved(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName)
}

func testAccInstanceConfig_monitoringRoleARN(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName)
}

const testAccInstanceBaseConfig = `
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"
}
`

func testAccInstanceConfig_RestoreToPointInTime_sourceID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceBaseConfig,
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = %[1]q
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_RestoreToPointInTime_sourceResourceID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceBaseConfig,
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = %[1]q
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_dbi_resource_id     = aws_db_instance.test.resource_id
    use_latest_restorable_time = true
  }
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]

  storage_type  = "io1"
  supports_iops = true
}

resource "aws_db_instance" "bar" {
  identifier           = %[1]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage = 200
  iops              = %[2]d
}
`, rName, iops)
}

func testAccInstanceConfig_SnapshotInstanceConfig_mySQLPort(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  port                 = 3306
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotInstanceConfig_updateMySQLPort(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_subnetGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  db_subnet_group_name = aws_db_subnet_group.test.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0
  apply_immediately       = true
}
`, rName))
}

func testAccInstanceConfig_subnetGroupUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_subnet" "test3" {
  cidr_block        = "10.10.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test2.id

  tags = {
    Name = "%[1]s-3"
  }
}

resource "aws_subnet" "test4" {
  cidr_block        = "10.10.4.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test2.id

  tags = {
    Name = "%[1]s-4"
  }
}

resource "aws_db_subnet_group" "test2" {
  name       = "%[1]s-2"
  subnet_ids = [aws_subnet.test3.id, aws_subnet.test4.id]

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  db_subnet_group_name = aws_db_subnet_group.test2.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0
  apply_immediately       = true

  depends_on = [aws_db_subnet_group.test]
}
`, rName))
}

func testAccInstanceConfig_MSSQL_timezone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.rds_one.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "somecrazypassword"
  skip_final_snapshot     = true
  username                = "somecrazyusername"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name = %[1]q

  description = %[1]q
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_MSSQL_timezone_AKST(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.rds_one.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "somecrazypassword"
  skip_final_snapshot     = true
  timezone                = "Alaskan Standard Time"
  username                = "somecrazyusername"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name = %[1]q

  description = %[1]q
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": "sts:AssumeRole",
		"Principal": {
			"Service": "rds.${data.aws_partition.current.dns_suffix}"
		},
		"Effect": "Allow",
		"Sid": ""
		}
	]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rName)
}

func testAccInstanceConfig_VPCConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_db_subnet_group" "test" {
  name = %[1]q

  subnet_ids = aws_subnet.test[*].id
}
`, rName)
}

func testAccInstanceConfig_MSSQLDomain_SharedConfig(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccInstanceConfig_VPCConfigBase(rName),
		testAccInstanceConfig_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_directory_service_directory" "directory" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

data "aws_partition" "current" {}
`, rName, domain))
}

func testAccInstanceConfig_mssqlDomain(rName, directory1, directory2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_MSSQLDomain_SharedConfig(rName, directory1),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "somecrazypassword"
  skip_final_snapshot     = true
  username                = "somecrazyusername"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory.id
  domain_iam_role_name = aws_iam_role.role.name
}

resource "aws_directory_service_directory" "directory-2" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, directory2))
}

func testAccInstanceConfig_mssqlUpdateDomain(rName, directory1, directory2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_MSSQLDomain_SharedConfig(rName, directory1),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "somecrazypassword"
  skip_final_snapshot     = true
  username                = "somecrazyusername"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory-2.id
  domain_iam_role_name = aws_iam_role.role.name
}

resource "aws_directory_service_directory" "directory-2" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, directory2))
}

func testAccInstanceConfig_mssqlDomainSnapshotRestore(rName, directory string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_MSSQLDomain_SharedConfig(rName, directory),
		fmt.Sprintf(`
resource "aws_db_instance" "origin" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "somecrazypassword"
  skip_final_snapshot = true
  username            = "somecrazyusername"
}

resource "aws_db_snapshot" "origin" {
  db_instance_identifier = aws_db_instance.origin.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = aws_db_instance.origin.engine
  engine_version          = aws_db_instance.origin.engine_version
  identifier              = "%[1]s-restore"
  instance_class          = aws_db_instance.origin.instance_class
  password                = "somecrazypassword"
  skip_final_snapshot     = true
  username                = "somecrazyusername"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory.id
  domain_iam_role_name = aws_iam_role.role.name

  snapshot_identifier = aws_db_snapshot.origin.id
}
`, rName))
}

func testAccInstanceConfig_mySQLSnapshotRestoreEngineVersion(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name        = %[1]q
  description = "db subnets"
  subnet_ids  = [aws_subnet.test.id, aws_subnet.test2.id]
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "test" {
  allocated_storage   = 20
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = "8.0.25" # test is from older to newer version, update when restore from this to current default version is incompatible
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  skip_final_snapshot = true
  username            = "root"
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "restore" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-restore"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "password"
  skip_final_snapshot     = true
  snapshot_identifier     = aws_db_snapshot.test.id
  username                = "root"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name = %[1]q

  description = "TF Testing"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_allowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage           = 10
  allow_major_version_upgrade = %[1]t
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  identifier                  = %[2]q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                     = "baz"
  password                    = "barbarbarbar"
  skip_final_snapshot         = true
  username                    = "foo"
}
`, allowMajorVersionUpgrade, rName))
}

func testAccInstanceConfig_autoMinorVersion(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  identifier          = %[1]q
  allocated_storage   = 10
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  password            = "barbarbarbar"
  username            = "foo"
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"
  subnet_ids  = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_db_instance" "bar" {
  identifier = %[1]q

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportConfigurationAdd(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"
  subnet_ids  = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "bar" {
  identifier = %[1]q

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
    "general",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportConfigurationModify(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"
  subnet_ids  = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "bar" {
  identifier = %[1]q

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "general",
    "slowquery",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportConfigurationDelete(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = %[1]q
  description = "db subnets for rds_one"
  subnet_ids  = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "bar" {
  identifier = %[1]q

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_ec2Classic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

# EC2-Classic specific
data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.m3.medium", "db.m3.large", "db.r3.large"]
}

resource "aws_db_instance" "bar" {
  identifier           = %[1]q
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type         = data.aws_rds_orderable_db_instance.test.storage_type
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  publicly_accessible  = true
  security_group_names = ["default"]
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_mariaDB(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_dbSubnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage    = 5
  db_subnet_group_name = aws_db_subnet_group.test.name
  engine               = data.aws_rds_orderable_db_instance.test.engine
  identifier           = %[1]q
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  deletion_protection = %[1]t
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[2]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, deletionProtection, rName))
}

func testAccInstanceConfig_EnabledCloudWatchLogsExports_oracle(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "oracle-se2"
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["alert", "listener", "trace"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  license_model                   = "bring-your-own-license"
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName)
}

func testAccInstanceConfig_NationalCharacterSet_oracle(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "oracle-se2"
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage        = 10
  engine                   = data.aws_rds_orderable_db_instance.test.engine
  identifier               = %[1]q
  instance_class           = data.aws_rds_orderable_db_instance.test.instance_class
  license_model            = "bring-your-own-license"
  nchar_character_set_name = "UTF8"
  password                 = "avoid-plaintext-passwords"
  username                 = "tfacctest"
  skip_final_snapshot      = true
}
`, rName)
}

func testAccInstanceConfig_NoNationalCharacterSet_oracle(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "oracle-se2"
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = "bring-your-own-license"
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName)
}

func testAccInstanceConfig_EnabledCloudWatchLogsExports_mssql(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerSe(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage               = 20
  enabled_cloudwatch_logs_exports = ["agent", "error"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  license_model                   = data.aws_rds_orderable_db_instance.test.license_model
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName))
}

func testAccInstanceConfig_EnabledCloudWatchLogsExports_postgreSQL(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "postgres"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[1]s]
}

resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  engine                          = data.aws_rds_engine_version.default.engine
  identifier                      = %[2]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, postgresPreferredInstanceClasses, rName)
}

func testAccInstanceConfig_maxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage     = 5
  engine                = data.aws_rds_orderable_db_instance.test.engine
  identifier            = %[1]q
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  max_allocated_storage = %[2]d
  password              = "avoid-plaintext-passwords"
  username              = "tfacctest"
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_password(rName, password string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = %[2]q
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName, password))
}

func testAccInstanceConfig_ReplicateSourceDB_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_namePrefix(identifierPrefix, sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier_prefix   = %[2]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, sourceName, identifierPrefix))
}

func testAccInstanceConfig_ReplicateSourceDB_nameGenerated(sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, sourceName))
}

func testAccInstanceConfig_ReplicateSourceDB_addLaterSetup() string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(), `
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}
`)
}

func testAccInstanceConfig_ReplicateSourceDB_addLater(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_ReplicateSourceDB_addLaterSetup(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_allocatedStorage(rName string, allocatedStorage int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = %[2]d
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccInstanceConfig_ReplicateSourceDB_iops(rName string, iops int) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 200
  backup_retention_period = 1
  engine                  = "mysql"
  identifier              = "%[1]s-source"
  instance_class          = "db.t2.micro"
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
  iops                    = 1100
  storage_type            = "io1"
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
  iops                = %[2]d
  storage_type        = "io1"
}
`, rName, iops)
}

func testAccInstanceConfig_ReplicateSourceDB_allocatedStorageAndIOPS(rName string, allocatedStorage, iops int) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = %[2]d
  backup_retention_period = 1
  engine                  = "mysql"
  identifier              = "%[1]s-source"
  instance_class          = "db.t2.micro"
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
  iops                    = 1000
  storage_type            = "io1"
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
  iops                = %[3]d
  storage_type        = "io1"
}
`, rName, allocatedStorage, iops)
}

func testAccInstanceConfig_ReplicateSourceDB_allowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[2]t
  identifier                  = %[1]q
  instance_class              = aws_db_instance.source.instance_class
  replicate_source_db         = aws_db_instance.source.id
  skip_final_snapshot         = true
}
`, rName, allowMajorVersionUpgrade))
}

func testAccInstanceConfig_ReplicateSourceDB_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  replicate_source_db        = aws_db_instance.source.id
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccInstanceConfig_ReplicateSourceDB_availabilityZone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_backupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  replicate_source_db     = aws_db_instance.source.id
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_ReplicateSourceDB_backupWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_ReplicateSourceDB_dbSubnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_engine_version" "default" {
  provider = "awsalternate"
  engine   = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_engine_version.default.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  replicate_source_db  = aws_db_instance.source.arn
  skip_final_snapshot  = true
}
`, rName, mySQLPreferredInstanceClasses))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func testAccAlternateAccountAndAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternateaccountalternateregion" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}

provider "awsalternateaccountsameregion" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[4]q
}

provider "awssameaccountalternateregion" {
  region = %[3]q
}
`, os.Getenv(conns.EnvVarAlternateAccessKeyId), os.Getenv(conns.EnvVarAlternateProfile), acctest.AlternateRegion(), os.Getenv(conns.EnvVarAlternateSecretAccessKey))
}

func testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(testAccAlternateAccountAndAlternateRegionProviderConfig(), fmt.Sprintf(`
data "aws_availability_zones" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_availability_zones" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "sameaccountalternateregion" {
  count    = 2
  provider = "awssameaccountalternateregion"

  availability_zone = data.aws_availability_zones.sameaccountalternateregion.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.sameaccountalternateregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  availability_zone = data.aws_availability_zones.alternateaccountsameregion.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.alternateaccountsameregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  name = %[1]q
}

resource "aws_ram_principal_association" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.arn
}

resource "aws_ram_resource_association" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  resource_arn       = aws_subnet.alternateaccountsameregion[count.index].arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.id
}

resource "aws_db_subnet_group" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  name       = %[1]q
  subnet_ids = aws_subnet.sameaccountalternateregion[*].id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name       = %[1]q
  subnet_ids = aws_subnet.alternateaccountsameregion[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name   = %[1]q
  vpc_id = aws_vpc.alternateaccountsameregion.id
}

data "aws_rds_engine_version" "default" {
  provider = "awssameaccountalternateregion"
  engine   = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awssameaccountalternateregion"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  provider = "awssameaccountalternateregion"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.sameaccountalternateregion.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName, mySQLPreferredInstanceClasses))
}

func testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_engine_version" "default" {
  provider = "awsalternate"
  engine   = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName, mySQLPreferredInstanceClasses))
}

func testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccInstanceConfig_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  replicate_source_db                 = aws_db_instance.source.id
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_ReplicateSourceDB_maintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_ReplicateSourceDB_maxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  replicate_source_db   = aws_db_instance.source.id
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_ReplicateSourceDB_monitoring(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_MonitoringRole(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccInstanceConfig_ReplicateSourceDB_multiAZ(rName string, multiAz bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_sameSetOnBoth(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.test.id
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  replicate_source_db  = aws_db_instance.source.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_differentSetOnBoth(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  replicate_source_db  = aws_db_instance.source.id
  skip_final_snapshot  = true
}

resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.source.id
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_parameter_group" "source" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = "%[1]s-source"

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_replicaCopiesValue(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.test.id
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_setOnReplica(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  replicate_source_db  = aws_db_instance.source.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_port(rName string, port int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccInstanceConfig_ReplicateSourceDB_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_caCertificateID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  ca_cert_identifier      = data.aws_rds_certificate.latest.id
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_replicaMode(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "bring-your-own-license"
  storage_type   = "gp2"

  preferred_instance_classes = [%[1]s]
}

resource "aws_db_instance" "source" {
  identifier              = "%[2]s-source"
  allocated_storage       = 20
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[2]q
  instance_class      = aws_db_instance.source.instance_class
  replica_mode        = "mounted"
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, oraclePreferredInstanceClasses, rName)
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "bring-your-own-license"
  storage_type   = "gp2"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 3

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  character_set_name   = "AL32UTF8"
  timeouts {
    update = "120m"
  }
  ca_cert_identifier = "rds-ca-2019"
}
`, rName, oraclePreferredInstanceClasses)
}

func testAccInstanceConfig_ReplicateSourceDB_parameterGroupTwoStep(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  replicate_source_db = aws_db_instance.source.id

  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  apply_immediately   = true

  parameter_group_name = aws_db_parameter_group.test.id
  ca_cert_identifier   = "rds-ca-2019"

  timeouts {
    update = "120m"
  }
}

resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q
}
`, rName))
}

func testAccInstanceConfig_MonitoringRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccInstanceConfig_snapshotID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotIdentifier_namePrefix(identifierPrefix, sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier_prefix   = %[2]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, sourceName, identifierPrefix))
}

func testAccInstanceConfig_SnapshotIdentifier_nameGenerated(sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, sourceName))
}

func testAccInstanceConfig_SnapshotID_associationRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_allocatedStorage(rName string, allocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccInstanceConfig_SnapshotID_io1Storage(rName string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mariadb"
  engine_version = "10.5.12"
  license_model  = "general-public-license"
  storage_type   = "io1"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 200
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
  allocated_storage   = 200
  iops                = %[2]d
  storage_type        = data.aws_rds_orderable_db_instance.test.storage_type
}
`, rName, iops)
}

func testAccInstanceConfig_SnapshotID_allowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "postgres13" {
  engine         = "postgres"
  engine_version = "13.5"
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.postgres13.engine
  engine_version      = data.aws_rds_orderable_db_instance.postgres13.engine_version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.postgres13.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

data "aws_rds_orderable_db_instance" "postgres14" {
  engine         = "postgres"
  engine_version = "14.1"
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[2]t
  engine                      = data.aws_rds_orderable_db_instance.postgres14.engine
  engine_version              = data.aws_rds_orderable_db_instance.postgres14.engine_version
  identifier                  = %[1]q
  instance_class              = aws_db_instance.source.instance_class
  snapshot_identifier         = aws_db_snapshot.test.id
  skip_final_snapshot         = true
}
`, rName, allowMajorVersionUpgrade)
}

func testAccInstanceConfig_SnapshotID_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  snapshot_identifier        = aws_db_snapshot.test.id
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccInstanceConfig_SnapshotID_availabilityZone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_backupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

func testAccInstanceConfig_SnapshotID_BackupRetentionPeriod_unset(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = 0
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_SnapshotID_backupWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_SnapshotID_dbSubnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccInstanceConfig_SnapshotID_iamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  snapshot_identifier                 = aws_db_snapshot.test.id
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_SnapshotID_maintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_SnapshotID_maxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  snapshot_identifier   = aws_db_snapshot.test.id
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_SnapshotID_monitoring(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccInstanceConfig_SnapshotID_multiAZ(rName string, multiAz bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_SnapshotID_MultiAZ_sqlServer(rName string, multiAz bool) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerSe(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  # InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
  backup_retention_period = 1
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  multi_az                = %[2]t
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_SnapshotID_parameterGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_port(rName string, port int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccInstanceConfig_SnapshotID_tags(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  tags = {
    key1 = "value-old"
  }
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_Tags_clear(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {}
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_VPCSecurityGroupIDs_tags(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccInstanceConfig_performanceInsightsDisabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 5
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "mydb"
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
  username                = "foo"
}
`, rName)
}

func testAccInstanceConfig_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[1]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                               = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName)
}

func testAccInstanceConfig_performanceInsightsKMSKeyIdDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  engine                  = data.aws_rds_engine_version.default.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 5
  backup_retention_period = 0
  db_name                 = "mydb"
  username                = "foo"
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
}
`, rName)
}

func testAccInstanceConfig_performanceInsightsKMSKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[1]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                               = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName)
}

func testAccInstanceConfig_performanceInsightsRetentionPeriod(rName string, performanceInsightsRetentionPeriod int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[1]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                               = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = %[2]d
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName, performanceInsightsRetentionPeriod)
}

func testAccInstanceConfig_ReplicateSourceDB_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "mustbeeightcharaters"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier                            = %[1]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  replicate_source_db                   = aws_db_instance.source.id
  skip_final_snapshot                   = true
}
`, rName)
}

func testAccInstanceConfig_SnapshotID_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier                            = %[1]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  snapshot_identifier                   = aws_db_snapshot.test.id
  skip_final_snapshot                   = true
}
`, rName)
}

func testAccInstanceConfig_noDeleteAutomatedBackups(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  backup_retention_period  = 1
  delete_automated_backups = false
}
`, rName))
}

func testAccInstanceConfig_Outpost_coIPEnabled(rName string, coipEnabled bool, backupRetentionPeriod int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "foo" {
  cidr_block = "10.128.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.128.1.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  vpc_id            = aws_vpc.foo.id
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "foo" {
  name       = %[1]q
  subnet_ids = [aws_subnet.foo.id]

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = data.aws_outposts_outpost.test.arn
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.foo.id
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.m5.large", "db.m5.xlarge", "db.r5.large", "db.r5.xlarge"]
}

resource "aws_db_instance" "test" {
  identifier                = %[1]q
  allocated_storage         = 20
  backup_retention_period   = %[3]d
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class            = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                   = "baz"
  parameter_group_name      = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                  = "barbarbarbar"
  skip_final_snapshot       = true
  username                  = "foo"
  db_subnet_group_name      = aws_db_subnet_group.foo.name
  storage_encrypted         = true
  customer_owned_ip_enabled = %[2]t
}
`, rName, coipEnabled, backupRetentionPeriod)
}

func testAccInstanceConfig_CoIPEnabled_restorePointInTime(rName string, sourceCoipEnabled bool, targetCoipEnabled bool) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_Outpost_coIPEnabled(rName, sourceCoipEnabled, 1),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "%[1]s-restore"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }
  storage_encrypted         = true
  skip_final_snapshot       = true
  db_subnet_group_name      = aws_db_instance.test.db_subnet_group_name
  customer_owned_ip_enabled = %[2]t
}
`, rName, targetCoipEnabled))
}

func testAccInstanceConfig_CoIPEnabled_snapshotID(rName string, sourceCoipEnabled bool, targetCoipEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_Outpost_coIPEnabled(rName, sourceCoipEnabled, 1), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "restore" {
  customer_owned_ip_enabled = %[2]t
  db_subnet_group_name      = aws_db_subnet_group.foo.name
  storage_encrypted         = true
  identifier                = "%[1]s-restore"
  instance_class            = aws_db_instance.test.instance_class
  snapshot_identifier       = aws_db_snapshot.test.id
  skip_final_snapshot       = true
}
`, rName, targetCoipEnabled))
}

func testAccInstanceConfig_license(rName, license string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClass("oracle-se2", license, "standard", oracleSE2PreferredInstanceClasses),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  apply_immediately   = true
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}
