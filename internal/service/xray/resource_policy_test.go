// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRayResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy types.ResourcePolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_check", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_revision_id", "1"),
				),
			},
		},
	})
}

func TestAccXRayResourcePolicy_policyDocument(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy types.ResourcePolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"
	policyDocument1 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccess","Effect":"Allow","Principal":{"AWS":"*"},"Action":["xray:*","xray:PutResourcePolicy"],"Resource":"*"}]}`
	policyDocument2 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccessUpdated","Effect":"Allow","Principal":{"AWS":"*"},"Action":["xray:PutTraceSegments","xray:PutTelemetryRecords"],"Resource":"*"}]}`
	policyDocument3 := `{"Version":"2012-10-17","Statement":[{"Sid":"AllowXRayAccessFinal","Effect":"Allow","Principal":{"Service":"sns.amazonaws.com"},"Action":"xray:PutTraceSegments","Resource":"*"}]}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_revision_id", "1"),
					testAccCheckResourcePolicyHasPolicyDocument(ctx, resourceName, policyDocument1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "policy_name",
				ImportStateIdFunc:                    testAccResourcePolicyImportStateIDFunc(resourceName),
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_check",
				},
			},
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_revision_id", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					testAccCheckResourcePolicyHasPolicyDocument(ctx, resourceName, policyDocument2),
				),
			},
			{
				Config: testAccResourcePolicyConfig_policyDocument(rName, policyDocument3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_revision_id", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					testAccCheckResourcePolicyHasPolicyDocument(ctx, resourceName, policyDocument3),
				),
			},
		},
	})
}

func TestAccXRayResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var resourcepolicy types.ResourcePolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_xray_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfxray.ResourceResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_xray_resource_policy" {
				continue
			}

			_, err := tfxray.FindResourcePolicyByName(ctx, conn, rs.Primary.Attributes["policy_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.XRay, create.ErrActionCheckingDestroyed, tfxray.ResNameResourcePolicy, rs.Primary.Attributes["policy_name"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, name string, resourcepolicy *types.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.XRay, create.ErrActionCheckingExistence, tfxray.ResNameResourcePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.XRay, create.ErrActionCheckingExistence, tfxray.ResNameResourcePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)

		output, err := tfxray.FindResourcePolicyByName(ctx, conn, rs.Primary.Attributes["policy_name"])

		if err != nil {
			return err
		}

		*resourcepolicy = *output

		return nil
	}
}

func testAccResourcePolicyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return rs.Primary.Attributes["policy_name"], nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_xray_resource_policy" "test" {
  policy_name                 = %[1]q
  policy_document             = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"AllowXRayAccess\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"xray:*\",\"xray:PutResourcePolicy\"],\"Resource\":\"*\"}]}"
  bypass_policy_lockout_check = true
}
`, rName)
}

func testAccResourcePolicyConfig_policyDocument(rName, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_xray_resource_policy" "test" {
  policy_name                 = %[1]q
  policy_document             = %[2]q
  bypass_policy_lockout_check = true
}
`, rName, policyDocument)
}

func testAccCheckResourcePolicyHasPolicyDocument(ctx context.Context, name string, expectedDocument string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.XRay, create.ErrActionCheckingExistence, tfxray.ResNameResourcePolicy, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)
		output, err := tfxray.FindResourcePolicyByName(ctx, conn, rs.Primary.Attributes["policy_name"])
		if err != nil {
			return err
		}

		if output.PolicyDocument == nil {
			return fmt.Errorf("policy_document is nil")
		}

		if aws.ToString(output.PolicyDocument) != expectedDocument {
			return fmt.Errorf("policy_document mismatch:\nexpected: %s\nactual:   %s", expectedDocument, aws.ToString(output.PolicyDocument))
		}

		return nil
	}
}
