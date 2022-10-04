package opsworks_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOpsWorksJavaAppLayer_basic(t *testing.T) {
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_java_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJavaAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJavaAppLayerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "app_server", "tomcat"),
					resource.TestCheckResourceAttr(resourceName, "app_server_version", "7"),
					resource.TestCheckResourceAttr(resourceName, "jvm_options", ""),
					resource.TestCheckResourceAttr(resourceName, "jvm_type", "openjdk"),
					resource.TestCheckResourceAttr(resourceName, "jvm_version", "7"),
					resource.TestCheckResourceAttr(resourceName, "name", "Java App Server"),
				),
			},
		},
	})
}

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func testAccCheckJavaAppLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_java_app_layer", s)
}

func testAccJavaAppLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), `
resource "aws_opsworks_java_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = aws_security_group.test[*].id
}
`)
}
