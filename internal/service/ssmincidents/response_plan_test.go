// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResponsePlan_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_basic(rName, rTitle, acctest.Ct3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.impact", acctest.Ct3),

					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", fmt.Sprintf("response-plan/%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				// We need to explicitly test destroying this resource instead of just using CheckDestroy,
				// because CheckDestroy will run after the replication set has been destroyed and destroying
				// the replication set will destroy all other resources.
				Config: testAccResponsePlanConfig_none(),
				Check:  testAccCheckResponsePlanDestroy(ctx),
			},
		},
	})
}

func testAccResponsePlan_updateRequiredFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	iniName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	iniTitle := "initialTitle"
	updTitle := "updatedTitle"
	updImpact := "5"

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_basic(iniName, iniTitle, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, iniName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", iniTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.impact", acctest.Ct1),

					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", fmt.Sprintf("response-plan/%s", iniName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_basic(iniName, updTitle, updImpact),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, iniName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", updTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.impact", updImpact),

					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", fmt.Sprintf("response-plan/%s", iniName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_basic(updName, updTitle, updImpact),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", updTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.impact", updImpact),

					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", fmt.Sprintf("response-plan/%s", updName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_updateTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	rKey1 := sdkacctest.RandString(26)
	rVal1Ini := sdkacctest.RandString(26)
	rVal1Upd := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)
	rVal2 := sdkacctest.RandString(26)
	rKey3 := sdkacctest.RandString(26)
	rVal3 := sdkacctest.RandString(26)

	rProviderKey1 := sdkacctest.RandString(26)
	rProviderVal1Ini := sdkacctest.RandString(26)
	rProviderVal1Upd := sdkacctest.RandString(26)
	rProviderKey2 := sdkacctest.RandString(26)
	rProviderVal2 := sdkacctest.RandString(26)
	rProviderKey3 := sdkacctest.RandString(26)
	rProviderVal3 := sdkacctest.RandString(26)

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(rProviderKey1, rProviderVal1Ini),
					testAccResponsePlanConfig_oneTag(rName, rTitle, rKey1, rVal1Ini),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Ini),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Ini),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(rProviderKey1, rProviderVal1Upd),
					testAccResponsePlanConfig_oneTag(rName, rTitle, rKey1, rVal1Upd),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Upd),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Upd),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2(rProviderKey2, rProviderVal2, rProviderKey3, rProviderVal3),
					testAccResponsePlanConfig_twoTags(rName, rTitle, rKey2, rVal2, rKey3, rVal3),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, rVal2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey3, rVal3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey2, rProviderVal2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey3, rProviderVal3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_updateEmptyTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	rKey1 := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_oneTag(rName, rTitle, rKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_twoTags(rName, rTitle, rKey1, "", rKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_oneTag(rName, rTitle, rKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_basic(rName, rTitle, acctest.Ct3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmincidents.ResourceResponsePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResponsePlan_incidentTemplateOptionalFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	rDedupeStringIni := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDedupeStringUpd := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSummaryIni := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSummaryUpd := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTagKeyIni := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTagValIni := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTagKeyUpd := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTagValUpd := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	snsTopic1 := "aws_sns_topic.topic1"
	snsTopic2 := "aws_sns_topic.topic2"
	snsTopic3 := "aws_sns_topic.topic3"

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_incidentTemplateOptionalFields(rName, rTitle, rDedupeStringIni, rSummaryIni, rTagKeyIni, rTagValIni, snsTopic1, snsTopic2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.dedupe_string", rDedupeStringIni),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.summary", rSummaryIni),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.incident_tags."+rTagKeyIni, rTagValIni),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic2, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_incidentTemplateOptionalFields(rName, rTitle, rDedupeStringUpd, rSummaryUpd, rTagKeyUpd, rTagValUpd, snsTopic2, snsTopic3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.dedupe_string", rDedupeStringUpd),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.summary", rSummaryUpd),
					resource.TestCheckResourceAttr(resourceName, "incident_template.0.incident_tags."+rTagKeyUpd, rTagValUpd),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic2, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic3, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_displayName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	oldDisplayName := rName + "-old-display-name"
	newDisplayName := rName + "-new-display-name"

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_displayName(rName, oldDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, oldDisplayName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_displayName(rName, newDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, newDisplayName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_chatChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	chatChannelTopic1 := "aws_sns_topic.topic1"
	chatChannelTopic2 := "aws_sns_topic.topic2"

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_chatChannel(rName, chatChannelTopic1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "chat_channel.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "chat_channel.0", chatChannelTopic1, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_chatChannel(rName, chatChannelTopic2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "chat_channel.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "chat_channel.0", chatChannelTopic2, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_twoChatChannels(rName, chatChannelTopic1, chatChannelTopic2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "chat_channel.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "chat_channel.*", chatChannelTopic1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "chat_channel.*", chatChannelTopic2, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_emptyChatChannel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "chat_channel.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_engagement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	//lintignore:AWSAT003
	//lintignore:AWSAT005
	contactArn1 := "arn:aws:ssm-contacts:us-east-2:111122223333:contact/test1"
	//lintignore:AWSAT003
	//lintignore:AWSAT005
	contactArn2 := "arn:aws:ssm-contacts:us-east-2:111122223333:contact/test2"

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_engagement(rName, contactArn1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "engagements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "engagements.0", contactArn1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_engagement(rName, contactArn2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "engagements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "engagements.0", contactArn2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_twoEngagements(rName, contactArn1, contactArn2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "engagements.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "engagements.0", contactArn1),
					resource.TestCheckResourceAttr(resourceName, "engagements.1", contactArn2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_emptyEngagements(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "engagements.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

func testAccResponsePlan_action(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_ssmincidents_response_plan.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanConfig_action1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.ssm_automation.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName,
						"action.0.ssm_automation.0.document_name",
						"aws_ssm_document.document1",
						names.AttrName,
					),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName,
						"action.0.ssm_automation.0.role_arn",
						"aws_iam_role.role1",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.document_version",
						"version1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.target_account",
						"RESPONSE_PLAN_OWNER_ACCOUNT",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.name",
						names.AttrKey,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.#",
						acctest.Ct2,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.0",
						acctest.CtValue1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.1",
						acctest.CtValue2,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.dynamic_parameters.anotherKey",
						"INVOLVED_RESOURCES",
					),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
			{
				Config: testAccResponsePlanConfig_action2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponsePlanExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.ssm_automation.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName,
						"action.0.ssm_automation.0.document_name",
						"aws_ssm_document.document2",
						names.AttrName,
					),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName,
						"action.0.ssm_automation.0.role_arn",
						"aws_iam_role.role2",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.document_version",
						"version2",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.target_account",
						"IMPACTED_ACCOUNT",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.name",
						"foo",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.#",
						acctest.Ct1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.0",
						"bar",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"action.0.ssm_automation.0.dynamic_parameters.someKey",
						"INCIDENT_RECORD_ARN",
					),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_set_arn"},
			},
		},
	})
}

//
//	Comment out integration test as the configured PagerDuty secretId is invalid and the test will fail,
//	as we do not want to expose credentials to public repository.
//
//	Tested locally and PagerDuty integration work with response plan.
//
//func testResponsePlan_integration(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//  ctx := acctest.Context(t)
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//
//	resourceName := "aws_ssmincidents_response_plan.test"
//	pagerdutyName := "pagerduty-test-terraform"
//	pagerdutyServiceId := "example"
//	pagerdutySecretId := "example"
//
//	resource.Test(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckResponsePlanDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccResponsePlanConfig_pagerdutyIntegration(
//					rName,
//					pagerdutyName,
//					pagerdutyServiceId,
//					pagerdutySecretId,
//				),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckResponsePlanExists(ctx,resourceName),
//					resource.TestCheckResourceAttr(resourceName, "integration.#", "1"),
//					resource.TestCheckResourceAttr(resourceName, "integration.0.pagerduty.#", "1"),
//					resource.TestCheckResourceAttr(
//						resourceName,
//						"integration.0.pagerduty.0.name",
//						pagerdutyName,
//					),
//					resource.TestCheckResourceAttr(
//						resourceName,
//						"integration.0.pagerduty.0.service_id",
//						pagerdutyServiceId,
//					),
//					resource.TestCheckResourceAttr(
//						resourceName,
//						"integration.0.pagerduty.0.secret_id",
//						pagerdutySecretId,
//					),
//				),
//			},
//			{
//				ResourceName:            resourceName,
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"replication_set_arn"},
//			},
//		},
//	})
//}

func testAccCheckResponsePlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient(ctx)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "aws_ssmincidents_response_plan" {
				continue
			}

			_, err := tfssmincidents.FindResponsePlanByID(ctx, client, resource.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameResponsePlan, resource.Primary.ID,
					errors.New("expected resource not found error, received an unexpected error"))
			}

			return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameResponsePlan, resource.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResponsePlanExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameResponsePlan, name, errors.New("not found"))
		}

		if resource.Primary.ID == "" {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameResponsePlan, name, errors.New("not set"))
		}

		client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient(ctx)

		_, err := tfssmincidents.FindResponsePlanByID(ctx, client, resource.Primary.ID)

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameResponsePlan, resource.Primary.ID, err)
		}

		return nil
	}
}

