package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_global_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_global_replication_group",
		F:    testSweepElasticacheGlobalReplicationGroups,
		Dependencies: []string{
			"aws_elasticache_replication_group",
		},
	})
}

func testSweepElasticacheGlobalReplicationGroups(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).elasticacheconn
	input := &elasticache.DescribeGlobalReplicationGroupsInput{}

	err = conn.DescribeGlobalReplicationGroupsPages(input, func(out *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		for _, globalReplicationGroup := range out.GlobalReplicationGroups {
			id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)
			input := &elasticache.DeleteGlobalReplicationGroupInput{
				GlobalReplicationGroupId: globalReplicationGroup.GlobalReplicationGroupId,
			}

			log.Printf("[INFO] Deleting Elasticache Global Replication Group: %s", id)

			_, err := conn.DeleteGlobalReplicationGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete ElastiCache Global Replication Group (%s): %s", id, err)
				continue
			}

			if err := waitForElasticacheGlobalReplicationGroupDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Failure while waiting for ElastiCache Global Replication Group (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Global Replication Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving ElastiCache Global Replication Groups: %s", err)
	}

	return nil
}

func TestAccAWSElasticacheGlobalReplicationGroup_basic(t *testing.T) {
	var globalReplcationGroup1 elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticacheGlobalReplicationGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplcationGroup1),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`globalreplicationgroup:\w{5}-`+rName)), // \w{5} is the AWS prefix
					resource.TestCheckResourceAttr(resourceName, "at_rest_encryption_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", "cache.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", "0"),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "retain_primary_replication_group", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", "false"),
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

func TestAccAWSElasticacheGlobalReplicationGroup_disappears(t *testing.T) {
	var globalReplcationGroup1 elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticacheGlobalReplicationGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplcationGroup1),
					testAccCheckAWSElasticacheGlobalReplicationGroupDisappears(&globalReplcationGroup1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName string, globalReplicationGroup *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elasticache Global Replication Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		cluster, err := elasticacheDescribeGlobalReplicationGroup(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("Elasticache Global Replication Group not found")
		}

		if aws.StringValue(cluster.Status) != "available" && aws.StringValue(cluster.Status) != "primary-only" {
			return fmt.Errorf("Elasticache Global Replication Group (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(cluster.Status))
		}

		*globalReplicationGroup = *cluster

		return nil
	}
}

func testAccCheckAWSElasticacheGlobalReplicationGroupDisappears(globalReplicationGroup *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		input := &elasticache.DeleteGlobalReplicationGroupInput{
			GlobalReplicationGroupId:      globalReplicationGroup.GlobalReplicationGroupId,
			RetainPrimaryReplicationGroup: aws.Bool(true),
		}

		_, err := conn.DeleteGlobalReplicationGroup(input)

		if err != nil {
			return err
		}

		return waitForElasticacheGlobalReplicationGroupDeletion(conn, aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId))
	}
}

func testAccCheckAWSElasticacheGlobalReplicationGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_global_replication_group" {
			continue
		}

		globalReplicationGroup, err := elasticacheDescribeGlobalReplicationGroup(conn, rs.Primary.ID)

		if isAWSErr(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if globalReplicationGroup == nil {
			continue
		}

		return fmt.Errorf("Elasticache Global Replication Group (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(globalReplicationGroup.Status))
	}

	return nil
}

func testAccPreCheckAWSElasticacheGlobalReplicationGroup(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}

	_, err := conn.DescribeGlobalReplicationGroups(input)

	if testAccPreCheckSkipError(err) || isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSElasticacheGlobalReplicationGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, rName)
}
