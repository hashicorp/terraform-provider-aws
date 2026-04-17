// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightTemplateAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var templateAlias awstypes.TemplateAlias
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	aliasName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_template_alias.test"
	resourceTemplateName := "aws_quicksight_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAliasConfig_basic(rId, rName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAliasExists(ctx, t, resourceName, &templateAlias),
					resource.TestCheckResourceAttr(resourceName, "alias_name", aliasName),
					resource.TestCheckResourceAttrPair(resourceName, "template_id", resourceTemplateName, "template_id"),
					resource.TestCheckResourceAttrPair(resourceName, "template_version_number", resourceTemplateName, "version_number"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("template/%[1]s/alias/%[2]s", rId, aliasName)),
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

func TestAccQuickSightTemplateAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var templateAlias awstypes.TemplateAlias
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	aliasName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_template_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAliasConfig_basic(rId, rName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAliasExists(ctx, t, resourceName, &templateAlias),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceTemplateAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTemplateAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_template_alias" {
				continue
			}

			_, err := tfquicksight.FindTemplateAliasByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["template_id"], rs.Primary.Attributes["alias_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Template Alias (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTemplateAliasExists(ctx context.Context, t *testing.T, n string, v *awstypes.TemplateAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindTemplateAliasByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["template_id"], rs.Primary.Attributes["alias_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTemplateAliasConfig_basic(rId, rName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_basic(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_template_alias" "test" {
  alias_name              = %[1]q
  template_id             = aws_quicksight_template.test.template_id
  template_version_number = aws_quicksight_template.test.version_number
}
`, aliasName))
}
