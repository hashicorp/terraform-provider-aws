package ec2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCTrafficMirrorFilterRule_basic(t *testing.T) {
	resourceName := "aws_ec2_traffic_mirror_filter_rule.test"
	dstCidr := "10.0.0.0/8"
	srcCidr := "0.0.0.0/0"
	ruleNum := 1
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
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilterRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorFilterRuleDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", ec2.ServiceName, regexp.MustCompile(`traffic-mirror-filter-rule/tmfr-.+`)),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckNoResourceAttr(resourceName, "destination_port_range"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "source_port_range"),
				),
			},
			// Add all optionals
			{
				Config: testAccTrafficMirrorFilterRuleConfig_full(dstCidr, srcCidr, action, direction, description, ruleNum, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.from_port", strconv.Itoa(dstPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.to_port", strconv.Itoa(dstPortTo)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.from_port", strconv.Itoa(srcPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.to_port", strconv.Itoa(srcPortTo)),
					resource.TestCheckResourceAttr(resourceName, "protocol", strconv.Itoa(protocol)),
				),
			},
			// remove optionals
			{
				Config: testAccTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", "0"),
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
	resourceName := "aws_ec2_traffic_mirror_filter_rule.test"
	dstCidr := "10.0.0.0/8"
	srcCidr := "0.0.0.0/0"
	ruleNum := 1
	action := "accept"
	direction := "ingress"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilterRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTrafficMirrorFilterRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrafficMirrorFilterRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		ruleId := rs.Primary.ID
		filterId := rs.Primary.Attributes["traffic_mirror_filter_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(filterId),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return fmt.Errorf("Traffic mirror filter %s not found", rs.Primary.ID)
		}

		filter := out.TrafficMirrorFilters[0]
		var ruleList []*ec2.TrafficMirrorFilterRule
		ruleList = append(ruleList, filter.IngressFilterRules...)
		ruleList = append(ruleList, filter.EgressFilterRules...)

		var exists bool
		for _, rule := range ruleList {
			if aws.StringValue(rule.TrafficMirrorFilterRuleId) == ruleId {
				exists = true
				break
			}
		}

		if !exists {
			return fmt.Errorf("Rule %s not found inside filter %s", ruleId, filterId)
		}

		return nil
	}
}

func testAccTrafficMirrorFilterRuleConfig_basic(dstCidr, srcCidr, action, dir string, num int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
}

resource "aws_ec2_traffic_mirror_filter_rule" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  destination_cidr_block   = "%s"
  rule_action              = "%s"
  rule_number              = %d
  source_cidr_block        = "%s"
  traffic_direction        = "%s"
}
`, dstCidr, action, num, srcCidr, dir)
}

func testAccTrafficMirrorFilterRuleConfig_full(dstCidr, srcCidr, action, dir, description string, ruleNum, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {}

resource "aws_ec2_traffic_mirror_filter_rule" "test" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.test.id
  destination_cidr_block   = "%s"
  rule_action              = "%s"
  rule_number              = %d
  source_cidr_block        = "%s"
  traffic_direction        = "%s"
  description              = "%s"
  protocol                 = %d
  source_port_range {
    from_port = %d
    to_port   = %d
  }
  destination_port_range {
    from_port = %d
    to_port   = %d
  }
}
`, dstCidr, action, ruleNum, srcCidr, dir, description, protocol, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo)
}

func testAccPreCheckTrafficMirrorFilterRule(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	_, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter rule acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorFilterRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_filter_rule" {
			continue
		}

		ruleId := rs.Primary.ID
		filterId := rs.Primary.Attributes["traffic_mirror_filter_id"]

		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(filterId),
			},
		})

		if tfawserr.ErrCodeEquals(err, "InvalidTrafficMirrorFilterId.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return nil
		}

		filter := out.TrafficMirrorFilters[0]
		var ruleList []*ec2.TrafficMirrorFilterRule
		ruleList = append(ruleList, filter.IngressFilterRules...)
		ruleList = append(ruleList, filter.EgressFilterRules...)

		for _, rule := range ruleList {
			if aws.StringValue(rule.TrafficMirrorFilterRuleId) == ruleId {
				return fmt.Errorf("Rule %s still exists in filter %s", ruleId, filterId)
			}
		}
	}

	return nil
}

func testAccTrafficMirrorFilterRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["traffic_mirror_filter_id"], rs.Primary.ID), nil
	}
}
