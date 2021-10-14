package redshift_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_redshift_cluster", &resource.Sweeper{
		Name: "aws_redshift_cluster",
		F:    sweepClusters,
	})
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RedshiftConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeClustersPages(&redshift.DescribeClustersInput{}, func(resp *redshift.DescribeClustersOutput, lastPage bool) bool {
		if len(resp.Clusters) == 0 {
			log.Print("[DEBUG] No Redshift clusters to sweep")
			return !lastPage
		}

		for _, c := range resp.Clusters {
			r := tfredshift.ResourceCluster()
			d := r.Data(nil)
			d.Set("skip_final_snapshot", true)
			d.SetId(aws.StringValue(c.ClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Redshift Clusters: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Redshift Clusters for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Redshift Cluster sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSRedshiftCluster_basic(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	config := testAccClusterConfig_basic(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "cluster_nodes.#", "1"),
					resource.TestCheckResourceAttrSet("aws_redshift_cluster.default", "cluster_nodes.0.public_ip_address"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "cluster_type", "single-node"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "publicly_accessible", "true"),
					resource.TestMatchResourceAttr("aws_redshift_cluster.default", "dns_name", regexp.MustCompile(fmt.Sprintf("^tf-redshift-cluster-%d.*\\.redshift\\..*", ri))),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_withFinalSnapshot(t *testing.T) {
	var v redshift.Cluster

	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterWithFinalSnapshotConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_kmsKey(t *testing.T) {
	var v redshift.Cluster

	resourceName := "aws_redshift_cluster.default"
	kmsResourceName := "aws_kms_key.foo"

	ri := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_enhancedVpcRoutingEnabled(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_enhancedVPCRoutingEnabled(ri)
	postConfig := testAccClusterConfig_enhancedVPCRoutingDisabled(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "enhanced_vpc_routing", "true"),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "enhanced_vpc_routing", "false"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_loggingEnabled(t *testing.T) {
	var v redshift.Cluster
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_loggingEnabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.enable", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.bucket_name", fmt.Sprintf("tf-test-redshift-logging-%d", rInt)),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccClusterConfig_loggingDisabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.enable", "false"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_snapshotCopy(t *testing.T) {
	var providers []*schema.Provider
	var v redshift.Cluster
	rInt := sdkacctest.RandInt()

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
				Config: testAccClusterConfig_snapshotCopyEnabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default",
						"snapshot_copy.0.destination_region", "data.aws_region.alternate", "name"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "snapshot_copy.0.retention_period", "1"),
				),
			},

			{
				Config: testAccClusterConfig_snapshotCopyDisabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "snapshot_copy.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_iamRoles(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_iamRoles(ri)
	postConfig := testAccClusterConfig_updateIAMRoles(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "iam_roles.#", "2"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_publiclyAccessible(t *testing.T) {
	var v redshift.Cluster
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_notPubliclyAccessible(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "publicly_accessible", "false"),
				),
			},

			{
				Config: testAccClusterConfig_updatePubliclyAccessible(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "publicly_accessible", "true"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_updateNodeCount(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_basic(ri)
	postConfig := testAccClusterConfig_updateNodeCount(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "1"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "2"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc1.large"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_updateNodeType(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_basic(ri)
	postConfig := testAccClusterConfig_updateNodeType(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc1.large"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "1"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc2.large"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_tags(t *testing.T) {
	var v redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_tags(ri)
	postConfig := testAccClusterConfig_updatedTags(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "tags.%", "3"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "tags.environment", "Production"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "tags.environment", "Production"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_forceNewUsername(t *testing.T) {
	var first, second redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_basic(ri)
	postConfig := testAccClusterConfig_updatedUsername(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &first),
					testAccCheckClusterMasterUsername(&first, "foo_test"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "master_username", "foo_test"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &second),
					testAccCheckClusterMasterUsername(&second, "new_username"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "master_username", "new_username"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeAvailabilityZone(t *testing.T) {
	var first, second redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_basic(ri)
	postConfig := testAccClusterConfig_updatedAvailabilityZone(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &first),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default", "availability_zone", "data.aws_availability_zones.available", "names.0"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &second),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default", "availability_zone", "data.aws_availability_zones.available", "names.1"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeEncryption1(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_basic(ri)
	postConfig := testAccClusterConfig_encrypted(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &cluster1),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "false"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "true"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeEncryption2(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster

	ri := sdkacctest.RandInt()
	preConfig := testAccClusterConfig_encrypted(ri)
	postConfig := testAccClusterConfig_unencrypted(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &cluster1),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "true"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_redshift_cluster.default", &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "false"),
				),
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

func testAccCheckClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

			snapshot_identifier := fmt.Sprintf("tf-acctest-snapshot-%d", rInt)

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, err := conn.DeleteClusterSnapshot(
				&redshift.DeleteClusterSnapshotInput{
					SnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if err != nil {
				return fmt.Errorf("error deleting Redshift Cluster Snapshot (%s): %w", snapshot_identifier, err)
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
		if *c.MasterUsername != value {
			return fmt.Errorf("Expected cluster's MasterUsername: %q, given: %q", value, *c.MasterUsername)
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

func testAccClusterConfig_updateNodeCount(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_updateNodeType(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 1
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_basic(rInt int) string {
	// "InvalidVPCNetworkStateFault: The requested AZ us-west-2a is not a valid AZ."
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_encrypted(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
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

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  encrypted                           = true
  kms_key_id                          = aws_kms_key.foo.arn
}
`, rInt))
}

func testAccClusterConfig_unencrypted(rInt int) string {
	// This is used along with the terraform config created testAccClusterConfig_encrypted, to test removal of encryption.
	//Removing the kms key here causes the key to be deleted before the redshift cluster is unencrypted, resulting in an unstable cluster. This is to be kept for the time-being unti we find a better way to handle this.
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
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

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterWithFinalSnapshotConfig(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = false
  final_snapshot_identifier           = "tf-acctest-snapshot-%[1]d"
}
`, rInt))
}

func testAccClusterConfig_kmsKey(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
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

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  kms_key_id                          = aws_kms_key.foo.arn
  encrypted                           = true
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_enhancedVPCRoutingEnabled(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = true
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_enhancedVPCRoutingDisabled(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_loggingDisabled(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable = false
  }

  skip_final_snapshot = true
}
`, rInt))
}

func testAccClusterConfig_loggingEnabled(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_redshift_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-test-redshift-logging-%[1]d"
  force_destroy = true

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
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-redshift-logging-%[1]d/*"
    },
    {
      "Sid": "Stmt137652664067",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_redshift_service_account.main.arn}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-redshift-logging-%[1]d"
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable      = true
    bucket_name = aws_s3_bucket.bucket.bucket
  }

  skip_final_snapshot = true
}
`, rInt))
}

func testAccClusterConfig_snapshotCopyDisabled(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_snapshotCopyEnabled(rInt int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  snapshot_copy {
    destination_region = data.aws_region.alternate.name
    retention_period   = 1
  }

  skip_final_snapshot = true
}
`, rInt))
}

func testAccClusterConfig_tags(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    environment = "Production"
    cluster     = "reader"
    Type        = "master"
  }
}
`, rInt))
}

func testAccClusterConfig_updatedTags(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    environment = "Production"
  }
}
`, rInt))
}

func testAccClusterConfig_notPubliclyAccessible(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-cluster-not-publicly-accessible"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    foo = "bar"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-foo"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-bar"
  }
}

resource "aws_subnet" "foobar" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-foobar"
  }
}

resource "aws_redshift_subnet_group" "foo" {
  name        = "foo-%[1]d"
  description = "foo description"
  subnet_ids  = [aws_subnet.foo.id, aws_subnet.bar.id, aws_subnet.foobar.id]
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.foo.name
  publicly_accessible                 = false
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.foo]
}
`, rInt))
}

func testAccClusterConfig_updatePubliclyAccessible(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-cluster-upd-publicly-accessible"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    foo = "bar"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-foo"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-bar"
  }
}

resource "aws_subnet" "foobar" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-foobar"
  }
}

resource "aws_redshift_subnet_group" "foo" {
  name        = "foo-%[1]d"
  description = "foo description"
  subnet_ids  = [aws_subnet.foo.id, aws_subnet.bar.id, aws_subnet.foobar.id]
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.foo.name
  publicly_accessible                 = true
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.foo]
}
`, rInt))
}

func testAccClusterConfig_iamRoles(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_iam_role" "ec2-role" {
  name = "test-role-ec2-%[1]d"
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

resource "aws_iam_role" "lambda-role" {
  name = "test-role-lambda-%[1]d"
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

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2-role.arn, aws_iam_role.lambda-role.arn]
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_updateIAMRoles(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_iam_role" "ec2-role" {
  name = "test-role-ec2-%[1]d"
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

resource "aws_iam_role" "lambda-role" {
  name = "test-role-lambda-%[1]d"
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

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2-role.arn]
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_updatedUsername(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "new_username"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccClusterConfig_updatedAvailabilityZone(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%[1]d"
  availability_zone                   = data.aws_availability_zones.available.names[1]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}
