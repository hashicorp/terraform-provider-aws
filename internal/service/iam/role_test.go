// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRole_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttrSet(resourceName, "create_date"),
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

func TestAccIAMRole_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "description", "This 1s a D3scr!pti0n with weird content: &@90ë\"'{«¡Çø}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "description", "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë\"'{«¡Çø}"),
				),
			},
			{
				Config: testAccRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "create_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccIAMRole_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccIAMRole_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_namePrefix(acctest.ResourcePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", acctest.ResourcePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", acctest.ResourcePrefix),
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

func TestAccIAMRole_testNameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_pre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inline_policy"},
			},
			{
				Config: testAccRoleConfig_post(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/23288
// https://github.com/hashicorp/terraform-provider-aws/issues/28833
func TestAccIAMRole_diffs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/28835
func TestAccIAMRole_diffsCondition(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_diffsCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffsCondition(rName),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffsCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffsCondition(rName),
				PlanOnly: true,
			},
			{
				Config: testAccRoleConfig_diffsCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
				),
			},
			{
				Config:   testAccRoleConfig_diffsCondition(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccIAMRole_badJSON(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccRoleConfig_badJSON(rName),
				ExpectError: regexache.MustCompile(`.*contains an invalid JSON policy:.*`),
			},
		},
	})
}

func TestAccIAMRole_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceRole(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRole_policiesForceDetach(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_forceDetachPolicies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					testAccAddRolePolicy(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_detach_policies", "inline_policy", "managed_policy_arns"},
			},
		},
	})
}

func TestAccIAMRole_maxSessionDuration(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccRoleConfig_maxSessionDuration(rName, 3599),
				ExpectError: regexache.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config:      testAccRoleConfig_maxSessionDuration(rName, 43201),
				ExpectError: regexache.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config: testAccRoleConfig_maxSessionDuration(rName, 3700),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3700"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleConfig_maxSessionDuration(rName, 3701),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3701"),
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

func TestAccIAMRole_permissionsBoundary(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", acctest.Partition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccRoleConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test update
			{
				Config: testAccRoleConfig_permissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckRolePermissionsBoundary(&role, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
			// Test removal
			{
				Config: testAccRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckRolePermissionsBoundary(&role, ""),
				),
			},
			// Test addition
			{
				Config: testAccRoleConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test drift detection
			{
				PreConfig: func() {
					// delete the boundary manually
					conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)
					input := &iam.DeleteRolePermissionsBoundaryInput{
						RoleName: role.RoleName,
					}
					_, err := conn.DeleteRolePermissionsBoundaryWithContext(ctx, input)
					if err != nil {
						t.Fatalf("Failed to delete permission_boundary from role (%s): %s", aws.StringValue(role.RoleName), err)
					}
				},
				Config: testAccRoleConfig_permissionsBoundary(rName, permissionsBoundary1),
				// check the boundary was restored
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccRoleConfig_permissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckRolePermissionsBoundary(&role, ""),
				),
			},
		},
	})
}

func TestAccIAMRole_InlinePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyInline(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				Config: testAccRoleConfig_policyInlineUpdate(rName, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				Config: testAccRoleConfig_policyInlineUpdateDown(rName, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inline_policy.0.policy"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19444
func TestAccIAMRole_InlinePolicy_ignoreOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyInlineActionOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				Config:   testAccRoleConfig_policyInlineActionOrder(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccRoleConfig_policyInlineActionNewOrder(rName),
				PlanOnly: true,
			},
			{
				Config:             testAccRoleConfig_policyInlineActionOrderActualDiff(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRole_InlinePolicy_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyEmptyInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
				),
			},
		},
	})
}

