// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	resourceName := "aws_identitystore_group.test"
	displayName := "Acceptance Test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(displayName),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectGlobalARNNoAccountIDFormat(resourceName, tfjsonpath.New(names.AttrARN), "identitystore", "group/{group_id}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(displayName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("group_id"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("identity_store_id"), "data.aws_ssoadmin_instances.test", tfjsonpath.New("identity_store_ids").AtSliceIndex(0), compare.ValuesSame()),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					acctest.CheckSDKResourceDisappears(ctx, t, tfidentitystore.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIdentityStoreGroup_descriptionChange(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	description1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_description(description1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(description1)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
				),
			},
			{
				Config: testAccGroupConfig_description(description2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(description2)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
				),
			},
		},
	})
}

func TestAccIdentityStoreGroup_displayNameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var group identitystore.DescribeGroupOutput
	displayName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	displayName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_displayName(displayName1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(displayName1)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
				),
			},
			{
				Config: testAccGroupConfig_displayName(displayName2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(displayName2)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IdentityStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_identitystore_group" {
				continue
			}

			_, err := tfidentitystore.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["group_id"])

			if retry.NotFound(err) {
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

func testAccCheckGroupExists(ctx context.Context, t *testing.T, n string, v *identitystore.DescribeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IdentityStoreClient(ctx)

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
resource "aws_identitystore_group" "test" {
  identity_store_id = data.aws_ssoadmin_instances.test.identity_store_ids[0]
  display_name      = %[1]q
}

data "aws_ssoadmin_instances" "test" {}
`, displayName)
}

func testAccGroupConfig_description(description string) string {
	return fmt.Sprintf(`
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "Test display name"
  description       = %[1]q
}

data "aws_ssoadmin_instances" "test" {}
`, description)
}

func testAccGroupConfig_displayName(displayName string) string {
	return fmt.Sprintf(`
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Test description"
}

data "aws_ssoadmin_instances" "test" {}
`, displayName)
}
