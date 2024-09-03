// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

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

func testAccPlan_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test_contact_one"
	planResourceName := "aws_ssmcontacts_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_oneStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.duration_in_minutes", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"contact_id",
						"ssm-contacts",
						"contact/test-contact-one-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We need to explicitly test destroying this resource instead of just using CheckDestroy,
				// because CheckDestroy will run after the replication set has been destroyed and destroying
				// the replication set will destroy all other resources.
				Config: testAccPlanConfig_none(rName),
				Check:  testAccCheckPlanDestroy(ctx),
			},
		},
	})
}

func testAccPlan_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test_contact_one"
	planResourceName := "aws_ssmcontacts_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_oneStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmcontacts.ResourcePlan(), planResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPlan_updateContactId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactOneResourceName := "aws_ssmcontacts_contact.test_contact_one"
	contactTwoResourceName := "aws_ssmcontacts_contact.test_contact_two"

	planResourceName := "aws_ssmcontacts_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_contactID(rName, contactOneResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactOneResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckTypeSetElemAttrPair(planResourceName, "contact_id", contactOneResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_contactID(rName, contactTwoResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactTwoResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckTypeSetElemAttrPair(planResourceName, "contact_id", contactTwoResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlan_updateStages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test_contact_one"
	planResourceName := "aws_ssmcontacts_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_oneStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.duration_in_minutes", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_twoStages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.#", acctest.Ct2),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.duration_in_minutes", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct0),
					resource.TestCheckResourceAttr(planResourceName, "stage.1.duration_in_minutes", acctest.Ct2),
					resource.TestCheckResourceAttr(planResourceName, "stage.1.target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_oneStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.duration_in_minutes", acctest.Ct1),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlan_updateDurationInMinutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactResourceName := "aws_ssmcontacts_contact.test_contact_one"
	planResourceName := "aws_ssmcontacts_plan.test"
	oldDurationInMinutes := 1
	newDurationInMinutes := 2

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_durationInMinutes(rName, oldDurationInMinutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.duration_in_minutes",
						strconv.Itoa(oldDurationInMinutes),
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_durationInMinutes(rName, newDurationInMinutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.duration_in_minutes",
						strconv.Itoa(newDurationInMinutes),
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlan_updateTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	escalationPlanResourceName := "aws_ssmcontacts_contact.test_escalation_plan_one"
	planResourceName := "aws_ssmcontacts_plan.test"
	testContactOneResourceArn := "aws_ssmcontacts_contact.test_contact_one.arn"
	testContactTwoResourceArn := "aws_ssmcontacts_contact.test_contact_two.arn"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_oneTarget(rName, testContactOneResourceArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.is_essential",
						acctest.CtFalse,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-one-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_twoTargets(rName, testContactOneResourceArn, testContactTwoResourceArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.is_essential",
						acctest.CtFalse,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-one-for-"+rName,
					),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.1.contact_target_info.0.is_essential",
						acctest.CtTrue,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.1.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-two-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_oneTarget(rName, testContactOneResourceArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.is_essential",
						acctest.CtFalse,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-one-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlan_updateContactTargetInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	escalationPlanResourceName := "aws_ssmcontacts_contact.test_escalation_plan_one"
	planResourceName := "aws_ssmcontacts_plan.test"
	testContactOneResourceArn := "aws_ssmcontacts_contact.test_contact_one.arn"
	testContactTwoResourceArn := "aws_ssmcontacts_contact.test_contact_two.arn"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_contactTargetInfo(rName, false, testContactOneResourceArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.is_essential",
						acctest.CtFalse,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-one-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_contactTargetInfo(rName, true, testContactTwoResourceArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.is_essential",
						acctest.CtTrue,
					),
					acctest.CheckResourceAttrRegionalARN(
						planResourceName,
						"stage.0.target.0.contact_target_info.0.contact_id",
						"ssm-contacts",
						"contact/test-contact-two-for-"+rName,
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlan_updateChannelTargetInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	escalationPlanResourceName := "aws_ssmcontacts_contact.test_escalation_plan_one"
	planResourceName := "aws_ssmcontacts_plan.test"
	contactChannelOneResourceName := "aws_ssmcontacts_contact_channel.test_channel_one"
	contactChannelTwoResourceName := "aws_ssmcontacts_contact_channel.test_channel_two"

	oldRetryIntervalInMinutes := 3
	newRetryIntervalInMinutes := 5

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig_channelTargetInfo(
					rName,
					contactChannelOneResourceName,
					oldRetryIntervalInMinutes,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(
						planResourceName,
						"stage.0.target.0.channel_target_info.0.contact_channel_id",
						contactChannelOneResourceName,
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.channel_target_info.0.retry_interval_in_minutes",
						strconv.Itoa(oldRetryIntervalInMinutes),
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlanConfig_channelTargetInfo(
					rName,
					contactChannelTwoResourceName,
					newRetryIntervalInMinutes,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, escalationPlanResourceName),
					testAccCheckPlanExists(ctx, planResourceName),
					resource.TestCheckResourceAttr(planResourceName, "stage.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(
						planResourceName,
						"stage.0.target.0.channel_target_info.0.contact_channel_id",
						contactChannelTwoResourceName,
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(
						planResourceName,
						"stage.0.target.0.channel_target_info.0.retry_interval_in_minutes",
						strconv.Itoa(newRetryIntervalInMinutes),
					),
				),
			},
			{
				ResourceName:      planResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPlanExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNamePlan, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNamePlan, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		output, err := conn.GetContact(ctx, &ssmcontacts.GetContactInput{
			ContactId: aws.String(rs.Primary.ID),
		})

		if err != nil || len(output.Plan.Stages) == 0 {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNamePlan, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_plan" {
				continue
			}

			input := &ssmcontacts.GetContactInput{
				ContactId: aws.String(rs.Primary.ID),
			}
			output, err := conn.GetContact(ctx, input)
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

			if len(output.Plan.Stages) == 0 {
				return nil
			}

			return create.Error(names.SSMContacts, create.ErrActionCheckingDestroyed, tfssmcontacts.ResNamePlan, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccPlanConfig_contactID(rName, contactName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = %[1]s

  stage {
    duration_in_minutes = 1
  }
}
`, contactName+".arn"))
}

func testAccPlanConfig_none(rName string) string {
	return testAccPlanConfig_base(rName)
}

func testAccPlanConfig_oneStage(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  stage {
    duration_in_minutes = 1
  }
}
`)
}

func testAccPlanConfig_twoStages(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  stage {
    duration_in_minutes = 1
  }

  stage {
    duration_in_minutes = 2
  }
}
`)
}

func testAccPlanConfig_durationInMinutes(rName string, durationInMinutes int) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  stage {
    duration_in_minutes = %[1]d
  }
}
`, durationInMinutes))
}

func testAccPlanConfig_oneTarget(rName, contactOneArn string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_escalation_plan_one.arn

  stage {
    duration_in_minutes = 0

    target {
      contact_target_info {
        is_essential = false
        contact_id   = %[1]s
      }
    }
  }
}
`, contactOneArn))
}

func testAccPlanConfig_twoTargets(rName, contactOneArn, contactTwoArn string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_escalation_plan_one.arn

  stage {
    duration_in_minutes = 0

    target {
      contact_target_info {
        is_essential = false
        contact_id   = %[1]s
      }
    }

    target {
      contact_target_info {
        is_essential = true
        contact_id   = %[2]s
      }
    }
  }
}
`, contactOneArn, contactTwoArn))
}

func testAccPlanConfig_contactTargetInfo(rName string, isEssential bool, contactId string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_escalation_plan_one.arn

  stage {
    duration_in_minutes = 0

    target {
      contact_target_info {
        is_essential = %[1]t
        contact_id   = %[2]s
      }
    }
  }
}
`, isEssential, contactId))
}

func testAccPlanConfig_channelTargetInfo(rName, contactChannelResourceName string, retryIntervalInMinutes int) string {
	domain := acctest.RandomDomainName()

	return acctest.ConfigCompose(
		testAccPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact_channel" "test_channel_one" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  delivery_address {
    simple_address = %[3]q
  }

  name = "Test Contact Channel 1"
  type = "EMAIL"
}

resource "aws_ssmcontacts_contact_channel" "test_channel_two" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  delivery_address {
    simple_address = %[4]q
  }

  name = "Test Contact Channel 2"
  type = "EMAIL"
}

resource "aws_ssmcontacts_plan" "test" {
  contact_id = aws_ssmcontacts_contact.test_contact_one.arn

  stage {
    duration_in_minutes = 1

    target {
      channel_target_info {
        contact_channel_id        = %[1]s.arn
        retry_interval_in_minutes = %[2]d
      }
    }
  }
}
`, contactChannelResourceName, retryIntervalInMinutes, acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)))
}

func testAccPlanConfig_base(alias string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}

resource "aws_ssmcontacts_contact" "test_contact_one" {
  alias = "test-contact-one-for-%[2]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test_contact_two" {
  alias = "test-contact-two-for-%[2]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test_escalation_plan_one" {
  alias = "test-escalation-plan-for-%[2]s"
  type  = "ESCALATION"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, acctest.Region(), alias)
}
