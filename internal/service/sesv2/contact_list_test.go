// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ContactList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	// Only one contact list is allowed per AWS account.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct0),
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

func TestAccSESV2ContactList_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	// Only one contact list is allowed per AWS account.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactListConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccSESV2ContactList_topic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	// Only one contact list is allowed per AWS account.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_topic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topic.0.default_subscription_status", "OPT_IN"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "topic.0.display_name", "topic1"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.topic_name", "topic1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactListConfig_topic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "topic.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "topic.0.default_subscription_status", "OPT_OUT"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.description", names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "topic.0.display_name", "topic2"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.topic_name", "topic2"),
				),
			},
		},
	})
}

func TestAccSESV2ContactList_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	// Only one contact list is allowed per AWS account.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactListConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSESV2ContactList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	// Only one contact list is allowed per AWS account.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceContactList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactListDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_contact_list" {
				continue
			}

			_, err := tfsesv2.FindContactListByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameContactList, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContactListExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContactList, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContactList, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindContactListByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContactList, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContactListConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q
}
`, rName)
}

func testAccContactListConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q
  description       = %[2]q
}
`, rName, description)
}

func testAccContactListConfig_topic1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q

  topic {
    default_subscription_status = "OPT_IN"
    display_name                = "topic1"
    topic_name                  = "topic1"
  }
}
`, rName)
}

func testAccContactListConfig_topic2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q

  topic {
    default_subscription_status = "OPT_OUT"
    description                 = "description"
    display_name                = "topic2"
    topic_name                  = "topic2"
  }
}
`, rName)
}

func testAccContactListConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccContactListConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
