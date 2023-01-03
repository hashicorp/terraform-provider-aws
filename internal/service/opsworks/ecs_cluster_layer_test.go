package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOpsWorksECSClusterLayer_basic(t *testing.T) {
	var v opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_ecs_cluster_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckECSClusterLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccECSClusterLayerConfig_basic(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_cluster_arn", "aws_ecs_cluster.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", "Ecs Cluster"),
				),
			},
		},
	})
}

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func testAccCheckECSClusterLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_ecs_cluster_layer", s)
}

func testAccECSClusterLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_opsworks_ecs_cluster_layer" "test" {
  stack_id        = aws_opsworks_stack.test.id
  ecs_cluster_arn = aws_ecs_cluster.test.arn

  custom_security_group_ids = aws_security_group.test[*].id
}
`, rName))
}
