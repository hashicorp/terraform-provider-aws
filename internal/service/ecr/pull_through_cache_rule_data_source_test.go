package ecr_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRPullThroughCacheRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	prefix := "ecr-public"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_basic(prefix),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccPullThroughCacheRuleDataSourceConfig_basic(prefix string) string {
	return fmt.Sprintf(`
data "aws_ecr_pull_through_cache_rule" "default" {
  ecr_repository_prefix = %q
}
`, prefix)
}
