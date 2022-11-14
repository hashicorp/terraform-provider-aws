package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2Enabler_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccEnabler_basic,
		"accountID":  testAccEnabler_accountID,
		"disappears": testAccEnabler_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccEnabler_basic(t *testing.T) {
	resourceName := "aws_inspector2_enabler.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic([]string{"ECR"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnablerExists([]string{"ECR"}),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "ECR"),
				),
			},
		},
	})
}

func testAccEnabler_accountID(t *testing.T) {
	resourceName := "aws_inspector2_enabler.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic([]string{"EC2", "ECR"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnablerExists([]string{"EC2", "ECR"}),
					resource.TestCheckResourceAttr(resourceName, "account_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_ids.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.1", "ECR"),
				),
			},
		},
	})
}

func testAccEnabler_disappears(t *testing.T) {
	resourceName := "aws_inspector2_enabler.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnablerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnablerConfig_basic([]string{"ECR"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnablerExists([]string{"ECR"}),
					acctest.CheckResourceDisappears(acctest.Provider, tfinspector2.ResourceEnabler(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnablerDestroy(s *terraform.State) error {
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

	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client

	st, err := tfinspector2.FindAccountStatuses(context.Background(), conn, id)
	if err != nil {
		return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id, err)
	}

	for _, s := range st {
		if s.Status != string(types.StatusDisabled) {
			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameEnabler, id, fmt.Errorf("after destroy, expected DISABLED for account %s, got: %s", s.AccountID, s.Status))
		}
	}

	return nil
}

func testAccCheckEnablerExists(t []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client

		id := tfinspector2.EnablerID([]string{acctest.Provider.Meta().(*conns.AWSClient).AccountID}, t)
		st, err := tfinspector2.FindAccountStatuses(context.Background(), conn, id)
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id, err)
		}

		for _, s := range st {
			if s.Status != string(types.StatusEnabled) {
				return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id, fmt.Errorf("after create, expected ENABLED for account %s, got: %s", s.AccountID, s.Status))
			}
		}
		return nil
	}
}

func testAccEnablerConfig_basic(types []string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_inspector2_enabler" "test" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["%[1]s"]
}
`, strings.Join(types, `", "`))
}
