package elasticache_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElastiCacheGlobalReplicationGroup_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(primaryReplicationGroupResourceName, &primaryReplicationGroup),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`globalreplicationgroup:`+tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryReplicationGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexp.MustCompile(tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", tfelasticache.EmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
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

func TestAccElastiCacheGlobalReplicationGroup_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandString(10)
	description2 := sdkacctest.RandString(10)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description2),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceGlobalReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_multipleSecondaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_MultipleSecondaries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_ReplaceSecondary_differentRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
			{
				Config: testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_clusterMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_ClusterMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
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

func testAccCheckGlobalReplicationGroupExists(resourceName string, v *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Global Replication Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		grg, err := tfelasticache.FindGlobalReplicationGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving ElastiCache Global Replication Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleting || aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleted {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) exists, but is in a non-available state: %s", rs.Primary.ID, aws.StringValue(grg.Status))
		}

		*v = *grg

		return nil
	}
}

func testAccCheckGlobalReplicationGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_global_replication_group" {
			continue
		}

		_, err := tfelasticache.FindGlobalReplicationGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Global Replication Group (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheckGlobalReplicationGroup(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	_, err := conn.DescribeGlobalReplicationGroups(input)

	if acctest.PreCheckSkipError(err) ||
		tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId)
}

func testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  global_replication_group_description = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, description)
}

func testAccGlobalReplicationGroupConfig_MultipleSecondaries(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "alternate", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "alternate" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.alternate.name

  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.secondary.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccGlobalReplicationGroupConfig_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test"

  engine         = "redis"
  engine_version = "6.2"
  node_type      = "cache.m5.large"

  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  cluster_mode {
    num_node_groups         = 2
    replicas_per_node_group = 1
  }
}
`, rName)
}

func testAccVPCBaseWithProvider(rName, name, provider string, subnetCount int) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsNoOptInConfigWithProvider(name, provider),
		fmt.Sprintf(`
resource "aws_vpc" "%[1]s" {
  provider = %[2]s

  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "%[1]s" {
  provider = %[2]s

  count = %[4]d

  vpc_id            = aws_vpc.%[1]s.id
  cidr_block        = "192.168.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.%[1]s.names[count.index]
}

resource "aws_elasticache_subnet_group" "%[1]s" {
  provider = %[2]s

  name       = %[3]q
  subnet_ids = aws_subnet.%[1]s[*].id
}
`, name, provider, rName, subnetCount),
	)
}

func testAccAvailableAZsNoOptInConfigWithProvider(name, provider string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "%[1]s" {
  provider = %[2]s

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, name, provider)
}
