package rds_test

import (
	"errors"
	"fmt"
	"regexp"
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
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestClusterIDRegionFromARN(t *testing.T) {
	testCases := []struct {
		TestName       string
		Input          string
		ExpectedID     string
		ExpectedRegion string
		ExpectedErr    bool
	}{
		{
			TestName:       "empty",
			Input:          "",
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
		{
			TestName:       "normal ARN",
			Input:          "arn:aws:rds:us-west-2:012345678901:cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-west-2", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "another good ARN",
			Input:          "arn:aws:rds:us-east-1:012345678901:cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-east-1", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "no account",
			Input:          "arn:aws:rds:us-east-2::cluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "tf-acc-test-1467354933239945971",
			ExpectedRegion: "us-east-2", // lintignore:AWSAT003
			ExpectedErr:    false,
		},
		{
			TestName:       "wrong service",
			Input:          "arn:aws:connect:us-west-2:012345678901:instance/1032bdc4-d72c-5490-a9fa-3c9b4dba67bb", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
		{
			TestName:       "wrong resource",
			Input:          "arn:aws:rds:us-east-2::notacluster:tf-acc-test-1467354933239945971", // lintignore:AWSAT003,AWSAT005
			ExpectedID:     "",
			ExpectedRegion: "",
			ExpectedErr:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotID, gotRegion, gotErr := tfrds.ClusterIDRegionFromARN(testCase.Input)

			if gotErr != nil && !testCase.ExpectedErr {
				t.Errorf("got no error, expected one: %s", testCase.Input)
			}

			if gotID != testCase.ExpectedID {
				t.Errorf("got %s, expected %s", gotID, testCase.ExpectedID)
			}

			if gotRegion != testCase.ExpectedRegion {
				t.Errorf("got %s, expected %s", gotRegion, testCase.ExpectedRegion)
			}
		})
	}
}

func TestAccRDSGlobalCluster_basic(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", ""),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", rName),
					resource.TestMatchResourceAttr(resourceName, "global_cluster_resource_id", regexp.MustCompile(`cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
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

func TestAccRDSGlobalCluster_disappears(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckGlobalClusterDisappears(&globalCluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSGlobalCluster_databaseName(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database2"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_deletionProtection(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_Engine_aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engine(rName, "aurora"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
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

func TestAccRDSGlobalCluster_EngineVersion_aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "aurora"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMinor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				//Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora", "5.6.mysql_aurora.1.22.2"),
				Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora-postgresql", "13.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				//Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora", "5.6.mysql_aurora.1.23.2"),
				Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora-postgresql", "13.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "13.5"),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMajor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora", "5.6.mysql_aurora.1.23.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				Config: testAccGlobalClusterConfig_primaryEngineVersion(rName, "aurora-mysql", "5.7.mysql_aurora.2.10.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.10.2"),
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

func TestAccRDSGlobalCluster_EngineVersion_updateMinorMultiRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rNameGlobal := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // don't need to be unique but makes debugging easier
	rNamePrimary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameSecondary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersionUpgradeMultiRegion(rNameGlobal, rNamePrimary, rNameSecondary, "aurora-mysql", "5.7.mysql_aurora.2.07.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				Config: testAccGlobalClusterConfig_engineVersionUpgradeMultiRegion(rNameGlobal, rNamePrimary, rNameSecondary, "aurora-mysql", "5.7.mysql_aurora.2.07.7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.07.7"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_EngineVersion_updateMajorMultiRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rNameGlobal := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // don't need to be unique but makes debugging easier
	rNamePrimary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameSecondary := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersionUpgradeMultiRegion(rNameGlobal, rNamePrimary, rNameSecondary, "aurora", "5.6.mysql_aurora.1.23.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
				),
			},
			{
				Config: testAccGlobalClusterConfig_engineVersionUpgradeMultiRegion(rNameGlobal, rNamePrimary, rNameSecondary, "aurora-mysql", "5.7.mysql_aurora.2.10.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.10.2"),
				),
			},
		},
	})
}

func TestAccRDSGlobalCluster_EngineVersion_auroraMySQL(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "aurora-mysql"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
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

func TestAccRDSGlobalCluster_EngineVersion_auroraPostgreSQL(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "aurora-postgresql"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", "data.aws_rds_engine_version.default", "version"),
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

func TestAccRDSGlobalCluster_sourceDBClusterIdentifier(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceClusterID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_SourceDBClusterIdentifier_storageEncrypted(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceClusterIDStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccRDSGlobalCluster_storageEncrypted(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalCluster(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckGlobalClusterExists(resourceName string, globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Global Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		cluster, err := tfrds.DescribeGlobalCluster(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("RDS Global Cluster not found")
		}

		if aws.StringValue(cluster.Status) != "available" {
			return fmt.Errorf("RDS Global Cluster (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(cluster.Status))
		}

		*globalCluster = *cluster

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_global_cluster" {
			continue
		}

		globalCluster, err := tfrds.DescribeGlobalCluster(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if globalCluster == nil {
			continue
		}

		return fmt.Errorf("RDS Global Cluster (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(globalCluster.Status))
	}

	return nil
}

func testAccCheckGlobalClusterDisappears(globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		input := &rds.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
		}

		_, err := conn.DeleteGlobalCluster(input)

		if err != nil {
			return err
		}

		return tfrds.WaitForGlobalClusterDeletion(conn, aws.StringValue(globalCluster.GlobalClusterIdentifier), tfrds.GlobalClusterRemovalTimeout)
	}
}

func testAccCheckGlobalClusterNotRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterArn) != aws.StringValue(j.GlobalClusterArn) {
			return fmt.Errorf("RDS Global Cluster was recreated. got: %s, expected: %s", aws.StringValue(i.GlobalClusterArn), aws.StringValue(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckGlobalClusterRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("RDS Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckGlobalCluster(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
}
`, rName)
}

func testAccGlobalClusterConfig_databaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  database_name             = %[1]q
  global_cluster_identifier = %[2]q
}
`, databaseName, rName)
}

func testAccGlobalClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  deletion_protection       = %[1]t
  global_cluster_identifier = %[2]q
}
`, deletionProtection, rName)
}

func testAccGlobalClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %[1]q
  global_cluster_identifier = %[2]q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_engineVersion(rName, engine string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

resource "aws_rds_global_cluster" "test" {
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
  global_cluster_identifier = %[2]q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_primaryEngineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = %[1]q
  engine_version             = %[2]q
  preferred_instance_classes = ["db.r5.large", "db.r5.xlarge", "db.r6g.large"] # Aurora global db may be limited to rx
}

resource "aws_rds_global_cluster" "test" {
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  global_cluster_identifier = %[3]q
}

resource "aws_rds_cluster" "test" {
  apply_immediately           = true
  allow_major_version_upgrade = true
  cluster_identifier          = %[3]q
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  master_password             = "mustbeeightcharacters"
  master_username             = "test"
  skip_final_snapshot         = true

  global_cluster_identifier = aws_rds_global_cluster.test.id

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_orderable_db_instance.test.engine
  engine_version     = data.aws_rds_orderable_db_instance.test.engine_version
  identifier         = %[3]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, engine, engineVersion, rName)
}

func testAccGlobalClusterConfig_engineVersionUpgradeMultiRegion(rNameGlobal, rNamePrimary, rNameSecondary, engine, engineVersion string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "alternate" {
  provider = "awsalternate"
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = %[2]q
  engine_version            = %[3]q
}

resource "aws_rds_cluster" "primary" {
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[4]q
  database_name               = "totoro"
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  master_password             = "satsukimae"
  master_username             = "maesatsuki"
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_instance" "primary" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.primary.id
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
  identifier         = %[4]q
  instance_class     = "db.r4.large"
}

resource "aws_vpc" "alternate" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[5]q
  }
}

resource "aws_subnet" "alternate" {
  provider          = "awsalternate"
  count             = 3
  vpc_id            = aws_vpc.alternate.id
  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"

  tags = {
    Name = %[5]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider   = "awsalternate"
  name       = %[5]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_rds_cluster" "secondary" {
  provider                    = "awsalternate"
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = %[5]q
  engine                      = aws_rds_global_cluster.test.engine
  engine_version              = aws_rds_global_cluster.test.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.test.id
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [
      replication_source_identifier,
      engine_version,
    ]
  }

  depends_on = [aws_rds_cluster_instance.primary]
}

resource "aws_rds_cluster_instance" "secondary" {
  provider           = "awsalternate"
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.secondary.id
  engine             = aws_rds_cluster.secondary.engine
  engine_version     = aws_rds_cluster.secondary.engine_version
  identifier         = %[5]q
  instance_class     = "db.r4.large"
}
`, rNameGlobal, engine, engineVersion, rNamePrimary, rNameSecondary))
}

func testAccGlobalClusterConfig_sourceClusterID(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "aurora-postgresql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_sourceClusterIDStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "aurora-postgresql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true
  storage_encrypted   = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_storageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  storage_encrypted         = %[2]t
}
`, rName, storageEncrypted)
}
