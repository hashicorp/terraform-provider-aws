package ecs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
)

func TestAccECSTag_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
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

func TestAccECSTag_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11951
func TestAccECSTag_ResourceARN_batchComputeEnvironment(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckBatch(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_resourceARNBatchComputeEnvironment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
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

func TestAccECSTag_value(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig_basic(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig_basic(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccTagConfig_basic(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_ecs_tag" "test" {
  resource_arn = aws_ecs_cluster.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}

func testAccTagConfig_resourceARNBatchComputeEnvironment(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "batch.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q
  service_role             = aws_iam_role.test.arn
  type                     = "UNMANAGED"

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_ecs_tag" "test" {
  resource_arn = aws_batch_compute_environment.test.ecs_cluster_arn
  key          = "testkey"
  value        = "testvalue"
}
`, rName)
}

func testAccPreCheckBatch(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn

	input := &batch.DescribeComputeEnvironmentsInput{}

	_, err := conn.DescribeComputeEnvironments(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