func TestAccIAMRole_ManagedPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyManaged(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
			{
				Config: testAccRoleConfig_policyManagedUpdate(rName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
				),
			},
			{
				Config: testAccRoleConfig_policyManagedUpdateDown(rName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
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

// TestAccIAMRole_PolicyOutOfBandRemovalAddedBack_managedNonEmpty: if a policy is detached
// out of band, it should be reattached.
func TestAccIAMRole_ManagedPolicy_outOfBandRemovalAddedBack(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyDetachManagedPolicy(ctx, &role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandRemovalAddedBack_inlineNonEmpty: if a policy is removed
// out of band, it should be recreated.
func TestAccIAMRole_InlinePolicy_outOfBandRemovalAddedBack(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyInline(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyRemoveInlinePolicy(ctx, &role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyInline(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
				),
			},
		},
	})
}

// TestAccIAMRole_ManagedPolicy_outOfBandAdditionRemoved: if managed_policy_arns arg
// exists and is non-empty, policy attached out of band should be removed
func TestAccIAMRole_ManagedPolicy_outOfBandAdditionRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyExtraManaged(rName, policyName1, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAttachManagedPolicy(ctx, &role, policyName2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyExtraManaged(rName, policyName1, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandAdditionRemoved_inlineNonEmpty: if inline_policy arg
// exists and is non-empty, policy added out of band should be removed
func TestAccIAMRole_InlinePolicy_outOfBandAdditionRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyInline(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAddInlinePolicy(ctx, &role, policyName2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyInline(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandAdditionIgnored_inlineNonExistent: if there is no
// inline_policy attribute, out of band changes should be ignored.
func TestAccIAMRole_InlinePolicy_outOfBandAdditionIgnored(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyNoInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAddInlinePolicy(ctx, &role, policyName1),
				),
			},
			{
				Config: testAccRoleConfig_policyNoInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAddInlinePolicy(ctx, &role, policyName2),
				),
			},
			{
				Config: testAccRoleConfig_policyNoInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyRemoveInlinePolicy(ctx, &role, policyName1),
					testAccCheckRolePolicyRemoveInlinePolicy(ctx, &role, policyName2),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandAdditionIgnored_managedNonExistent: if there is no
// managed_policy_arns attribute, out of band changes should be ignored.
func TestAccIAMRole_ManagedPolicy_outOfBandAdditionIgnored(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyNoManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAttachManagedPolicy(ctx, &role, policyName),
				),
			},
			{
				Config: testAccRoleConfig_policyNoManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyDetachManagedPolicy(ctx, &role, policyName),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandAdditionRemoved_inlineEmpty: if inline is added
// out of band with empty inline arg, should be removed
func TestAccIAMRole_InlinePolicy_outOfBandAdditionRemovedEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyEmptyInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAddInlinePolicy(ctx, &role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyEmptyInline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
				),
			},
		},
	})
}

// TestAccIAMRole_PolicyOutOfBandAdditionRemoved_managedEmpty: if managed is attached
// out of band with empty managed arg, should be detached
func TestAccIAMRole_ManagedPolicy_outOfBandAdditionRemovedEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var role iam.Role
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig_policyEmptyManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
					testAccCheckRolePolicyAttachManagedPolicy(ctx, &role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRoleConfig_policyEmptyManaged(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, resourceName, &role),
				),
			},
		},
	})
}

func testAccCheckRoleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_role" {
				continue
			}

			_, err := tfiam.FindRoleByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Role %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRoleExists(ctx context.Context, n string, v *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Role ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		output, err := tfiam.FindRoleByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// Attach inline policy out of band (outside of terraform)
func testAccAddRolePolicy(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		input := &iam.PutRolePolicyInput{
			RoleName: aws.String(rs.Primary.ID),
			PolicyDocument: aws.String(`{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}`),
			PolicyName: aws.String(id.UniqueId()),
		}

		_, err := conn.PutRolePolicyWithContext(ctx, input)
		return err
	}
}

func testAccCheckRolePermissionsBoundary(role *iam.Role, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if role.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *role.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckRolePolicyDetachManagedPolicy(ctx context.Context, role *iam.Role, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		var managedARN string
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: role.RoleName,
		}

		err := conn.ListAttachedRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			for _, v := range page.AttachedPolicies {
				if *v.PolicyName == policyName {
					managedARN = *v.PolicyArn
					break
				}
			}
			return !lastPage
		})
		if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
		}
		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err = conn.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(managedARN),
			RoleName:  role.RoleName,
		})

		return err
	}
}

