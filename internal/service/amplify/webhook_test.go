// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWebhook_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/.+/webhooks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrURL, regexache.MustCompile(fmt.Sprintf(`^https://webhooks.amplify.%s.%s/.+$`, acctest.Region(), acctest.PartitionDNSSuffix()))),
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

func testAccWebhook_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamplify.ResourceWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccWebhook_update(t *testing.T) {
	ctx := acctest.Context(t)
	var webhook types.Webhook
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig_description(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_name", fmt.Sprintf("%s-1", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testdescription1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebhookConfig_descriptionUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(ctx, resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testdescription2"),
				),
			},
		},
	})
}

func testAccCheckWebhookExists(ctx context.Context, resourceName string, v *types.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Webhook ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		webhook, err := tfamplify.FindWebhookByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *webhook

		return nil
	}
}

func testAccCheckWebhookDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_amplify_webhook" {
				continue
			}

			_, err := tfamplify.FindWebhookByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amplify Webhook %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccWebhookConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test.branch_name
}
`, rName)
}

func testAccWebhookConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test1" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-1"
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test1.branch_name
  description = "testdescription1"
}
`, rName)
}

func testAccWebhookConfig_descriptionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test1" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-1"
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test2.branch_name
  description = "testdescription2"
}
`, rName)
}
