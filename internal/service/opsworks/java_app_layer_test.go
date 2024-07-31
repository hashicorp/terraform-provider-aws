// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksJavaAppLayer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_java_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJavaAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJavaAppLayerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "app_server", "tomcat"),
					resource.TestCheckResourceAttr(resourceName, "app_server_version", "7"),
					resource.TestCheckResourceAttr(resourceName, "jvm_options", ""),
					resource.TestCheckResourceAttr(resourceName, "jvm_type", "openjdk"),
					resource.TestCheckResourceAttr(resourceName, "jvm_version", "7"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "Java App Server"),
				),
			},
		},
	})
}

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func testAccCheckJavaAppLayerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error { return testAccCheckLayerDestroy(ctx, "aws_opsworks_java_app_layer", s) }
}

func testAccJavaAppLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), `
resource "aws_opsworks_java_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = aws_security_group.test[*].id
}
`)
}
