// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVerifiedAccessGroup_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "deletion_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_instance_id"),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.0.customer_managed_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.0.kms_key_arn", ""),
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

func testAccVerifiedAccessGroup_kms(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_kms(rName, policyDoc),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sse_configuration.0.customer_managed_key_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "sse_configuration.0.kms_key_arn"),
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

func testAccVerifiedAccessGroup_updateKMS(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"
	description := sdkacctest.RandString(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDoc),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "deletion_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_instance_id"),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.0.customer_managed_key_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVerifiedAccessGroupConfig_kms(rName, policyDoc),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sse_configuration.0.customer_managed_key_enabled"),
					resource.TestCheckResourceAttr(resourceName, "sse_configuration.0.customer_managed_key_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "sse_configuration.0.kms_key_arn"),
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

func testAccVerifiedAccessGroup_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVerifiedAccessGroup_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccVerifiedAccessGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVerifiedAccessGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccVerifiedAccessGroup_policy(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandString(100)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccVerifiedAccessGroup_updatePolicy(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandString(100)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"
	policyDocUpdate := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"POST\"\n};"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDocUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDocUpdate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}
func testAccVerifiedAccessGroup_setPolicy(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandString(100)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "deletion_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_instance_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckVerifiedAccessGroupExists(ctx context.Context, n string, v *types.VerifiedAccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVerifiedAccessGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_group" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Group %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccVerifiedAccessGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_trust_provider" "test" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "oidc"

  oidc_options {
    authorization_endpoint = "https://example.com/authorization_endpoint"
    client_id              = "s6BhdRkqt3"
    client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
    issuer                 = "https://example.com"
    scope                  = "test"
    token_endpoint         = "https://example.com/token_endpoint"
    user_info_endpoint     = "https://example.com/user_info_endpoint"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance_trust_provider_attachment" "test" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.test.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}
`, rName)
}

func testAccVerifiedAccessGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), `
resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id
}
`)
}

func testAccVerifiedAccessGroupConfig_kms(rName, policyDoc string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test_key" {
  description = "KMS key for Verified Access Group test"
}

resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id
  policy_document            = %[1]q

  sse_configuration {
    kms_key_arn                  = aws_kms_key.test_key.arn
    customer_managed_key_enabled = true
  }
}
`, policyDoc))
}

func testAccVerifiedAccessGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVerifiedAccessGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVerifiedAccessGroupConfig_policy(rName, description, policy string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id
  description                = %[2]q
  policy_document            = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, description, policy))
}
