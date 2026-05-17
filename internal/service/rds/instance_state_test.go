// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstanceState_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_instance_state.test"
	state := "available"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(rName, "available"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, state),
				),
			},
		},
	})
}

func TestAccRDSInstanceState_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_instance_state.test"
	stateAvailable := "available"
	stateStopped := "stopped"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(rName, "available"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, stateAvailable),
				),
			},
			{
				Config: testAccInstanceStateConfig_basic(rName, "stopped"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, stateStopped),
				),
			},
		},
	})
}

func testAccCheckInstanceStateExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameInstanceState, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrIdentifier] == "" {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameInstanceState, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)
		out, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

		if err != nil {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameInstanceState, rs.Primary.Attributes[names.AttrIdentifier], err)
		}

		if out == nil {
			return fmt.Errorf("Instance State %q does not exist", rs.Primary.Attributes[names.AttrIdentifier])
		}

		return nil
	}
}

func TestAccRDSInstanceState_disappears_Instance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_instance_state.test"
	parentResourceName := "aws_db_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig_basic(rName, "available"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStateExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceInstance(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccInstanceStateConfig_basic(rName, state string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_rds_instance_state" "test" {
  identifier = aws_db_instance.test.identifier
  state      = %[1]q
}
`, state))
}
