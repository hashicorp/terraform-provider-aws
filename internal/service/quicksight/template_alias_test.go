// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightTemplateAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var templateAlias awstypes.TemplateAlias
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	aliasName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_template_alias.test"
	resourceTemplateName := "aws_quicksight_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAliasConfig_basic(rId, rName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAliasExists(ctx, resourceName, &templateAlias),
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
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	aliasName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_template_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAliasConfig_basic(rId, rName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAliasExists(ctx, resourceName, &templateAlias),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceTemplateAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTemplateAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_template_alias" {
				continue
			}

			_, err := tfquicksight.FindTemplateAliasByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["template_id"], rs.Primary.Attributes["alias_name"])

			if tfresource.NotFound(err) {
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

func testAccCheckTemplateAliasExists(ctx context.Context, n string, v *awstypes.TemplateAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

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