func testAccCheckRolePolicyAttachManagedPolicy(ctx context.Context, role *iam.Role, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		var managedARN string
		input := &iam.ListPoliciesInput{
			PathPrefix:        aws.String("/tf-testing/"),
			PolicyUsageFilter: aws.String("PermissionsPolicy"),
			Scope:             aws.String("Local"),
		}

		err := conn.ListPoliciesPagesWithContext(ctx, input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
			for _, v := range page.Policies {
				if *v.PolicyName == policyName {
					managedARN = *v.Arn
					break
				}
			}
			return !lastPage
		})
		if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
		}
		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err = conn.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
			PolicyArn: aws.String(managedARN),
			RoleName:  role.RoleName,
		})

		return err
	}
}

func testAccCheckRolePolicyAddInlinePolicy(ctx context.Context, role *iam.Role, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		_, err := conn.PutRolePolicyWithContext(ctx, &iam.PutRolePolicyInput{
			PolicyDocument: aws.String(testAccRolePolicyExtraInlineConfig()),
			PolicyName:     aws.String(inlinePolicy),
			RoleName:       role.RoleName,
		})

		return err
	}
}

func testAccCheckRolePolicyRemoveInlinePolicy(ctx context.Context, role *iam.Role, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		_, err := conn.DeleteRolePolicyWithContext(ctx, &iam.DeleteRolePolicyInput{
			PolicyName: aws.String(inlinePolicy),
			RoleName:   role.RoleName,
		})

		return err
	}
}

func testAccRoleConfig_maxSessionDuration(rName string, maxSessionDuration int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name                 = %[1]q
  path                 = "/"
  max_session_duration = %[2]d

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName, maxSessionDuration)
}

func testAccRoleConfig_permissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  name                 = %[1]q
  path                 = "/"
  permissions_boundary = %[2]q
}
`, rName, permissionsBoundary)
}

func testAccRoleConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_diffs(rName, tags string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_user" "user1" {
  name = "%[1]s-baa204a2"
  path = "/"
}

resource "aws_iam_user" "user2" {
  name = "%[1]s-fee06121"
  path = "/"
}

resource "aws_iam_user" "user3" {
  name = "%[1]s-2f0d132b"
  path = "/"
}

resource "aws_iam_user" "user4" {
  name = "%[1]s-d1eaee06"
  path = "/"
}

resource "aws_iam_user" "user5" {
  name = "%[1]s-d4a38c26"
  path = "/"
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Id      = %[1]q
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        AWS = [
          aws_iam_user.user1.arn,
          aws_iam_user.user2.arn,
          aws_iam_user.user3.arn,
          aws_iam_user.user4.arn,
          aws_iam_user.user5.arn,
        ]
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  %[2]s
}
`, rName, tags)
}

func testAccRoleConfig_diffsCondition(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_user" "user1" {
  name = "%[1]s-cde2c453"
  path = "/"
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRoleWithSAML"
      Condition = {
        IpAddress = {
          "aws:SourceIp" = [
            "0.0.0.0/0",
          ]
        }
        StringEquals = {
          "SAML:aud" = "https://signin.aws.amazon.com/saml"
        }
      }
      Principal = {
        AWS = [
          aws_iam_user.user1.arn,
        ]
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_description(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name        = %[1]q
  description = "This 1s a D3scr!pti0n with weird content: &@90ë\"'{«¡Çø}"
  path        = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name        = %[1]q
  description = "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë\"'{«¡Çø}"
  path        = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_nameGenerated() string {
	return `
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`
}

func testAccRoleConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name_prefix = %[1]q
  path        = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_pre(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "%[1]s-2"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "%[1]s-2"
  path = "/test/"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccRoleConfig_post(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "%[1]s-2"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "%[1]s-2"
  path = "/test/"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccRoleConfig_badJSON(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<INTENTIONALLYBAD
{
  "Version": "2012-10-17",
  "Statement": [
  {
    "Action": "sts:AssumeRole",
    "Principal": {
    "Service": "ec2.${data.aws_partition.current.dns_suffix}",
    },
    "Effect": "Allow",
    "Sid": ""
  }
  ]
}
INTENTIONALLYBAD
}
`, rName)
}

func testAccRoleConfig_forceDetachPolicies(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_role" "test" {
  name                  = %[1]q
  force_detach_policies = true

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_tags0(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, rName)
}

func testAccRoleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRoleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRoleConfig_tagsNull(rName, tagKey1 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  tags = {
    %[2]q = null
  }
}
`, rName, tagKey1)
}

func testAccRoleConfig_policyInline(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }
}
`, roleName, policyName)
}

func testAccRoleConfig_policyInlineUpdate(roleName, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }

  inline_policy {
    name = %[3]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }
}
`, roleName, policyName2, policyName3)
}

