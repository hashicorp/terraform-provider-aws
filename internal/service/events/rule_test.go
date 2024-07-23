// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/google/go-cmp/cmp"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.EventsServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Operation is disabled in this region",
		"not a supported service for a target",
	)
}

func TestRuleParseResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputID       string
		ExpectedError bool
		ExpectedPart0 string
		ExpectedPart1 string
	}{
		{
			TestName:      "empty ID",
			InputID:       "",
			ExpectedError: true,
		},
		{
			TestName:      "single part",
			InputID:       "TestRule",
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "two parts",
			InputID:       tfevents.RuleCreateResourceID("TestEventBus", "TestRule"),
			ExpectedPart0: "TestEventBus",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "two parts with default event bus",
			InputID:       tfevents.RuleCreateResourceID(tfevents.DefaultEventBusName, "TestRule"),
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "partner event bus 1",
			InputID:       "aws.partner/example.com/Test/TestRule",
			ExpectedPart0: "aws.partner/example.com/Test",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "partner event bus 2",
			InputID:       "aws.partner/example.net/id/18554d09-58ff-aa42-ba9c-c4c33899006f/test",
			ExpectedPart0: "aws.partner/example.net/id/18554d09-58ff-aa42-ba9c-c4c33899006f",
			ExpectedPart1: "test",
		},
		{
			TestName: "ARN event bus",
			//lintignore:AWSAT003,AWSAT005
			InputID: tfevents.RuleCreateResourceID("arn:aws:events:us-east-2:123456789012:event-bus/default", "TestRule"),
			//lintignore:AWSAT003,AWSAT005
			ExpectedPart0: "arn:aws:events:us-east-2:123456789012:event-bus/default",
			ExpectedPart1: "TestRule",
		},
		{
			TestName: "ARN based partner event bus",
			// lintignore:AWSAT003,AWSAT005
			InputID: "arn:aws:events:us-east-2:123456789012:event-bus/aws.partner/genesys.com/cloud/a12bc345-d678-90e1-2f34-gh5678i9012ej/_genesys/TestRule",
			// lintignore:AWSAT003,AWSAT005
			ExpectedPart0: "arn:aws:events:us-east-2:123456789012:event-bus/aws.partner/genesys.com/cloud/a12bc345-d678-90e1-2f34-gh5678i9012ej/_genesys",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "empty both parts",
			InputID:       "/",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part",
			InputID:       "/TestRule",
			ExpectedError: true,
		},
		{
			TestName:      "empty second part",
			InputID:       "TestEventBus/",
			ExpectedError: true,
		},
		{
			TestName:      "empty partner event rule",
			InputID:       "aws.partner/example.com/Test/",
			ExpectedError: true,
		},
		{
			TestName:      "three parts",
			InputID:       "TestEventBus/TestRule/Suffix",
			ExpectedError: true,
		},
		{
			TestName:      "four parts",
			InputID:       "abc.partner/TestEventBus/TestRule/Suffix",
			ExpectedError: true,
		},
		{
			TestName:      "five parts",
			InputID:       "test/aws.partner/example.com/Test/TestRule",
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotPart0, gotPart1, err := tfevents.RuleParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotPart0 != testCase.ExpectedPart0 {
				t.Errorf("got part 0 %s, expected %s", gotPart0, testCase.ExpectedPart0)
			}

			if gotPart1 != testCase.ExpectedPart1 {
				t.Errorf("got part 1 %s, expected %s", gotPart1, testCase.ExpectedPart1)
			}
		})
	}
}

