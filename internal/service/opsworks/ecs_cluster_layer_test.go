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

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`.

func TestAccOpsWorksECSClusterLayer_basic(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_ecs_cluster_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckECSClusterLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccECSClusterLayerBasic(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_cluster_arn", "aws_ecs_cluster.test", "arn"),
				),
			},
		},
	})
}

func TestAccOpsWorksECSClusterLayer_tags(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_ecs_cluster_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckECSClusterLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccECSClusterLayerTags1Config(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccECSClusterLayerTags2Config(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccECSClusterLayerTags1Config(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckECSClusterLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_ecs_cluster_layer", s)
}

func testAccECSClusterLayerBasic(name string) string {
	return acctest.ConfigCompose(
		testAccStackVPCCreateConfig(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_opsworks_ecs_cluster_layer" "test" {
  stack_id        = aws_opsworks_stack.test.id
  name            = %[1]q
  ecs_cluster_arn = aws_ecs_cluster.test.arn

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
}
`, name))
}

func testAccECSClusterLayerTags1Config(name, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccStackVPCCreateConfig(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_opsworks_ecs_cluster_layer" "test" {
  stack_id        = aws_opsworks_stack.test.id
  name            = %[1]q
  ecs_cluster_arn = aws_ecs_cluster.test.arn

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1))
}

func testAccECSClusterLayerTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccStackVPCCreateConfig(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_opsworks_ecs_cluster_layer" "test" {
  stack_id        = aws_opsworks_stack.test.id
  name            = %[1]q
  ecs_cluster_arn = aws_ecs_cluster.test.arn

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2))
}
