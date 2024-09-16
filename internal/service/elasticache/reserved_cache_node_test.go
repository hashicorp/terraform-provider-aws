package elasticache_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheReservedCacheNode_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("TF_TEST_ELASTICACHE_RESERVED_CACHE_NODE") == "" {
		t.Skip("Environment variable TF_TEST_ELASTICACHE_RESERVED_CACHE_NODE is not set")
	}

	var reservation awstypes.ReservedCacheNode
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_reserved_cache_node.test"
	dataSourceName := "data.aws_elasticache_reserved_cache_node_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccReservedInstanceExists(ctx, resourceName, &reservation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticache", regexache.MustCompile(`reserved-instance:.+`)),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_node_type", resourceName, "cache_node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDuration, resourceName, names.AttrDuration),
					resource.TestCheckResourceAttrPair(dataSourceName, "fixed_price", resourceName, "fixed_price"),
					resource.TestCheckResourceAttr(resourceName, "cache_node_count", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "offering_id", resourceName, "offering_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "offering_type", resourceName, "offering_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "product_description", resourceName, "product_description"),
					resource.TestCheckResourceAttrSet(resourceName, "recurring_charges"),
					resource.TestCheckResourceAttr(resourceName, "reservation_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStartTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(resourceName, "usage_price"),
				),
			},
		},
	})
}

func testAccReservedInstanceExists(ctx context.Context, n string, reservation *awstypes.ReservedCacheNode) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Reserved Cache Node reservation id is set")
		}

		resp, err := tfelasticache.FindReservedCacheNodeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*reservation = resp

		return nil
	}
}

func testAccReservedInstanceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_reserved_cache_node" "test" {
  offering_id      = data.aws_elasticache_reserved_cache_node_offering.test.offering_id
  reservation_id   = %[1]q
}

data "aws_elasticache_reserved_cache_node_offering" "test" {
  cache_node_type     = "cache.t4g.small"
  duration            = 31536000
  offering_type       = "No Upfront"
  product_description = "redis"
}
`, rName)
}