func testAccResponsePlanConfig_base() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test_replication_set" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccResponsePlanConfig_baseSNSTopic() string {
	return `
resource "aws_sns_topic" "topic1" {}
resource "aws_sns_topic" "topic2" {}
resource "aws_sns_topic" "topic3" {}
`
}

func testAccResponsePlanConfig_basic(name, title, impact string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[2]q
    impact = %[3]q
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, title, impact))
}

func testAccResponsePlanConfig_none() string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
	)
}

func testAccResponsePlanConfig_oneTag(name, title, tagKey, tagVal string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[2]q
    impact = "3"
  }

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, title, tagKey, tagVal))
}

func testAccResponsePlanConfig_twoTags(name, title, tag1Key, tag1Val, tag2Key, tag2Val string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[2]q
    impact = "3"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, title, tag1Key, tag1Val, tag2Key, tag2Val))
}

func testAccResponsePlanConfig_incidentTemplateOptionalFields(name, title, dedupeString, summary, tagKey, tagVal, snsTopic1, snsTopic2 string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		testAccResponsePlanConfig_baseSNSTopic(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title         = %[2]q
    impact        = "3"
    dedupe_string = %[3]q
    summary       = %[4]q

    incident_tags = {
      %[5]q = %[6]q
    }

    notification_target {
      sns_topic_arn = %[7]s
    }

    notification_target {
      sns_topic_arn = %[8]s
    }
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, title, dedupeString, summary, tagKey, tagVal, snsTopic1+".arn", snsTopic2+".arn"))
}

