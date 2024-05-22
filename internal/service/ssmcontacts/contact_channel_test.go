// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccContactChannel_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test"
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "EMAIL"),
					resource.TestCheckResourceAttrPair(channelResourceName, "contact_id", contactResourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(channelResourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile("contact-channel/test-contact-for-"+rName+"/.")),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We need to explicitly test destroying this resource instead of just using CheckDestroy,
				// because CheckDestroy will run after the replication set has been destroyed and destroying
				// the replication set will destroy all other resources.
				Config: testAccContactChannelConfig_none(),
				Check:  testAccCheckContactChannelDestroy(ctx),
			},
		},
	})
}

func testAccContactChannel_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactChannelExists(ctx, channelResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmcontacts.ResourceContactChannel(), channelResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContactChannel_contactID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testContactOneResourceName := "aws_ssmcontacts_contact.test_contact_one"
	testContactTwoResourceName := "aws_ssmcontacts_contact.test_contact_two"
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig_withTwoContacts(rName, testContactOneResourceName+".arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, testContactOneResourceName),
					testAccCheckContactExists(ctx, testContactTwoResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttrPair(channelResourceName, "contact_id", testContactOneResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactChannelConfig_withTwoContacts(rName, testContactTwoResourceName+".arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, testContactOneResourceName),
					testAccCheckContactExists(ctx, testContactTwoResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttrPair(channelResourceName, "contact_id", testContactTwoResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccContactChannel_deliveryAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)
	contactResourceName := "aws_ssmcontacts_contact.test"
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig(rName, rName, "EMAIL", address1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", address1),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "EMAIL"),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactChannelConfig(rName, rName, "EMAIL", address2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", address2),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "EMAIL"),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccContactChannel_name(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix + acctest.Ct1)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix + acctest.Ct2)
	contactResourceName := "aws_ssmcontacts_contact.test"
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig(rName1, "update-name-test", "EMAIL", acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactChannelConfig(rName2, "update-name-test", "EMAIL", acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrName, rName2),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccContactChannel_type(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test"
	channelResourceName := "aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelConfig_defaultEmail(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "EMAIL"),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactChannelConfig_defaultSMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", "+12065550100"),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "SMS"),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactChannelConfig_defaultVoice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckContactChannelExists(ctx, channelResourceName),
					resource.TestCheckResourceAttr(channelResourceName, "activation_status", "NOT_ACTIVATED"),
					resource.TestCheckResourceAttr(channelResourceName, "delivery_address.0.simple_address", "+12065550199"),
					resource.TestCheckResourceAttr(channelResourceName, names.AttrType, "VOICE"),
				),
			},
			{
				ResourceName:      channelResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckContactChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_contact_channel" {
				continue
			}

			input := &ssmcontacts.GetContactChannelInput{
				ContactChannelId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetContactChannel(ctx, input)

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

			return create.Error(names.SSMContacts, create.ErrActionCheckingDestroyed, tfssmcontacts.ResNameContactChannel, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContactChannelExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContactChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContactChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)
		_, err := conn.GetContactChannel(ctx, &ssmcontacts.GetContactChannelInput{
			ContactChannelId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContactChannel, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContactChannelConfig_basic(rName string) string {
	return testAccContactChannelConfig_defaultEmail(rName)
}

func testAccContactChannelConfig_none() string {
	return testAccContactChannelConfig_base()
}

func testAccContactChannelConfig_defaultEmail(rName string) string {
	return testAccContactChannelConfig(rName, rName, "EMAIL", acctest.DefaultEmailAddress)
}

func testAccContactChannelConfig_defaultSMS(rName string) string {
	return testAccContactChannelConfig(rName, rName, "SMS", "+12065550100")
}

func testAccContactChannelConfig_defaultVoice(rName string) string {
	return testAccContactChannelConfig(rName, rName, "VOICE", "+12065550199")
}

func testAccContactChannelConfig(rName string, contactAliasDisambiguator string, channelType string, address string) string {
	return acctest.ConfigCompose(
		testAccContactChannelConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = %[4]q
  }

  name = %[1]q
  type = %[3]q
}
`, rName, contactAliasDisambiguator, channelType, address))
}

func testAccContactChannelConfig_withTwoContacts(rName, contactArn string) string {
	return acctest.ConfigCompose(
		testAccContactChannelConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact_one" {
  alias = "test-contact-one-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test_contact_two" {
  alias = "test-contact-two-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = %[2]s

  delivery_address {
    simple_address = %[3]q
  }

  name = %[1]q
  type = "EMAIL"
}
`, rName, contactArn, acctest.DefaultEmailAddress))
}

func testAccContactChannelConfig_base() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}
