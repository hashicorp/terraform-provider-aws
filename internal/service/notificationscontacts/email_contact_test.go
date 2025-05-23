// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationscontacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notificationscontacts/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &emailcontact),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", rEmailAddress),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications-contacts", regexache.MustCompile(`emailcontact/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccObjectImportStateIdFunc(resourceName),
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
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress, acctest.CtKey1, acctest.CtValue1),
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

func TestAccNotificationsContactsEmailContact_update(t *testing.T) {
	ctx := acctest.Context(t)

	var v1, v2 awstypes.EmailContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccEmailContactConfig_basic(rName, rEmailAddress, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEmailContactExists(ctx, resourceName, &v2),
					testAccCheckEmailContactNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEmailContactConfig_basic(rNameUpdated, rEmailAddress, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "email_address", rEmailAddress),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications-contacts", regexache.MustCompile(`emailcontact/.+$`)),
				),
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
				return nil
			}
			if err != nil {
				return create.Error(names.NotificationsContacts, create.ErrActionCheckingDestroyed, tfnotificationscontacts.ResNameEmailContact, rs.Primary.Attributes[names.AttrARN], err)
			}

			return create.Error(names.NotificationsContacts, create.ErrActionCheckingDestroyed, tfnotificationscontacts.ResNameEmailContact, rs.Primary.Attributes[names.AttrARN], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEmailContactExists(ctx context.Context, name string, emailcontact *awstypes.EmailContact) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NotificationsContacts, create.ErrActionCheckingExistence, tfnotificationscontacts.ResNameEmailContact, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.NotificationsContacts, create.ErrActionCheckingExistence, tfnotificationscontacts.ResNameEmailContact, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsContactsClient(ctx)

		resp, err := tfnotificationscontacts.FindEmailContactByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.NotificationsContacts, create.ErrActionCheckingExistence, tfnotificationscontacts.ResNameEmailContact, rs.Primary.Attributes[names.AttrARN], err)
		}

		*emailcontact = *resp

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

func testAccObjectImportStateIdFunc(resourceName string) func(state *terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccCheckEmailContactNotRecreated(before, after *awstypes.EmailContact) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before != after {
			return create.Error(names.NotificationsContacts, create.ErrActionCheckingNotRecreated, tfnotificationscontacts.ResNameEmailContact, before, errors.New("recreated"))
		}
		return nil
	}
}

func testAccEmailContactConfig_basic(rName, rEmail, rTagKey, rTagValue string) string {
	return fmt.Sprintf(`

resource "aws_notificationscontacts_email_contact" "test" {
  name          = %[1]q
  email_address = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rEmail, rTagKey, rTagValue)
}
