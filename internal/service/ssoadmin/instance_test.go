// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminInstance_serial(t *testing.T) {
	t.Parallel()
	testCases := map[string]func(*testing.T){
		acctest.CtBasic:      testAccSSOAdminInstance_basic,
		acctest.CtDisappears: testAccSSOAdminInstance_disappears,
		"ListBasic":          testAccSSOAdminInstance_listBasic,
	}
	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSSOAdminInstance_listBasic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			testAccInstancePreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.SSOAdminServiceID), ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{{Config: testAccInstanceConfig_list(rName), Check: resource.ComposeAggregateTestCheckFunc(
			testAccCheckInstanceExists(ctx, t, resourceName), resource.TestCheckResourceAttr("data.aws_ssoadmin_instances.test", "arns.#", "1"),
			resource.TestCheckResourceAttrPair("data.aws_ssoadmin_instances.test", "arns.0", resourceName, names.AttrARN),
		)}},
	})
}

func testAccSSOAdminInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			testAccInstancePreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.SSOAdminServiceID), ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{Config: testAccInstanceConfig_basic(rName), Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckInstanceExists(ctx, t, resourceName), resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.key_type", "AWS_OWNED_KMS_KEY"),
				resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "1"), resource.TestCheckResourceAttr(resourceName, names.AttrTags+".Name", rName),
			)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: names.AttrARN, ImportStateVerifyIgnore: []string{"client_token"}},
			{Config: testAccInstanceConfig_updated(rName), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-updated"), resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "2"),
			)},
		},
	})
}

func testAccSSOAdminInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			testAccInstancePreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.SSOAdminServiceID), ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: testAccCheckInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{{
			Config: testAccInstanceConfig_basic(rName),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckInstanceExists(ctx, t, resourceName),
				acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceInstance, resourceName),
			),
			ExpectNonEmptyPlan: true,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
				},
			},
		}},
	})
}

func testAccInstancePreCheck(ctx context.Context, t *testing.T) {
	output, err := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx).ListInstances(ctx, &ssoadmin.ListInstancesInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
	if len(output.Instances) > 0 {
		t.Skip("skipping acceptance testing: account already has an IAM Identity Center instance")
	}
}

func testAccCheckInstanceExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameInstance, name, errors.New("not found"))
		}
		_, err := tfssoadmin.FindInstanceByARN(ctx, acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx), rs.Primary.Attributes[names.AttrARN])
		return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameInstance, rs.Primary.Attributes[names.AttrARN], err)
	}
}

func testAccCheckInstanceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_instance" {
				continue
			}
			_, err := tfssoadmin.FindInstanceByARN(ctx, acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx), rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return errors.New("SSO Admin Instance still exists")
		}
		return nil
	}
}

func testAccInstanceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name = %[1]q

  encryption_configuration {
    key_type = "AWS_OWNED_KMS_KEY"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccInstanceConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name = %[2]q

  encryption_configuration {
    key_type = "AWS_OWNED_KMS_KEY"
  }

  tags = {
    Name        = %[1]q
    Environment = "test"
  }
}
`, rName, rName+"-updated")
}

func testAccInstanceConfig_list(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name = %[1]q
}

data "aws_ssoadmin_instances" "test" {
  depends_on = [aws_ssoadmin_instance.test]
}
`, rName)
}