func testAccRoleConfig_policyInlineUpdateDown(roleName, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ec2:Describe*",
    "Resource": "*",
    "Condition": {
      "DateGreaterThan": {"aws:CurrentTime": "2017-07-01T00:00:00Z"},
      "DateLessThan": {"aws:CurrentTime": "2017-12-31T23:59:59Z"}
    }
  }
}
EOF
  }
}
`, roleName, policyName3)
}

func testAccRoleConfig_policyInlineActionOrder(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[1]q

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "ec2:DescribeScheduledInstances",
          "ec2:DescribeScheduledInstanceAvailability",
          "ec2:DescribeFastSnapshotRestores",
          "ec2:DescribeElasticGpus",
        ]
        Effect   = "Allow"
        Resource = "*"
      }]
    })
  }
}
`, roleName)
}

func testAccRoleConfig_policyInlineActionOrderActualDiff(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[1]q

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "ec2:DescribeScheduledInstances",
          "ec2:DescribeElasticGpus",
          "ec2:DescribeScheduledInstanceAvailability",
        ]
        Effect   = "Allow"
        Resource = "*"
      }]
    })
  }
}
`, roleName)
}

func testAccRoleConfig_policyInlineActionNewOrder(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {
    name = %[1]q

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "ec2:DescribeElasticGpus",
          "ec2:DescribeScheduledInstances",
          "ec2:DescribeFastSnapshotRestores",
          "ec2:DescribeScheduledInstanceAvailability",
        ]
        Effect   = "Allow"
        Resource = "*"
      }]
    })
  }
}
`, roleName)
}

func testAccRoleConfig_policyManaged(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[2]q
  managed_policy_arns = [aws_iam_policy.test.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, policyName, roleName)
}

func testAccRoleConfig_policyManagedUpdate(roleName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test3" {
  name = %[3]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[4]q
  managed_policy_arns = [aws_iam_policy.test2.arn, aws_iam_policy.test3.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, policyName1, policyName2, policyName3, roleName)
}

func testAccRoleConfig_policyManagedUpdateDown(roleName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test3" {
  name = %[3]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[4]q
  managed_policy_arns = [aws_iam_policy.test3.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, policyName1, policyName2, policyName3, roleName)
}

func testAccRoleConfig_policyExtraManaged(roleName, policyName1, policyName2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[3]q
  managed_policy_arns = [aws_iam_policy.test.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, policyName1, policyName2, roleName)
}

func testAccRolePolicyExtraInlineConfig() string {
	return `{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": [
			"ec2:Describe*"
		],
		"Effect": "Allow",
		"Resource": "*"
		}
	]
}`
}

func testAccRoleConfig_policyNoInline(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}
`, roleName)
}

func testAccRoleConfig_policyNoManaged(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_iam_policy" "managed-policy1" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}
`, roleName, policyName)
}

func testAccRoleConfig_policyEmptyInline(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  inline_policy {}
}
`, roleName)
}

func testAccRoleConfig_policyEmptyManaged(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  managed_policy_arns = []
}

resource "aws_iam_policy" "managed-policy1" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}
`, roleName, policyName)
}