func TestRuleEventPatternJSONDecoder(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    string
		expected string
	}
	tests := map[string]testCase{
		"lessThanGreaterThan": {
			input:    `{"detail":{"count":[{"numeric":["\u003e",0,"\u003c",5]}]}}`,
			expected: `{"detail":{"count":[{"numeric":[">",0,"<",5]}]}}`,
		},
		"ampersand": {
			input:    `{"detail":{"count":[{"numeric":["\u0026",0,"\u0026",5]}]}}`,
			expected: `{"detail":{"count":[{"numeric":["&",0,"&",5]}]}}`,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfevents.RuleEventPatternJSONDecoder(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestAccEventsRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeRuleOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s$`, rName1))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "event_pattern"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					testAccCheckRuleEnabled(ctx, resourceName, "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccRuleNoBusNameImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					testAccCheckRuleRecreated(&v1, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					testAccCheckRuleEnabled(ctx, resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccRuleConfig_defaultBusName(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v3),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					testAccCheckRuleNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
				),
			},
		},
	})
}

func TestAccEventsRule_eventBusName(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeRuleOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	busName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	busName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_busName(rName1, busName1, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName1, rName1))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_busName(rName1, busName1, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					testAccCheckRuleNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName1),
				),
			},
			{
				Config: testAccRuleConfig_busName(rName2, busName2, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v3),
					testAccCheckRuleRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName2, rName2))),
				),
			},
		},
	})
}

func TestAccEventsRule_role(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_role(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_description(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccEventsRule_pattern(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_pattern(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, ""),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_pattern(rName, "{\"source\":[\"aws.lambda\"]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.lambda\"]}"),
				),
			},
		},
	})
}

func TestAccEventsRule_patternJSONEncoder(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_patternJSONEncoder(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, ""),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", `{"detail":{"count":[{"numeric":[">",0,"<",5]}]}}`),
				),
			},
		},
	})
}

func TestAccEventsRule_scheduleAndPattern(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_scheduleAndPattern(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 hour)"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeRuleOutput
	rName := "tf-acc-test-prefix-"
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_namePrefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_Name_generated(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_nameGenerated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEventsRule_isEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_isEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
					testAccCheckRuleEnabled(ctx, resourceName, "DISABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_isEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					testAccCheckRuleEnabled(ctx, resourceName, "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_isEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
					testAccCheckRuleEnabled(ctx, resourceName, "DISABLED"),
				),
			},
		},
	})
}

func TestAccEventsRule_state(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_state(rName, string(types.RuleStateDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.RuleStateDisabled)),
					testAccCheckRuleEnabled(ctx, resourceName, types.RuleStateDisabled),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccRuleConfig_state(rName, string(types.RuleStateEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.RuleStateEnabled)),
					testAccCheckRuleEnabled(ctx, resourceName, types.RuleStateEnabled),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_partnerEventBus(t *testing.T) {
	ctx := acctest.Context(t)
	key := "EVENT_BRIDGE_PARTNER_EVENT_BUS_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_partnerBus(rName, busName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_eventBusARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"
	eventBusName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_busARN(rName, eventBusName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf(`rule/%s/%s$`, eventBusName, rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrPair(resourceName, "event_bus_name", "aws_cloudwatch_event_bus.test", names.AttrARN),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccEventsRule_migrateV0(t *testing.T) {
	const resourceName = "aws_cloudwatch_event_rule.test"

	t.Parallel()

	testcases := map[string]struct {
		config            string
		expectedIsEnabled string
		expectedState     types.RuleState
	}{
		acctest.CtBasic: {
			config:            testAccRuleConfig_basic(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)),
			expectedIsEnabled: acctest.CtTrue,
			expectedState:     "ENABLED",
		},

		names.AttrEnabled: {
			config:            testAccRuleConfig_isEnabled(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), true),
			expectedIsEnabled: acctest.CtTrue,
			expectedState:     "ENABLED",
		},

		"disabled": {
			config:            testAccRuleConfig_isEnabled(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), false),
			expectedIsEnabled: acctest.CtFalse,
			expectedState:     "DISABLED",
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v eventbridge.DescribeRuleOutput

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:   acctest.ErrorCheck(t, names.EventsServiceID),
				CheckDestroy: testAccCheckRuleDestroy(ctx),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"aws": {
								Source:            "hashicorp/aws",
								VersionConstraint: "5.26.0",
							},
						},
						Config: testcase.config,
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckRuleExists(ctx, resourceName, &v),
							resource.TestCheckResourceAttr(resourceName, "is_enabled", testcase.expectedIsEnabled),
							testAccCheckRuleEnabled(ctx, resourceName, testcase.expectedState),
						),
					},
					{
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						Config:                   testcase.config,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectEmptyPlan(),
							},
						},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName, "is_enabled", testcase.expectedIsEnabled),
							resource.TestCheckResourceAttr(resourceName, names.AttrState, string(testcase.expectedState)),
							testAccCheckRuleEnabled(ctx, resourceName, testcase.expectedState),
						),
					},
				},
			})
		})
	}
}

func TestAccEventsRule_migrateV0_Equivalent(t *testing.T) {
	const resourceName = "aws_cloudwatch_event_rule.test"

	t.Parallel()

	testcases := map[string]struct {
		enabled           bool
		state             string
		expectedIsEnabled string
		expectedState     types.RuleState
	}{
		names.AttrEnabled: {
			enabled:           true,
			state:             string(types.RuleStateEnabled),
			expectedIsEnabled: acctest.CtTrue,
			expectedState:     types.RuleStateEnabled,
		},

		"disabled": {
			enabled:           false,
			state:             string(types.RuleStateDisabled),
			expectedIsEnabled: acctest.CtFalse,
			expectedState:     types.RuleStateDisabled,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
			var v eventbridge.DescribeRuleOutput

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:   acctest.ErrorCheck(t, names.EventsServiceID),
				CheckDestroy: testAccCheckRuleDestroy(ctx),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"aws": {
								Source:            "hashicorp/aws",
								VersionConstraint: "5.26.0",
							},
						},
						Config: testAccRuleConfig_isEnabled(rName, testcase.enabled),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckRuleExists(ctx, resourceName, &v),
							resource.TestCheckResourceAttr(resourceName, "is_enabled", testcase.expectedIsEnabled),
							testAccCheckRuleEnabled(ctx, resourceName, testcase.expectedState),
						),
					},
					{
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						Config:                   testAccRuleConfig_state(rName, testcase.state),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectEmptyPlan(),
							},
						},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName, "is_enabled", testcase.expectedIsEnabled),
							resource.TestCheckResourceAttr(resourceName, names.AttrState, string(testcase.expectedState)),
							testAccCheckRuleEnabled(ctx, resourceName, testcase.expectedState),
						),
					},
				},
			})
		})
	}
}

func testAccCheckRuleExists(ctx context.Context, n string, v *eventbridge.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		output, err := tfevents.FindRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["event_bus_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRuleEnabled(ctx context.Context, n string, want types.RuleState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		output, err := tfevents.FindRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["event_bus_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		if got := output.State; got != want {
			return fmt.Errorf("EventBridge Rule State = %v, want %v", got, want)
		}

		return nil
	}
}

func testAccCheckRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_rule" {
				continue
			}

			_, err := tfevents.FindRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["event_bus_name"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleRecreated(i, j *eventbridge.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.Arn) == aws.ToString(j.Arn) {
			return fmt.Errorf("EventBridge rule not recreated, but expected it to be")
		}
		return nil
	}
}

func testAccCheckRuleNotRecreated(i, j *eventbridge.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.Arn) != aws.ToString(j.Arn) {
			return fmt.Errorf("EventBridge rule recreated, but expected it to not be")
		}
		return nil
	}
}

func testAccRuleNoBusNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
}
`, rName)
}

func testAccRuleConfig_defaultBusName(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  event_bus_name      = "default"
}
`, rName)
}

func testAccRuleConfig_busName(rName, eventBusName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  description    = %[2]q
  event_pattern  = <<PATTERN
{
	"source": [
		"aws.ec2"
	]
}
PATTERN
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[3]q
}
`, rName, description, eventBusName)
}

