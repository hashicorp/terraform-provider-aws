// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCTrafficMirrorFilterRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_traffic_mirror_filter_rule.test"
	dstCidr := "10.0.0.0/8"
	srcCidr := "0.0.0.0/0"
	ruleNum1 := 1
	ruleNum2 := 2
	action := "accept"
	direction := "ingress"
	description := "test rule"
	protocol := 6
	srcPortFrom := 32000
	srcPortTo := 64000
	dstPortFrom := 10000
	dstPortTo := 10001

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorFilterRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorFilterRuleDestroy(ctx),
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccVPCTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`traffic-mirror-filter-rule/tmfr-.+`)),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexache.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum1)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", acctest.Ct0),
				),
			},
			// Add all optionals
			{
				Config: testAccVPCTrafficMirrorFilterRuleConfig_full(dstCidr, srcCidr, action, direction, description, ruleNum1, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexache.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum1)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.from_port", strconv.Itoa(dstPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.to_port", strconv.Itoa(dstPortTo)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.from_port", strconv.Itoa(srcPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.to_port", strconv.Itoa(srcPortTo)),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, strconv.Itoa(protocol)),
				),
			},
			// Updates
			{
				Config: testAccVPCTrafficMirrorFilterRuleConfig_full(dstCidr, srcCidr, action, direction, description, ruleNum2, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexache.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum2)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.from_port", strconv.Itoa(dstPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.to_port", strconv.Itoa(dstPortTo)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.from_port", strconv.Itoa(srcPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.to_port", strconv.Itoa(srcPortTo)),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, strconv.Itoa(protocol)),
				),
			},
			// remove optionals
			{
				Config: testAccVPCTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexache.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum1)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrafficMirrorFilterRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCTrafficMirrorFilterRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_traffic_mirror_filter_rule.test"
	dstCidr := "10.0.0.0/8"
	srcCidr := "0.0.0.0/0"
	ruleNum := 1
	action := "accept"
	direction := "ingress"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrafficMirrorFilterRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorFilterRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTrafficMirrorFilterRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckTrafficMirrorFilterRule(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.DescribeTrafficMirrorFilters(ctx, &ec2.DescribeTrafficMirrorFiltersInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter rule acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorFilterRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_traffic_mirror_filter_rule" {
				continue
			}

			_, err := tfec2.FindTrafficMirrorFilterRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["traffic_mirror_filter_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Traffic Mirror Filter Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrafficMirrorFilterRuleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindTrafficMirrorFilterRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["traffic_mirror_filter_id"], rs.Primary.ID)

		return err
	}
}

func testAccTrafficMirrorFilterRuleImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["traffic_mirror_filter_id"], rs.Primary.ID), nil
	}
}

func testAccVPCTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, dir string, num int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
}

resource "aws_ec2_traffic_mirror_filter_rule" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  destination_cidr_block   = %[1]q
  rule_action              = %[2]q
  rule_number              = %[3]d
  source_cidr_block        = %[4]q
  traffic_direction        = %[5]q
}
`, dstCidr, action, num, srcCidr, dir)
}

func testAccVPCTrafficMirrorFilterRuleConfig_full(dstCidr, srcCidr, action, dir, description string, ruleNum, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {}

resource "aws_ec2_traffic_mirror_filter_rule" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  destination_cidr_block   = %[1]q
  rule_action              = %[2]q
  rule_number              = %[3]d
  source_cidr_block        = %[4]q
  traffic_direction        = %[5]q
  description              = %[6]q
  protocol                 = %[7]d
  source_port_range {
    from_port = %[8]d
    to_port   = %[9]d
  }
  destination_port_range {
    from_port = %[10]d
    to_port   = %[11]d
  }
}
`, dstCidr, action, ruleNum, srcCidr, dir, description, protocol, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo)
}
