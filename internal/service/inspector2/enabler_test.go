// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnabler_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_accountID(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2, types.ResourceScanTypeEcr}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceEnabler(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccEnabler_updateResourceTypes(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	originalResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2}
	update1ResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2, types.ResourceScanTypeLambda}
	update2ResourceTypes := []types.ResourceScanType{types.ResourceScanTypeLambda}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(originalResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, originalResourceTypes),
					testAccCheckEnablerID(resourceName, originalResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(update1ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, update1ResourceTypes),
					testAccCheckEnablerID(resourceName, update1ResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(update2ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, update2ResourceTypes),
					testAccCheckEnablerID(resourceName, update2ResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
			},
		},
	})
}

func testAccEnabler_updateResourceTypes_disjoint(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	originalResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2}
	updatedResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(originalResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, originalResourceTypes),
					testAccCheckEnablerID(resourceName, originalResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(updatedResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, updatedResourceTypes),
					testAccCheckEnablerID(resourceName, updatedResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_lambda(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeLambda}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
			},
		},
	})
}

func testAccEnabler_memberAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.member"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	providers := make(map[string]*schema.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers)),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.member", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_memberAccount_disappearsMemberAssociation(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.member"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	providers := make(map[string]*schema.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceMemberAssociation(), "aws_inspector2_member_association.member"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccEnabler_memberAccount_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.members"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	providers := make(map[string]*schema.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckThirdAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers, acctest.ProviderName, acctest.ProviderNameAlternate, acctest.ProviderNameThird),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount_Multiple(t, resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers),
						acctest.NamedProviderFunc(acctest.ProviderNameThird, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.third", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_memberAccount_updateMemberAccounts(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.members"
	resourceTypes := []types.ResourceScanType{types.ResourceScanTypeEcr}

	providers := make(map[string]*schema.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckThirdAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers, acctest.ProviderName, acctest.ProviderNameAlternate, acctest.ProviderNameThird),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsAlternate(t, resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsMultiple(t, resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers),
						acctest.NamedProviderFunc(acctest.ProviderNameThird, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.third", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsThird(t, resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, resourceTypes),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameThird, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.third", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccEnabler_memberAccount_updateMemberAccountsAndScanTypes(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.members"
	originalResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2}
	update1ResourceTypes := []types.ResourceScanType{types.ResourceScanTypeEc2, types.ResourceScanTypeLambda}
	update2ResourceTypes := []types.ResourceScanType{types.ResourceScanTypeLambda}

	providers := make(map[string]*schema.Provider)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckThirdAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers, acctest.ProviderName, acctest.ProviderNameAlternate, acctest.ProviderNameThird),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsAlternate(t, originalResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, originalResourceTypes),
					testAccCheckEnablerIDProvider(resourceName, originalResourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
				),
			},
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsMultiple(t, update1ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, update1ResourceTypes),
					testAccCheckEnablerIDProvider(resourceName, update1ResourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers),
						acctest.NamedProviderFunc(acctest.ProviderNameThird, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.third", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
			},
			{
				Config: testAccEnablerConfig_MemberAccount_UpdateMemberAccountsThird(t, update2ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceName, update2ResourceTypes),
					testAccCheckEnablerIDProvider(resourceName, update2ResourceTypes,
						acctest.NamedProviderFunc(acctest.ProviderNameThird, providers),
					),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.third", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
				// PlanOnly: true,
			},
		},
	})
}

func testAccCheckEnablerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_enabler" {
				continue
			}

			accountIDs, _, err := tfinspector2.ParseEnablerID(rs.Primary.ID)
			if err != nil {
				return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, rs.Primary.ID, err)
			}

			st, err := tfinspector2.AccountStatuses(ctx, conn, accountIDs)
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, rs.Primary.ID, err)
			}

			var errs []error
			for k, v := range st {
				if v.Status != types.StatusDisabled {
					errs = append(errs,
						create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, rs.Primary.ID,
							fmt.Errorf("after destroy, expected DISABLED for account %s, got: %s", k, v),
						),
					)
				}
			}

			if err := errors.Join(errs...); err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckEnablerExists(ctx context.Context, name string, t []types.ResourceScanType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		accountIDs, _, err := tfinspector2.ParseEnablerID(rs.Primary.ID)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, name, err)
		}

		id := tfinspector2.EnablerID(accountIDs, t)
		st, err := tfinspector2.AccountStatuses(ctx, conn, accountIDs)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, name, err)
		}

		var errs []error
		for k, s := range st {
			if s.Status != types.StatusEnabled {
				errs = append(errs, create.Error(
					names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id,
					fmt.Errorf("after create, expected ENABLED for account %s, got: %s", k, s.Status)),
				)
			}
		}

		return errors.Join(errs...)
	}
}

