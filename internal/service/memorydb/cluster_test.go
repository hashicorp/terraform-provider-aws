package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMemoryDBCluster_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, memorydb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "acl_name", "aws_memorydb_acl.test", "id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "number_of_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_group_name", "aws_memorydb_subnet_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", "true"),
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

func TestAccMemoryDBCluster_defaults(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, memorydb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl_name", "open-access"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "number_of_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_group_name", "default"), // created automatically & matches the default vpc
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", "true"),
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

func TestAccMemoryDBCluster_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, memorydb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBCluster_update_maintenanceWindow(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, memorydb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withMaintenanceWindow(rName, "thu:09:00-thu:10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "thu:09:00-thu:10:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withMaintenanceWindow(rName, "fri:09:00-fri:10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "fri:09:00-fri:10:00"),
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

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_cluster" {
			continue
		}

		_, err := tfmemorydb.FindClusterByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MemoryDB Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClusterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindClusterByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccClusterConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id
}
`, rName),
	)
}

func testAccClusterConfigUserAndACL(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }
}

resource "aws_memorydb_acl" "test" {
  name       = %[1]q
  user_names = [aws_memorydb_user.test.id]
}
`, rName)
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBase(rName),
		testAccClusterConfigUserAndACL(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name           = aws_memorydb_acl.test.id
  name               = %[1]q
  node_type          = "db.t4g.small"
  subnet_group_name  = aws_memorydb_subnet_group.test.id

  tags = {
    Test = "test"
  }
}
`, rName),
	)
}

func testAccClusterConfig_defaults(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name  = "open-access"
  name      = %[1]q
  node_type = "db.t4g.small"
}
`, rName),
	)
}

func testAccClusterConfig_withMaintenanceWindow(rName, maintenanceWindow string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name           = "open-access"
  maintenance_window = %[2]q
  name               = %[1]q
  node_type          = "db.t4g.small"
  subnet_group_name  = aws_memorydb_subnet_group.test.id
}
`, rName, maintenanceWindow),
	)
}
