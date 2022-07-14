package redshift_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftCluster_basic(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_nodes.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_nodes.0.public_ip_address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(fmt.Sprintf("^%s.*\\.redshift\\..*", rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "auto"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_track_name", "current"),
					resource.TestCheckResourceAttr(resourceName, "manual_snapshot_retention_period", "-1"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccRedshiftCluster_aqua(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_aqua(rName, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
			{
				Config: testAccClusterConfig_aqua(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "disabled"),
				),
			},
			{
				Config: testAccClusterConfig_aqua(rName, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "enabled"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_disappears(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftCluster_withFinalSnapshot(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyClusterSnapshot(rName),
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
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccRedshiftCluster_kmsKey(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	keyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", keyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccRedshiftCluster_enhancedVPCRoutingEnabled(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enhancedVPCRoutingEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enhanced_vpc_routing", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
			{
				Config: testAccClusterConfig_enhancedVPCRoutingDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enhanced_vpc_routing", "false"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_loggingEnabled(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	bucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_loggingEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging.0.enable", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.bucket_name", bucketResourceName, "bucket"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
			{
				Config: testAccClusterConfig_loggingDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging.0.enable", "false"),
				),
			},
			{
				Config: testAccClusterConfig_loggingCloudWatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.log_destination_type", "cloudwatch"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_snapshotCopy(t *testing.T) {
	var providers []*schema.Provider
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotCopyEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_copy.0.destination_region", "data.aws_region.alternate", "name"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_copy.0.retention_period", "1"),
				),
			},
			{
				Config: testAccClusterConfig_snapshotCopyDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "snapshot_copy.#", "0"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_iamRoles(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccClusterConfig_updateIAMRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_publiclyAccessible(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_publiclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},

			{
				Config: testAccClusterConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_updateNodeCount(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "1"),
				),
			},
			{
				Config: testAccClusterConfig_updateNodeCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.large"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_updateNodeType(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateNodeType(rName, "dc2.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.large"),
				),
			},
			{
				Config: testAccClusterConfig_updateNodeType(rName, "dc2.8xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.8xlarge"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_tags(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
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
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
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

func TestAccRedshiftCluster_forceNewUsername(t *testing.T) {
	var v1, v2 redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v1),
					testAccCheckClusterMasterUsername(&v1, "foo_test"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "foo_test"),
				),
			},
			{
				Config: testAccClusterConfig_updatedUsername(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v2),
					testAccCheckClusterRecreated(&v1, &v2),
					testAccCheckClusterMasterUsername(&v2, "new_username"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "new_username"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZone(t *testing.T) {
	var v1, v2 redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v2),
					testAccCheckClusterNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.1"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZoneAndSetAvailabilityZoneRelocation(t *testing.T) {
	var v1, v2 redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v2),
					testAccCheckClusterNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.1"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZone_availabilityZoneRelocationNotSet(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config:      testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 1),
				ExpectError: regexp.MustCompile("cannot change `availability_zone` if `availability_zone_relocation_enabled` is not true"),
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption1(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
				),
			},
			{
				Config: testAccClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption2(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
				),
			},
			{
				Config: testAccClusterConfig_unencrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_availabilityZoneRelocation(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_availabilityZoneRelocation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"apply_immediately",
				},
			},
			{
				Config: testAccClusterConfig_availabilityZoneRelocation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", "false"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_availabilityZoneRelocation_publiclyAccessible(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_availabilityZoneRelocationPubliclyAccessible(rName),
				ExpectError: regexp.MustCompile("`availability_zone_relocation_enabled` cannot be true when `publicly_accessible` is true"),
			},
		},
	})
}

func TestAccRedshiftCluster_restoreFromSnapshot(t *testing.T) {
	var v redshift.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyClusterSnapshot(rName),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_createSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.8xlarge"),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
				),
			},
			// Apply a configuration without the source cluster to ensure final snapshot creation.
			{ // nosemgrep:test-config-funcs-correct-form
				Config: acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
			},
			{
				Config: testAccClusterConfig_restoreFromSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "data.aws_availability_zones.available", "names.1"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.large"),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
					"apply_immediately",
				},
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_cluster" {
			continue
		}

		_, err := tfredshift.FindClusterByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDestroyClusterSnapshot(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

			_, err := conn.DeleteClusterSnapshot(
				&redshift.DeleteClusterSnapshotInput{
					SnapshotIdentifier: aws.String(rName),
				})

			if err != nil {
				return fmt.Errorf("error deleting Redshift Cluster Snapshot (%s): %w", rName, err)
			}

			_, err = tfredshift.FindClusterByID(conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(n string, v *redshift.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		output, err := tfredshift.FindClusterByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterMasterUsername(c *redshift.Cluster, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got, want := aws.StringValue(c.MasterUsername), value; got != want {
			return fmt.Errorf("Expected cluster's MasterUsername: %q, given: %q", want, got)
		}
		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *redshift.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// In lieu of some other uniquely identifying attribute from the API that always changes
		// when a cluster is destroyed and recreated with the same identifier, we use the SSH key
		// as it will get regenerated when a cluster is destroyed.
		// Certain update operations (e.g KMS encrypting a cluster) will change ClusterCreateTime.
		// Clusters with the same identifier can/will have an overlapping Endpoint.Address.
		if aws.StringValue(i.ClusterPublicKey) != aws.StringValue(j.ClusterPublicKey) {
			return errors.New("Redshift Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *redshift.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// In lieu of some other uniquely identifying attribute from the API that always changes
		// when a cluster is destroyed and recreated with the same identifier, we use the SSH key
		// as it will get regenerated when a cluster is destroyed.
		// Certain update operations (e.g KMS encrypting a cluster) will change ClusterCreateTime.
		// Clusters with the same identifier can/will have an overlapping Endpoint.Address.
		if aws.StringValue(i.ClusterPublicKey) == aws.StringValue(j.ClusterPublicKey) {
			return errors.New("Redshift Cluster was not recreated")
		}

		return nil
	}
}

func testAccClusterConfig_updateNodeCount(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_updateNodeType(rName, nodeType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = %[2]q
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rName, nodeType))
}

func testAccClusterConfig_basic(rName string) string {
	// "InvalidVPCNetworkStateFault: The requested AZ us-west-2a is not a valid AZ."
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_aqua(rName, status string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  aqua_configuration_status           = %[2]q
  apply_immediately                   = true
}
`, rName, status))
}

func testAccClusterConfig_encrypted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_kms_key" "test" {
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

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  encrypted                           = true
  kms_key_id                          = aws_kms_key.test.arn
}
`, rName))
}

func testAccClusterConfig_unencrypted(rName string) string {
	// This is used along with the terraform config created testAccClusterConfig_encrypted, to test removal of encryption.
	//Removing the kms key here causes the key to be deleted before the redshift cluster is unencrypted, resulting in an unstable cluster. This is to be kept for the time-being unti we find a better way to handle this.
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_kms_key" "test" {
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

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_finalSnapshot(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = false
  final_snapshot_identifier           = %[1]q
}
`, rName))
}

func testAccClusterConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_kms_key" "test" {
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

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  kms_key_id                          = aws_kms_key.test.arn
  encrypted                           = true
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_enhancedVPCRoutingEnabled(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = true
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_enhancedVPCRoutingDisabled(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_loggingDisabled(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable = false
  }

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_loggingEnabled(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_redshift_service_account" "main" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "Stmt1376526643067",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_redshift_service_account.main.arn}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    },
    {
      "Sid": "Stmt137652664067",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_redshift_service_account.main.arn}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s"
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable      = true
    bucket_name = aws_s3_bucket.test.bucket
  }

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_loggingCloudWatch(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable               = true
    log_destination_type = "cloudwatch"
    log_exports          = ["connectionlog"]
  }

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_snapshotCopyDisabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_snapshotCopyEnabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  snapshot_copy {
    destination_region = data.aws_region.alternate.name
    retention_period   = 1
  }

  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
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
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id, aws_subnet.test3.id]
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.test.name
  publicly_accessible                 = %[2]t
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.test]
}
`, rName, publiclyAccessible))
}

func testAccClusterConfig_iamRoles(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_iam_role" "ec2" {
  name = "%[1]s-ec2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2.arn, aws_iam_role.lambda.arn]
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_updateIAMRoles(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_iam_role" "ec2" {
  name = "%[1]s-ec2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2.arn]
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_updatedUsername(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "new_username"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccClusterConfig_updateAvailabilityZone(rName string, regionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = true
  availability_zone                    = data.aws_availability_zones.available.names[%[2]d]
}
`, rName, regionIndex))
}

func testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName string, regionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = false
  availability_zone                    = data.aws_availability_zones.available.names[%[2]d]
}
`, rName, regionIndex))
}

func testAccClusterConfig_availabilityZoneRelocation(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  number_of_nodes                     = 2
  cluster_type                        = "multi-node"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = %[2]t
}
`, rName, enabled))
}

func testAccClusterConfig_availabilityZoneRelocationPubliclyAccessible(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = true
  availability_zone_relocation_enabled = true
}
`, rName))
}

func testAccClusterConfig_createSnapshot(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier        = %[1]q
  availability_zone         = data.aws_availability_zones.available.names[0]
  database_name             = "mydb"
  master_username           = "foo_test"
  master_password           = "Mustbe8characters"
  node_type                 = "dc2.8xlarge"
  number_of_nodes           = 2
  final_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccClusterConfig_restoreFromSnapshot(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier  = %[1]q
  snapshot_identifier = %[1]q
  availability_zone   = data.aws_availability_zones.available.names[1]
  database_name       = "mydb"
  master_username     = "foo_test"
  master_password     = "Mustbe8characters"
  node_type           = "dc2.large"
  number_of_nodes     = 8
  skip_final_snapshot = true
}
`, rName))
}
