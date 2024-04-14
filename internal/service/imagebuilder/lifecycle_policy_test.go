// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
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
	resourceName := "aws_imagebuilder_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("lifecycle-policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Used for setting lifecycle policies"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.type", string(awstypes.LifecyclePolicyDetailActionTypeDelete)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.type", string(awstypes.LifecyclePolicyDetailFilterTypeAge)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.value", "6"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.retain_at_least", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.unit", string(awstypes.LifecyclePolicyTimeUnitYears)),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tag_map.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tag_map.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.tag_map.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", string(awstypes.LifecyclePolicyResourceTypeAmiImage)),
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

func TestAccImageBuilderLifecyclePolicy_policyDetails(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_imagebuilder_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_policyDetails(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.type", string(awstypes.LifecyclePolicyDetailActionTypeDisable)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.include_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.include_resources.0.amis", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.is_public", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.0.unit", string(awstypes.LifecyclePolicyTimeUnitDays)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.0.value", "7"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.tag_map.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.tag_map.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.type", string(awstypes.LifecyclePolicyDetailFilterTypeAge)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.value", "6"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.retain_at_least", "5"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.unit", string(awstypes.LifecyclePolicyTimeUnitYears)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_policyDetailsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.type", string(awstypes.LifecyclePolicyDetailActionTypeDelete)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.include_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.include_resources.0.amis", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.include_resources.0.snapshots", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.is_public", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.0.unit", string(awstypes.LifecyclePolicyTimeUnitWeeks)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.last_launched.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.tag_map.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.tag_map.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusion_rules.0.amis.0.tag_map.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.type", string(awstypes.LifecyclePolicyDetailFilterTypeCount)),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.filter.0.value", "10"),
				),
			},
		},
	})
}

func TestAccImageBuilderLifecyclePolicy_resourceSelection(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_imagebuilder_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_resourceSelection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.0.semantic_version", "1.0.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_resourceSelectionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_selection.0.recipes.0.semantic_version", "2.0.0"),
				),
			},
		},
	})
}

func TestAccImageBuilderLifecyclePolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderLifecyclePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceLifecyclePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLifecyclePolicyExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderClient(ctx)
		_, err := tfimagebuilder.FindLifecyclePolicyByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ImageBuilder, create.ErrActionCheckingExistence, tfimagebuilder.ResNameLifecyclePolicy, rs.Primary.ID, err)
		}

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

func testAccLifecyclePolicyBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

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
`, rName)
}

func testAccLifecyclePolicyBaseConfigComponent(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
  `, rName)
}

func testAccLifecyclePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
    }
    filter {
      type            = "AGE"
      value           = 6
      retain_at_least = 10
      unit            = "YEARS"
    }
  }
  resource_selection {
    tag_map = {
      "key1" = "value1"
      "key2" = "value2"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccLifecyclePolicyConfig_policyDetails(rName string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DISABLE"
      include_resources {
        amis = true
      }
    }
    exclusion_rules {
      amis {
        is_public = false
        last_launched {
          unit  = "DAYS"
          value = 7
        }
        tag_map = {
          "key1" = "value1"
        }
      }
    }
    filter {
      type            = "AGE"
      value           = "6"
      retain_at_least = "5"
      unit            = "YEARS"
    }
  }
  resource_selection {
    tag_map = {
      "key1" = "value1"
      "key2" = "value2"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccLifecyclePolicyConfig_policyDetailsUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
      include_resources {
        amis      = true
        snapshots = true
      }
    }
    exclusion_rules {
      amis {
        is_public = true
        regions   = ["${data.aws_region.current.name}"]
        last_launched {
          unit  = "WEEKS"
          value = 2
        }
        tag_map = {
          "key1" = "value1"
          "key2" = "value2"
        }
      }
    }
    filter {
      type  = "COUNT"
      value = "10"
    }
  }
  resource_selection {
    tag_map = {
      "key1" = "value1"
      "key2" = "value2"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccLifecyclePolicyConfig_resourceSelection(rName string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		testAccLifecyclePolicyBaseConfigComponent(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
    }
    filter {
      type            = "AGE"
      value           = 6
      retain_at_least = 10
      unit            = "YEARS"
    }
  }
  resource_selection {
    recipes {
      name             = aws_imagebuilder_image_recipe.test.name
      semantic_version = aws_imagebuilder_image_recipe.test.version
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccLifecyclePolicyConfig_resourceSelectionUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		testAccLifecyclePolicyBaseConfigComponent(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "2.0.0"
}

resource "aws_imagebuilder_lifecycle_policy" "test" {
  name           = %[1]q
  description    = "Used for setting lifecycle policies"
  execution_role = aws_iam_role.test.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
    }
    filter {
      type            = "AGE"
      value           = 6
      retain_at_least = 10
      unit            = "YEARS"
    }
  }
  resource_selection {
    recipes {
      name             = aws_imagebuilder_image_recipe.test.name
      semantic_version = aws_imagebuilder_image_recipe.test.version
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccLifecyclePolicyConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccLifecyclePolicyConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
