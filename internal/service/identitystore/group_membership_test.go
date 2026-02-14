// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroupMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var groupMembership identitystore.DescribeGroupMembershipOutput
	userResourceName := "aws_identitystore_user.test"
	groupResourceName := "aws_identitystore_group.test"
	resourceName := "aws_identitystore_group_membership.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // Group Name
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // User Name

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					resource.TestCheckResourceAttrPair(resourceName, "member_id", userResourceName, "user_id"),
					resource.TestCheckResourceAttrPair(resourceName, "group_id", groupResourceName, "group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_id"),
					resource.TestCheckResourceAttrSet(resourceName, "membership_id"),
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

func TestAccIdentityStoreGroupMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var groupMembership identitystore.DescribeGroupMembershipOutput
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // Group Name
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // User Name
	resourceName := "aws_identitystore_group_membership.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					acctest.CheckSDKResourceDisappears(ctx, t, tfidentitystore.ResourceGroupMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIdentityStoreGroupMembership_GroupId(t *testing.T) {
	ctx := acctest.Context(t)
	var groupMembership identitystore.DescribeGroupMembershipOutput
	groupResourceName := "aws_identitystore_group.test"
	resourceName := "aws_identitystore_group_membership.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // Group Name 1
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // Group Name 2
	rName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // User Name

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(rName1, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					resource.TestCheckResourceAttrPair(resourceName, "group_id", groupResourceName, "group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupMembershipConfig_basic(rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					resource.TestCheckResourceAttrPair(resourceName, "group_id", groupResourceName, "group_id"),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroupMembership_MemberId(t *testing.T) {
	ctx := acctest.Context(t)
	var groupMembership identitystore.DescribeGroupMembershipOutput
	groupResourceName := "aws_identitystore_user.test"
	resourceName := "aws_identitystore_group_membership.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // Group Name
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // User Name 1
	rName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) // User Name 2

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					resource.TestCheckResourceAttrPair(resourceName, "member_id", groupResourceName, "user_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupMembershipConfig_basic(rName1, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName, &groupMembership),
					resource.TestCheckResourceAttrPair(resourceName, "member_id", groupResourceName, "user_id"),
				),
			},
		},
	})
}

func testAccCheckGroupMembershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IdentityStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_identitystore_group_membership" {
				continue
			}

			_, err := tfidentitystore.FindGroupMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["membership_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IdentityStore Group Membership %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupMembershipExists(ctx context.Context, t *testing.T, n string, v *identitystore.DescribeGroupMembershipOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IdentityStoreClient(ctx)

		output, err := tfidentitystore.FindGroupMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["membership_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGroupMembershipConfig_basic(groupName, userName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[2]q
  description       = "Acceptance Test"
}

resource "aws_identitystore_group_membership" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  group_id  = aws_identitystore_group.test.group_id
  member_id = aws_identitystore_user.test.user_id
}
`, userName, groupName)
}
