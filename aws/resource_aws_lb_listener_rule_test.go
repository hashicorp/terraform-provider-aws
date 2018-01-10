package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestLBListenerARNFromRuleARN(t *testing.T) {
	cases := []struct {
		name     string
		arn      string
		expected string
	}{
		{
			name:     "valid listener rule arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener-rule/app/name/0123456789abcdef/abcdef0123456789/456789abcedf1234",
			expected: "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789",
		},
		{
			name:     "listener arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789",
			expected: "",
		},
		{
			name:     "some other arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067",
			expected: "",
		},
		{
			name:     "not an arn",
			arn:      "blah blah blah",
			expected: "",
		},
		{
			name:     "empty arn",
			arn:      "",
			expected: "",
		},
	}

	for _, tc := range cases {
		actual := lbListenerARNFromRuleARN(tc.arn)
		if actual != tc.expected {
			t.Fatalf("incorrect arn returned: %q\nExpected: %s\n     Got: %s", tc.name, tc.expected, actual)
		}
	}
}

func TestAccAWSLBListenerRule_basic(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRuleBackwardsCompatibility(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_alb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_alb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_updateRulePriority(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "101"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_changeListenerRuleArnForcesNew(t *testing.T) {
	var before, after elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &before),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &after),
					testAccCheckAWSLbListenerRuleRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_multipleConditionThrowsError(t *testing.T) {
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBListenerRuleConfig_multipleConditions(lbName, targetGroupName),
				ExpectError: regexp.MustCompile(`attribute supports 1 item maximum`),
			},
		},
	})
}

func testAccCheckAWSLbListenerRuleRecreated(t *testing.T,
	before, after *elbv2.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleArn == *after.RuleArn {
			t.Fatalf("Expected change of Listener Rule ARNs, but both were %v", before.RuleArn)
		}
		return nil
	}
}

func testAccCheckAWSLBListenerRuleExists(n string, res *elbv2.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Listener Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		describe, err := conn.DescribeRules(&elbv2.DescribeRulesInput{
			RuleArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.Rules) != 1 ||
			*describe.Rules[0].RuleArn != rs.Primary.ID {
			return errors.New("Listener Rule not found")
		}

		*res = *describe.Rules[0]
		return nil
	}
}

func testAccCheckAWSLBListenerRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_listener_rule" && rs.Type != "aws_alb_listener_rule" {
			continue
		}

		describe, err := conn.DescribeRules(&elbv2.DescribeRulesInput{
			RuleArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.Rules) != 0 &&
				*describe.Rules[0].RuleArn == rs.Primary.ID {
				return fmt.Errorf("Listener Rule %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isRuleNotFound(err) {
			return nil
		} else {
			return errwrap.Wrapf("Unexpected error checking LB Listener Rule destroyed: {{err}}", err)
		}
	}

	return nil
}

func testAccAWSLBListenerRuleConfig_multipleConditions(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*", "static"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_alb_listener_rule" "static" {
  listener_arn = "${aws_alb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_alb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_alb_listener" "front_end" {
   load_balancer_arn = "${aws_alb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_alb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 101

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end_ruleupdate.arn}"
  priority = 101

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb_listener" "front_end_ruleupdate" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "8080"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}
