// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	resourceName := "aws_identitystore_group.test"
	displayName := "Acceptance Test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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

func TestAccIdentityStoreGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfidentitystore.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIdentityStoreGroup_descriptionChange(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	description1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_description(description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
				),
			},
			{
				Config: testAccGroupConfig_description(description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_identitystore_group" {
				continue
			}

			_, err := tfidentitystore.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["group_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IdentityStore Group %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckGroupExists(ctx context.Context, n string, v *identitystore.DescribeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, n, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, n, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient(ctx)

		output, err := tfidentitystore.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["group_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGroupConfig_basic(displayName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Example description"
}
`, displayName)
}

func testAccGroupConfig_description(description string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "Test display name"
  description       = %[1]q
}
`, description)
}
