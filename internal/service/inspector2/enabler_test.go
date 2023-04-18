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
	resourceTypes := []types.ResourceScanType{"ECR"}

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
					resource.TestCheckTypeSetElemAttr(resourceName, "resource_types.*", "ECR"),
				),
			},
		},
	})
}

func testAccEnabler_accountID(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{"EC2", "ECR"}

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
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.1", "ECR"),
				),
			},
		},
	})
}

func testAccEnabler_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_inspector2_enabler.test"
	resourceTypes := []types.ResourceScanType{"ECR"}

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
	originalResourceTypes := []types.ResourceScanType{"EC2"}
	updatedResourceTypes := []types.ResourceScanType{"ECR"}

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
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "EC2"),
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
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "ECR"),
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

		st, err := tfinspector2.AccountStatuses(ctx, conn, id)
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
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		id := tfinspector2.EnablerID([]string{acctest.Provider.Meta().(*conns.AWSClient).AccountID}, t)
		st, err := tfinspector2.AccountStatuses(ctx, conn, id)
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
	return func(s *terraform.State) error {
		accountID := acctest.AccountID()
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