func testAccResponsePlanConfig_displayName(name, displayName string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  display_name = %[2]q

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, displayName))
}

func testAccResponsePlanConfig_chatChannel(name, chatChannelTopic string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		testAccResponsePlanConfig_baseSNSTopic(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  chat_channel = [%[2]s]

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, chatChannelTopic+".arn"))
}

func testAccResponsePlanConfig_twoChatChannels(name, chatChannelOneTopic, chatChannelTwoTopic string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		testAccResponsePlanConfig_baseSNSTopic(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  chat_channel = [%[2]s, %[3]s]

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, chatChannelOneTopic+".arn", chatChannelTwoTopic+".arn"))
}

func testAccResponsePlanConfig_emptyChatChannel(name string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  chat_channel = []

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name))
}

func testAccResponsePlanConfig_engagement(name, contactArn string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  engagements = [%[2]q]

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, contactArn))
}

func testAccResponsePlanConfig_twoEngagements(name, contactArn1, contactArn2 string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  engagements = [%[2]q, %[3]q]

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name, contactArn1, contactArn2))
}

func testAccResponsePlanConfig_emptyEngagements(name string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  engagements = []

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name))
}

func testAccResponsePlanConfig_action1(name string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		testAccResponsePlanConfig_baseIAMRole(name),
		testAccResponsePlanConfig_baseSSMDocument(name),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  action {
    ssm_automation {
      document_name    = aws_ssm_document.document1.name
      role_arn         = aws_iam_role.role1.arn
      document_version = "version1"
      target_account   = "RESPONSE_PLAN_OWNER_ACCOUNT"
      parameter {
        name   = "key"
        values = ["value1", "value2"]
      }
      dynamic_parameters = {
        anotherKey = "INVOLVED_RESOURCES"
      }
    }
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name))
}

func testAccResponsePlanConfig_action2(name string) string {
	return acctest.ConfigCompose(
		testAccResponsePlanConfig_base(),
		testAccResponsePlanConfig_baseIAMRole(name),
		testAccResponsePlanConfig_baseSSMDocument(name),
		fmt.Sprintf(`
resource "aws_ssmincidents_response_plan" "test" {
  name = %[1]q

  incident_template {
    title  = %[1]q
    impact = "1"
  }

  action {
    ssm_automation {
      document_name    = aws_ssm_document.document2.name
      role_arn         = aws_iam_role.role2.arn
      document_version = "version2"
      target_account   = "IMPACTED_ACCOUNT"
      parameter {
        name   = "foo"
        values = ["bar"]
      }
      dynamic_parameters = {
        someKey = "INCIDENT_RECORD_ARN"
      }
    }
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}
`, name))
}

func testAccResponsePlanConfig_baseIAMRole(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role1" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_iam_role" "role2" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[2]q
}
`, name+"-role-one", name+"-role-two")
}

func testAccResponsePlanConfig_baseSSMDocument(name string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "document1" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_document" "document2" {
  name          = %[2]q
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}
`, name+"-test-documen-one", name+"-test-documen-two")
}

//func testAccResponsePlanConfig_pagerdutyIntegration(
//	name,
//	pagerdutyName,
//	pagerdutyServiceId,
//	pagerdutySecretId string) string {
//	return acctest.ConfigCompose(
//		testAccResponsePlanConfigBase(),
//		fmt.Sprintf(`
//resource "aws_ssmincidents_response_plan" "test" {
//  name = %[1]q
//
//  incident_template {
//    title  = %[1]q
//    impact = "1"
//  }
//
//  integration {
//    pagerduty {
//      name       = %[2]q
//      service_id = %[3]q
//      secret_id  = %[4]q
//    }
//  }
//
//  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
//}
//`, name, pagerdutyName, pagerdutyServiceId, pagerdutySecretId))
//}
