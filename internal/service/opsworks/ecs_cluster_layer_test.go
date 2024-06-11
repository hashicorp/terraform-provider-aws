// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksECSClusterLayer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_ecs_cluster_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckECSClusterLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccECSClusterLayerConfig_basic(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_cluster_arn", "aws_ecs_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "Ecs Cluster"),
				),
			},
		},
	})
}

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func testAccCheckECSClusterLayerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckLayerDestroy(ctx, "aws_opsworks_ecs_cluster_layer", s)
	}
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
