package rds_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSClusterInstance_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					testAccCheckClusterInstanceAttributes(&v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceModifiedConfig(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					testAccCheckClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_isAlreadyBeingDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(fmt.Sprintf("tf-cluster-instance-%d", rInt)),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccClusterInstanceConfig(rInt),
				Destroy: true,
			},
		},
	})
}

func TestAccRDSClusterInstance_az(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_az(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					testAccCheckClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", availabilityZonesDataSourceName, "names.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_namePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rInt := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					testAccCheckClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", fmt.Sprintf("tf-test-%d", rInt)),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-cluster-instance-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_generatedName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_generatedName(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					testAccCheckClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_kmsKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceKMSKeyConfig(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/5350
func TestAccRDSClusterInstance_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceClusterInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterInstance_publiclyAccessible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_PubliclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceConfig_PubliclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_copyTagsToSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rNameSuffix := sdkacctest.RandInt()
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_monitoringInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceMonitoringIntervalConfig(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceMonitoringIntervalConfig(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccClusterInstanceMonitoringIntervalConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccClusterInstanceMonitoringIntervalConfig(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_enabledToDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceMonitoringRoleARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceMonitoringIntervalConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_enabledToRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceMonitoringRoleARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceMonitoringRoleARNRemovedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_removedToEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceMonitoringRoleARNRemovedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstanceMonitoringRoleARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsEnabled_auroraMySQL1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL1Config(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsEnabled_auroraMySQL2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsPreCheck(t, engine, engineVersion) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL2Config(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsEnabled_auroraPostgresql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraPostgresqlConfig(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyID_auroraMySQL1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL1Config(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyIDAuroraMySQL1_defaultKeyToCustomKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL1Config(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL1Config(rName, engine),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You .* change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyID_auroraMySQL2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsPreCheck(t, engine, engineVersion) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL2Config(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyIDAuroraMySQL2_defaultKeyToCustomKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsPreCheck(t, engine, engineVersion) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL2Config(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL2Config(rName, engine, engineVersion),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You .* change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccRDSClusterInstance_performanceInsightsRetentionPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, "aurora") },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsRetentionPeriodConfig(rName, 731),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "731"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccClusterInstancePerformanceInsightsRetentionPeriodConfig(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyID_auroraPostgresql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraPostgresqlConfig(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyIDAuroraPostgresql_defaultKeyToCustomKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstancePerformanceInsightsEnabledAuroraPostgresqlConfig(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraPostgresqlConfig(rName, engine),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You .* change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccRDSClusterInstance_caCertificateIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_CACertificateIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func testAccPerformanceInsightsDefaultVersionPreCheck(t *testing.T, engine string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBEngineVersionsInput{
		DefaultOnly: aws.Bool(true),
		Engine:      aws.String(engine),
	}

	result, err := conn.DescribeDBEngineVersions(input)
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(result.DBEngineVersions) < 1 {
		t.Fatalf("unexpected PreCheck error, no default version for engine: %s", engine)
	}

	testAccPerformanceInsightsPreCheck(t, engine, aws.StringValue(result.DBEngineVersions[0].EngineVersion))
}

func testAccPerformanceInsightsPreCheck(t *testing.T, engine string, engineVersion string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion), // version cuts response time
	}

	supportsPerformanceInsights := false
	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *rds.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			if aws.BoolValue(instanceOption.SupportsPerformanceInsights) {
				supportsPerformanceInsights = true
				return false // stop processing pages
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, "InvalidParameterCombination") {
		t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if !supportsPerformanceInsights {
		t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
	}
}

func testAccCheckClusterInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != "aurora" && *v.Engine != "aurora-postgresql" && *v.Engine != "aurora-mysql" {
			return fmt.Errorf("bad engine, expected \"aurora\", \"aurora-mysql\" or \"aurora-postgresql\": %#v", *v.Engine)
		}

		if !strings.HasPrefix(*v.DBClusterIdentifier, "tf-aurora-cluster") {
			return fmt.Errorf("Bad Cluster Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster", *v.DBClusterIdentifier)
		}

		return nil
	}
}

func testAccCheckClusterInstanceExists(n string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Cluster Instance ID is set")
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

func testAccCheckClusterInstanceDestroy(s *terraform.State) error {
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

		return fmt.Errorf("RDS Cluster Instance %s still exists", rs.Primary.ID)
	}

	return nil
}

// Add some random to the name, to avoid collision
func testAccClusterInstanceConfig(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n))
}

func testAccClusterInstanceModifiedConfig(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier                 = "tf-cluster-instance-%[1]d"
  cluster_identifier         = aws_rds_cluster.default.id
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name    = aws_db_parameter_group.bar.name
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n))
}

func testAccClusterInstanceConfig_az(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
  availability_zone       = data.aws_availability_zones.available.names[0]
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n))
}

func testAccClusterInstanceConfig_namePrefix(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  identifier_prefix  = "tf-cluster-instance-"
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%[1]d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccClusterInstanceConfig_generatedName(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%[1]d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccClusterInstanceKMSKeyConfig(n int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %[1]d"

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

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.foo.arn
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n))
}

func testAccClusterInstanceMonitoringIntervalConfig(rName string, monitoringInterval int) string {
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName, monitoringInterval)
}

func testAccClusterInstanceMonitoringRoleARNRemovedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, rName)
}

func testAccClusterInstanceMonitoringRoleARNConfig(rName string) string {
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL1Config(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine)
}

func testAccClusterInstancePerformanceInsightsEnabledAuroraMySQL2Config(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  engine_version      = %[3]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  engine_version               = aws_rds_cluster.test.engine_version
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine, engineVersion)
}

func testAccClusterInstancePerformanceInsightsEnabledAuroraPostgresqlConfig(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine)
}

func testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL1Config(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine)
}

func testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraMySQL2Config(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  engine_version      = %[3]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  engine_version                  = aws_rds_cluster.test.engine_version
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine, engineVersion)
}

func testAccClusterInstancePerformanceInsightsKMSKeyIdAuroraPostgresqlConfig(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine)
}

func testAccClusterInstancePerformanceInsightsRetentionPeriodConfig(rName string, performanceInsightsRetentionPeriod int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier                    = aws_rds_cluster.test.id
  engine                                = aws_rds_cluster.test.engine
  identifier                            = %[1]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled          = true
  performance_insights_retention_period = %[2]d
}
`, rName, performanceInsightsRetentionPeriod)
}

func testAccClusterInstanceConfig_PubliclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately   = true
  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  publicly_accessible = %[2]t
}
`, rName, publiclyAccessible)
}

func testAccClusterInstanceConfig_CopyTagsToSnapshot(n int, f bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier            = "tf-cluster-instance-%[1]d"
  cluster_identifier    = aws_rds_cluster.default.id
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  promotion_tier        = "3"
  copy_tags_to_snapshot = %t
}
`, n, f))
}

func testAccClusterInstanceConfig_CACertificateIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  ca_cert_identifier = data.aws_rds_certificate.latest.id
}
`, rName)
}
