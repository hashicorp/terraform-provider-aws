package docdbelastic_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdbelastic/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdocdbelastic "github.com/hashicorp/terraform-provider-aws/internal/service/docdbelastic"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDocDBElasticCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdbelastic_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.DocDBElasticEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBElasticEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "shard_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "shard_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_user_name", "testuser"),
					resource.TestCheckResourceAttr(resourceName, "admin_user_password", "testpassword"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "tue:04:00-tue:04:30"),
					resource.TestCheckResourceAttr(resourceName, "client_token", fmt.Sprintf("%s-token", rName)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "cluster_arn", "docdb-elastic", regexp.MustCompile(`cluster/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"admin_user_password",
				},
			},
		},
	})
}

func TestAccDocDBElasticCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdbelastic_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.DocDBElasticEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBElasticEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdocdbelastic.ResourceCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBElasticClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdbelastic_cluster" {
				continue
			}

			_, err := tfdocdbelastic.FindClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.DocDBElastic, create.ErrActionCheckingDestroyed, tfdocdbelastic.ResNameCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, name string, cluster *awstypes.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DocDBElastic, create.ErrActionCheckingExistence, tfdocdbelastic.ResNameCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DocDBElastic, create.ErrActionCheckingExistence, tfdocdbelastic.ResNameCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBElasticClient(ctx)
		resp, err := tfdocdbelastic.FindClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.DocDBElastic, create.ErrActionCheckingExistence, tfdocdbelastic.ResNameCluster, rs.Primary.ID, err)
		}

		*cluster = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBElasticClient(ctx)

	input := &docdbelastic.ListClustersInput{}
	_, err := conn.ListClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.10.${count.index}.0/24"
  vpc_id            = aws_vpc.main.id
}

resource "aws_security_group" "basic" {
  name   = "%[1]s-security-group"
  vpc_id = aws_vpc.main.id
}

resource "aws_docdbelastic_cluster" "test" {
  name   = %[1]q
  shard_capacity = 2
  shard_count    = 1

  admin_user_name     = "testuser"
  admin_user_password = "testpassword"
  auth_type           = "PLAIN_TEXT"

  preferred_maintenance_window = "tue:04:00-tue:04:30"

  vpc_security_group_ids = [
    aws_security_group.basic.id
  ]

  subnet_ids = [
    aws_subnet.test[0].id,
    aws_subnet.test[1].id
  ]


  tags = {
    Name = %[1]q
  }
}
`, rName)
}
