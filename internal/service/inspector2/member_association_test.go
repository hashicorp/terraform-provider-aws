package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	// "github.com/aws/aws-sdk-go-v2/service/inspector2"
	// "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	// "github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2MemberAssociation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":      testAccMemberAssociation_basic,
		"disappears": testAccMemberAssociation_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccMemberAssociation_basic(t *testing.T) {
	resourceName := "aws_inspector2_member_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", "data.aws_caller_identity", "account_id"),
				),
			},
		},
	})
}

func testAccMemberAssociation_disappears(t *testing.T) {
	resourceName := "aws_inspector2_member_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfinspector2.ResourceMemberAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMemberAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector2_member_association" {
			continue
		}

		_, st, err := tfinspector2.FindAssociatedMemberStatus(ctx, conn, rs.Primary.ID)

		// if st == "" && errs.Contains(err, "account not found") {
		// 	return nil
		// }

		if st == "REMOVED" {
			return nil
		}

		if err != nil {
			return err
		}

		return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameMemberAssociation, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckMemberAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameMemberAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameMemberAssociation, name, errors.New("not set"))
		}

		id := rs.Primary.ID

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		ai, st, err := tfinspector2.FindAssociatedMemberStatus(context.Background(), conn, id)

		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameMemberAssociation, id, err)
		}

		if st != "ENABLED" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameEnabler, id, fmt.Errorf("after create, expected ENABLED for account %s, got: %s", ai, st))
		}

		return nil

	}

}

func testAccMemberAssociationConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_inspector2_member_association" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`
}
