package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudhsmv2/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func init() {
	resource.AddTestSweepers("aws_cloudhsm_v2_cluster", &resource.Sweeper{
		Name:         "aws_cloudhsm_v2_cluster",
		F:            testSweepCloudhsmv2Clusters,
		Dependencies: []string{"aws_cloudhsm_v2_hsm"},
	})
}

func testSweepCloudhsmv2Clusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudHSMV2Conn
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(cluster.ClusterId))
			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudHSMv2 Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudHSMv2 Clusters (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudHSMv2 Clusters (%s): %w", region, err)
	}

	return nil
}

func testAccAWSCloudHsmV2Cluster_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "cluster_id", regexp.MustCompile(`^cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_state", cloudhsmv2.ClusterStateUninitialized),
					resource.TestCheckResourceAttr(resourceName, "hsm_type", "hsm1.medium"),
					resource.TestMatchResourceAttr(resourceName, "security_group_id", regexp.MustCompile(`^sg-.+`)),
					resource.TestCheckResourceAttr(resourceName, "source_backup_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
		},
	})
}

func testAccAWSCloudHsmV2Cluster_disappears(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceCluster(), resourceName),
					// Verify Delete error handling
					acctest.CheckResourceDisappears(acctest.Provider, ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSCloudHsmV2Cluster_Tags(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags1("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1updated", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key3", "value3"),
				),
			},
		},
	})
}

func testAccAWSCloudHsmV2ClusterConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}
`
}

func testAccAWSCloudHsmV2ClusterConfig() string {
	return acctest.ConfigCompose(testAccAWSCloudHsmV2ClusterConfigBase(), `
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id
}
`)
}

func testAccAWSCloudHsmV2ClusterConfigTags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAWSCloudHsmV2ClusterConfigBase(), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAWSCloudHsmV2ClusterConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAWSCloudHsmV2ClusterConfigBase(), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCheckAWSCloudHsmV2ClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_cluster" {
			continue
		}
		cluster, err := finder.Cluster(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster != nil && aws.StringValue(cluster.State) != cloudhsmv2.ClusterStateDeleted {
			return fmt.Errorf("CloudHSM cluster still exists %s", cluster)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsmV2ClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudHSMV2Conn
		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := finder.Cluster(conn, it.Primary.ID)

		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