func testAccRuleConfig_pattern(rName, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name          = %[1]q
  event_pattern = <<PATTERN
	%[2]s
PATTERN
}
`, rName, pattern)
}

func testAccRuleConfig_patternJSONEncoder(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name          = %[1]q
  event_pattern = jsonencode({ "detail" : { "count" : [{ "numeric" : [">", 0, "<", 5] }] } })
}
`, rName)
}

func testAccRuleConfig_scheduleAndPattern(rName, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  event_pattern       = <<PATTERN
	%[2]s
PATTERN
}
`, rName, pattern)
}

func testAccRuleConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  description         = %[2]q
  schedule_expression = "rate(1 hour)"
}
`, rName, description)
}

func testAccRuleConfig_isEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  is_enabled          = %[2]t
}
`, rName, enabled)
}

func testAccRuleConfig_state(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  state               = %[2]q
}
`, rName, state)
}

func testAccRuleConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name_prefix         = %[1]q
  schedule_expression = "rate(5 minutes)"
}
`, namePrefix)
}

const testAccRuleConfig_nameGenerated = `
resource "aws_cloudwatch_event_rule" "test" {
  schedule_expression = "rate(5 minutes)"
}
`

func testAccRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRuleConfig_role(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  role_arn            = aws_iam_role.test.arn
}
`, rName)
}

func testAccRuleConfig_partnerBus(rName, eventBusName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = %[2]q

  event_pattern = <<PATTERN
{
  "source": ["aws.ec2"]
}
PATTERN
}
`, rName, eventBusName)
}

func testAccRuleConfig_busARN(rName, eventBusName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[2]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.arn

  event_pattern = <<PATTERN
{
  "source": ["aws.ec2"]
}
PATTERN
}
`, rName, eventBusName)
}
