package neptune_test

import (
	//"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
)

func TestAccNeptuneCluster_basic(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "neptune_cluster_parameter_group_name", "default.neptune1"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_copyTagsToSnapshot(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_copyTags(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
			{
				Config: testAccClusterConfig_copyTags(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
				),
			},
			{
				Config: testAccClusterConfig_copyTags(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
		},
	})
}

func TestAccNeptuneCluster_namePrefix(t *testing.T) {
	var v neptune.DBCluster
	rName := "tf-test-"
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_takeFinalSnapshot(t *testing.T) {
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterSnapshot(rName),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_tags(t *testing.T) {
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
			{
				Config: testAccClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNeptuneCluster_updateIAMRoles(t *testing.T) {
	var v neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_includingIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
				),
			},
			{
				Config: testAccClusterConfig_addIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccClusterConfig_removeIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_kmsKey(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	keyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_encrypted(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_backupsUpdate(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_backups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccClusterConfig_backupsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_iamAuth(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamAuth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_updateCloudWatchLogsExports(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckNoResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#"),
				),
			},
			{
				Config: testAccClusterConfig_cloudWatchLogsExports(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enable_cloudwatch_logs_exports.*", "audit"),
				),
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_updateEngineVersion(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.0.2.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.0.2.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.0.5.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.0.5.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_updateEngineMajorVersion(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersion(rName, "1.0.2.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.0.2.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
			{
				Config: testAccClusterConfig_engineMajorVersionUpdate(rName, "1.1.1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.1.1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
		},
	})
}

func TestAccNeptuneCluster_deleteProtection(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"skip_final_snapshot",
					"allow_major_version_upgrade",
				},
			},
			{
				Config: testAccClusterConfig_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccClusterConfig_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccNeptuneCluster_disappears(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dbCluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfneptune.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	return testAccCheckClusterDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).NeptuneConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&neptune.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
				return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if err != nil {
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckClusterExists(n string, v *neptune.DBCluster) resource.TestCheckFunc {
	return testAccCheckClusterExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterExistsWithProvider(n string, v *neptune.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).NeptuneConn
		resp, err := conn.DescribeDBClusters(&neptune.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, c := range resp.DBClusters {
			if *c.DBClusterIdentifier == rs.Primary.ID {
				*v = *c
				return nil
			}
		}

		return fmt.Errorf("Neptune Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckClusterSnapshot(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			awsClient := acctest.Provider.Meta().(*conns.AWSClient)
			conn := awsClient.NeptuneConn

			log.Printf("[INFO] Deleting the Snapshot %s", rName)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&neptune.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(rName),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&neptune.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
					return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the cluster is already destroyed
			if err != nil {
				if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
					return nil
				}
			}

			return err
		}

		return nil
	}
}

func testAccClusterBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}
`)
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterConfig_copyTags(rName string, copy bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
  copy_tags_to_snapshot                = %[2]t
}
`, rName, copy))
}

func testAccClusterConfig_deleteProtection(rName string, isProtected bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
  deletion_protection                  = %t
}
`, rName, isProtected))
}

func testAccClusterConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix            = %q
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName)
}

func testAccClusterConfig_finalSnapshot(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  final_snapshot_identifier            = %[1]q
}
`, rName))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_includingIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccClusterConfig_addIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn, aws_iam_role.test-2.arn]

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccClusterConfig_removeIAMRoles(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn]

  depends_on = [aws_iam_role.test]
}
`, rName))
}

func testAccClusterConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`

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

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  storage_encrypted                    = true
  kms_key_arn                          = aws_kms_key.test.arn
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterConfig_encrypted(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %q
  availability_zones  = local.availability_zone_names
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_backups(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, rName))
}

func testAccClusterConfig_backupsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, rName))
}

func testAccClusterConfig_iamAuth(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                  = %q
  availability_zones                  = local.availability_zone_names
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_cloudWatchLogsExports(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier             = %q
  availability_zones             = local.availability_zone_names
  skip_final_snapshot            = true
  enable_cloudwatch_logs_exports = ["audit"]
}
`, rName))
}

func testAccClusterConfig_engineVersionBase(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(), fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine         = "neptune"
  engine_version = aws_neptune_cluster.test.engine_version
  license_model  = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "test" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.test.id
  apply_immediately            = true
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  neptune_parameter_group_name = aws_neptune_cluster.test.neptune_cluster_parameter_group_name
  promotion_tier               = "3"
}
`, rName))
}

func testAccClusterConfig_engineVersion(rName, engineVersion string) string {
	return acctest.ConfigCompose(testAccClusterConfig_engineVersionBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  apply_immediately                    = true
  availability_zones                   = local.availability_zone_names
  engine_version                       = %[2]q
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName, engineVersion))
}

func testAccClusterConfig_engineMajorVersionUpdate(rName, engineVersion string) string {
	return acctest.ConfigCompose(testAccClusterConfig_engineVersionBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  apply_immediately                    = true
  availability_zones                   = local.availability_zone_names
  engine_version                       = %[2]q
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
  allow_major_version_upgrade          = true
}
`, rName, engineVersion))
}
