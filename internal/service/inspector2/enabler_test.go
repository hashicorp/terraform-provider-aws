package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "2"),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceTypes),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(originalResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, originalResourceTypes),
					testAccCheckEnablerID(resourceName, originalResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(update1ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, update1ResourceTypes),
					testAccCheckEnablerID(resourceName, update1ResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeLambda)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(update2ResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, update2ResourceTypes),
					testAccCheckEnablerID(resourceName, update2ResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(originalResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, originalResourceTypes),
					testAccCheckEnablerID(resourceName, originalResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEc2)),
				),
			},
			{
				Config: testAccEnablerConfig_basic(updatedResourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, updatedResourceTypes),
					testAccCheckEnablerID(resourceName, updatedResourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExists(ctx, resourceTypes),
					testAccCheckEnablerID(resourceName, resourceTypes),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers),
		CheckDestroy:             testAccCheckEnablerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_MemberAccount(resourceTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnablerExistsProvider(ctx, resourceTypes, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers)),
					testAccCheckEnablerIDProvider(resourceName, resourceTypes, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers)),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "account_ids.*", "data.aws_caller_identity.member", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", string(types.ResourceScanTypeEcr)),
				),
			},
		},
	})
}

func testAccCheckEnablerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := ""
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_enabler" {
				continue
			}

			id = rs.Primary.ID
			break
		}

		if id == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id, errors.New("not in state"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		accountIDs, _, err := tfinspector2.ParseEnablerID(id)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id, err)
		}

		st, err := tfinspector2.AccountStatuses(ctx, conn, accountIDs)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id, err)
		}

		for k, v := range st {
			if v.Status != types.StatusDisabled {
				err = multierror.Append(err,
					create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id,
						fmt.Errorf("after destroy, expected DISABLED for account %s, got: %s", k, v),
					),
				)
			}
		}
		return err
	}
}

func testAccCheckEnablerExists(ctx context.Context, t []types.ResourceScanType) resource.TestCheckFunc {
	return testAccCheckEnablerExistsProvider(ctx, t, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckEnablerExistsProvider(ctx context.Context, t []types.ResourceScanType, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This is using `acctest.Provider`, as the resource is created in the primary account,
		// using the account ID of the secondary account
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		accountID := acctest.ProviderAccountID(providerF())
		accountIDs := []string{accountID}
		id := tfinspector2.EnablerID(accountIDs, t)
		st, err := tfinspector2.AccountStatuses(ctx, conn, accountIDs)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id, err)
		}

		for k, s := range st {
			if s.Status != types.StatusEnabled {
				err = multierror.Append(err, create.Error(
					names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id,
					fmt.Errorf("after create, expected ENABLED for account %s, got: %s", k, s.Status)),
				)
			}
		}
		return err
	}
}

func testAccCheckEnablerID(resourceName string, types []types.ResourceScanType) resource.TestCheckFunc {
	return testAccCheckEnablerIDProvider(resourceName, types, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckEnablerIDProvider(resourceName string, types []types.ResourceScanType, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		accountID := acctest.ProviderAccountID(providerF())
		id := tfinspector2.EnablerID([]string{accountID}, types)
		return resource.TestCheckResourceAttr(resourceName, "id", id)(s)
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

func testAccEnablerConfig_MultiAccount(types []types.ResourceScanType) string {
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

  depends_on = [aws_inspector2_member_association.test]
}

resource "aws_inspector2_member_association" "test" {
  account_id = data.aws_caller_identity.member.account_id

  depends_on = [aws_inspector2_delegated_admin_account.test]
}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`, strings.Join(enum.Slice(types...), `", "`)),
	)
}
