package opsworks_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOpsWorksMemcachedLayer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_memcached_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemcachedLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemcachedLayerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "allocated_memory", "512"),
					resource.TestCheckResourceAttr(resourceName, "name", "Memcached"),
				),
			},
		},
	})
}

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func testAccCheckMemcachedLayerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckLayerDestroy(ctx, "aws_opsworks_memcached_layer", s)
	}
}

func testAccMemcachedLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), `
resource "aws_opsworks_memcached_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = aws_security_group.test[*].id
}
`)
}
