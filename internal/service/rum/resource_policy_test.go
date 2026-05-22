// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rum"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchrum "github.com/hashicorp/terraform-provider-aws/internal/service/rum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRUMResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var output rum.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_resource_policy.test"

	policyDoc1 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowRUM","Effect":"Allow","Principal":{"AWS":"*"},"Action":"rum:PutRumMetricsDestination","Resource":"*"}]}`
	policyDoc2 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowRUMUpdated","Effect":"Allow","Principal":{"AWS":"*"},"Action":"rum:PutRumMetricsDestination","Resource":"*"}]}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName, policyDoc1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_revision_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourcePolicyConfig_basic(rName, policyDoc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_revision_id"),
				),
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RUMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rum_resource_policy" {
				continue
			}

			_, err := tfcloudwatchrum.FindResourcePolicy(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch RUM Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *rum.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RUMClient(ctx)

		output, err := tfcloudwatchrum.FindResourcePolicy(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName, policyDoc string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"
}

resource "aws_rum_resource_policy" "test" {
  app_monitor_name = aws_rum_app_monitor.test.name
  policy_document  = %[2]q
}
`, rName, policyDoc)
}
