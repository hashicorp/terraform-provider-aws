// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderLifecyclePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecyclePolicy types.LifecyclePolicy
	resourceName := "aws_imagebuilder_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilder),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName, &lifecyclePolicy),
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

func testAccCheckLifecyclePolicyExists(ctx context.Context, name string, lifecyclePolicy *types.LifecyclePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderClient(ctx)
		resp, err := tfimagebuilder.FindLifecyclePolicyByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, rs.Primary.ID, err)
		}

		*lifecyclePolicy = *resp

		return nil
	}
}

func testAccCheckLifecyclePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_lifecycle_policy" {
				continue
			}

			_, err := tfimagebuilder.FindLifecyclePolicyByARN(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ImageBuilder, create.ErrActionCheckingDestroyed, tfimagebuilder.ResNameLifecyclePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.ImageBuilder, create.ErrActionCheckingDestroyed, tfimagebuilder.ResNameLifecyclePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccLifecyclePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "imagebuilder.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
  name = %[1]q
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/EC2ImageBuilderLifecycleExecutionPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
    }
    filter {
      type  = "AGE"
      value = 6
      unit  = "YEARS"
    }
  }
  resource_selection {
    tag_map = {
      "key" = "value"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}