func testAccCheckEnablerID(resourceName string, types []types.ResourceScanType) resource.TestCheckFunc {
	return testAccCheckEnablerIDProvider(resourceName, types, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckEnablerIDProvider(resourceName string, types []types.ResourceScanType, providerF ...func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// accountID := acctest.ProviderAccountID(providerF())
		accountIDs := tfslices.ApplyToAll(providerF, func(f func() *schema.Provider) string {
			return acctest.ProviderAccountID(f())
		})
		id := tfinspector2.EnablerID(accountIDs, types)
		return resource.TestCheckResourceAttr(resourceName, names.AttrID, id)(s)
	}
}

func testAccEnablerConfig_basic(types []types.ResourceScanType) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_inspector2_enabler" "test" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["%[1]s"]
}
`, strings.Join(enum.Slice(types...), `", "`))
}

func testAccEnablerConfig_MemberAccount(types []types.ResourceScanType) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_inspector2_enabler" "member" {
  account_ids    = [data.aws_caller_identity.member.account_id]
  resource_types = ["%[1]s"]

  depends_on = [aws_inspector2_member_association.member]
}

resource "aws_inspector2_member_association" "member" {
  account_id = data.aws_caller_identity.member.account_id

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}

func testAccEnablerConfig_MemberAccount_Multiple(t *testing.T, types []types.ResourceScanType) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleAccountProvider(t, 3),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_caller_identity" "third" {
  provider = "awsthird"
}

locals {
  member_account_ids = [
    data.aws_caller_identity.alternate.account_id,
    data.aws_caller_identity.third.account_id,
  ]
}

resource "aws_inspector2_enabler" "members" {
  account_ids    = local.member_account_ids
  resource_types = ["%[1]s"]

  depends_on = [aws_inspector2_member_association.members]
}

resource "aws_inspector2_member_association" "members" {
  count = length(local.member_account_ids)

  account_id = local.member_account_ids[count.index]

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}

func testAccEnablerConfig_MemberAccount_UpdateMemberAccountsAlternate(t *testing.T, types []types.ResourceScanType) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleAccountProvider(t, 3),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

locals {
  member_account_ids = [
    data.aws_caller_identity.alternate.account_id,
  ]
}

resource "aws_inspector2_enabler" "members" {
  account_ids    = local.member_account_ids
  resource_types = ["%[1]s"]

  depends_on = [aws_inspector2_member_association.members]
}

resource "aws_inspector2_member_association" "members" {
  count = length(local.member_account_ids)

  account_id = local.member_account_ids[count.index]

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}

func testAccEnablerConfig_MemberAccount_UpdateMemberAccountsMultiple(t *testing.T, types []types.ResourceScanType) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleAccountProvider(t, 3),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_caller_identity" "third" {
  provider = "awsthird"
}

locals {
  member_account_ids = [
    data.aws_caller_identity.alternate.account_id,
    data.aws_caller_identity.third.account_id,
  ]
}

resource "aws_inspector2_enabler" "members" {
  account_ids    = local.member_account_ids
  resource_types = ["%[1]s"]

  depends_on = [aws_inspector2_member_association.members]
}

resource "aws_inspector2_member_association" "members" {
  count = length(local.member_account_ids)

  account_id = local.member_account_ids[count.index]

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}

func testAccEnablerConfig_MemberAccount_UpdateMemberAccountsThird(t *testing.T, types []types.ResourceScanType) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleAccountProvider(t, 3),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_caller_identity" "third" {
  provider = "awsthird"
}

locals {
  member_account_ids = [
    data.aws_caller_identity.third.account_id,
  ]
}

resource "aws_inspector2_enabler" "members" {
  account_ids    = local.member_account_ids
  resource_types = ["%[1]s"]

  depends_on = [aws_inspector2_member_association.members]
}

resource "aws_inspector2_member_association" "members" {
  count = length(local.member_account_ids)

  account_id = local.member_account_ids[count.index]

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}
