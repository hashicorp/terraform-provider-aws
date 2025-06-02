// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationscontacts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notificationscontacts/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnotificationscontacts "github.com/hashicorp/terraform-provider-aws/internal/service/notificationscontacts"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsContactsEmailContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var emailcontact awstypes.EmailContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notificationscontacts_email_contact.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsContactsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &emailcontact),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("notifications-contacts", regexache.MustCompile(`emailcontact/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("email_address"), knownvalue.StringExact(rEmailAddress)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNotificationsContactsEmailContact_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var emailcontact awstypes.EmailContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notificationscontacts_email_contact.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsContactsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &emailcontact),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotificationscontacts.ResourceEmailContact, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNotificationsContactsEmailContact_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.EmailContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEmailAddress := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_notificationscontacts_email_contact.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsContactsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailContactConfig_tags1(rName, rEmailAddress, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccEmailContactConfig_tags2(rName, rEmailAddress, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccEmailContactConfig_tags1(rName, rEmailAddress, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckEmailContactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notificationscontacts_email_contact" {
				continue
			}

			_, err := tfnotificationscontacts.FindEmailContactByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("User Notifications Contacts Email Contact %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckEmailContactExists(ctx context.Context, n string, v *awstypes.EmailContact) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsContactsClient(ctx)

		output, err := tfnotificationscontacts.FindEmailContactByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsContactsClient(ctx)

	var input notificationscontacts.ListEmailContactsInput

	_, err := conn.ListEmailContacts(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEmailContactConfig_basic(rName, rEmail string) string {
	return fmt.Sprintf(`
resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q
}
`, rName, rEmail)
}

func testAccEmailContactConfig_tags1(rName, rEmail, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rEmail, tag1Key, tag1Value)
}

func testAccEmailContactConfig_tags2(rName, rEmail, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, rEmail, tag1Key, tag1Value, tag2Key, tag2Value)
}
