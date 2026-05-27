// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ContactList_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccContactList_basic,
		acctest.CtDisappears: testAccContactList_disappears,
		"tags":               testAccSESV2ContactList_tagsSerial,
		"description":        testAccContactList_description,
		"topic":              testAccContactList_topic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccContactList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "topic.#", "0"),
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

func testAccContactList_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, t, resourceName),
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
					testAccCheckContactListExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccContactList_topic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_topic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "topic.#", "1"),
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
					testAccCheckContactListExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "topic.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.default_subscription_status", "OPT_OUT"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.description", names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "topic.0.display_name", "topic2"),
					resource.TestCheckResourceAttr(resourceName, "topic.0.topic_name", "topic2"),
				),
			},
		},
	})
}

func testAccContactList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact_list.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactListConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactListExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceContactList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactListDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_contact_list" {
				continue
			}

			_, err := tfsesv2.FindContactListByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Contact List %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckContactListExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindContactListByID(ctx, conn, rs.Primary.ID)

		return err
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
