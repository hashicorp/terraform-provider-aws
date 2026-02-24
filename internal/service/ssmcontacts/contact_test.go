// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccContact_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "PERSONAL"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`contact/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We need to explicitly test destroying this resource instead of just using CheckDestroy,
				// because CheckDestroy will run after the replication set has been destroyed and destroying
				// the replication set will destroy all other resources.
				Config: testAccContactConfig_none(),
				Check:  testAccCheckContactDestroy(ctx, t),
			},
		},
	})
}

func testAccContact_updateAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	oldAlias := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	newAlias := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_ssmcontacts_contact.contact_one"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_alias(oldAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, oldAlias),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_alias(newAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, newAlias),
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

func testAccContact_updateType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	personalType := "PERSONAL"
	escalationType := "ESCALATION"

	resourceName := "aws_ssmcontacts_contact.contact_one"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_type(name, personalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, personalType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_type(name, escalationType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, escalationType),
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

func testAccContact_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssmcontacts.ResourceContact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContact_updateDisplayName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	oldDisplayName := sdkacctest.RandString(26)
	newDisplayName := sdkacctest.RandString(26)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_displayName(rName, oldDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, oldDisplayName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_displayName(rName, newDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, newDisplayName),
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

func testAccCheckContactDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_contact" {
				continue
			}

			input := &ssmcontacts.GetContactInput{
				ContactId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetContact(ctx, input)

			if err != nil {
				// Getting resources may return validation exception when the replication set has been destroyed
				var ve *types.ValidationException
				if errors.As(err, &ve) {
					continue
				}

				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					continue
				}

				return err
			}

			return create.Error(names.SSMContacts, create.ErrActionCheckingDestroyed, tfssmcontacts.ResNameContact, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContactExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)

		_, err := conn.GetContact(ctx, &ssmcontacts.GetContactInput{
			ContactId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContactPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)

	input := &ssmcontacts.ListContactsInput{}
	_, err := conn.ListContacts(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccContactConfig_base() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccContactConfig_basic(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccContactConfig_none() string {
	return testAccContactConfig_base()
}

func testAccContactConfig_alias(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccContactConfig_type(name, typeValue string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = %[2]q

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, name, typeValue))
}

func testAccContactConfig_displayName(alias, displayName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias        = %[1]q
  display_name = %[2]q
  type         = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias, displayName))
}
