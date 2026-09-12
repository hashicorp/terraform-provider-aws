// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2ApplicationStatusCheck_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^asc-[0-9a-f]+$`)),
					resource.TestCheckResourceAttr(resourceName, "aggregation", "included"),
					resource.TestCheckResourceAttr(resourceName, "device_index", "0"),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "initialization_grace_period_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInterval, "60"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope", "private"),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "http"),
					resource.TestCheckResourceAttr(resourceName, "status_code_matcher", "200"),
					resource.TestCheckResourceAttr(resourceName, "success_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "6"),
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

func TestAccEC2ApplicationStatusCheck_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceApplicationStatusCheck, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheck_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckConfig_full("excluded", 1, 3, 120, "ipv6", "/ready", 443, "https", "200,204", 3, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", "excluded"),
					resource.TestCheckResourceAttr(resourceName, "device_index", "1"),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "initialization_grace_period_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/ready"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "https"),
					resource.TestCheckResourceAttr(resourceName, "status_code_matcher", "200,204"),
					resource.TestCheckResourceAttr(resourceName, "success_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "10"),
				),
			},
			{
				Config: testAccApplicationStatusCheckConfig_full("included", 0, 4, 240, "ipv4", "/status", 8080, "http", "200-299", 4, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", "included"),
					resource.TestCheckResourceAttr(resourceName, "device_index", "0"),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "initialization_grace_period_seconds", "240"),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "http"),
					resource.TestCheckResourceAttr(resourceName, "status_code_matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "success_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "15"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheck_healthCheckPath(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckConfig_healthCheckPath(names.AttrSource, names.AttrDestination),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_path.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.source.0.security_group_id", "aws_security_group.source", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.source.0.subnet_id", "aws_subnet.source", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "health_check_path.0.destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.destination.0.security_group_id", "aws_security_group.destination", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.destination.0.subnet_id", "aws_subnet.destination", names.AttrID),
				),
			},
			{
				Config: testAccApplicationStatusCheckConfig_healthCheckPath(names.AttrDestination, names.AttrSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.source.0.security_group_id", "aws_security_group.destination", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.source.0.subnet_id", "aws_subnet.destination", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.destination.0.security_group_id", "aws_security_group.source", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "health_check_path.0.destination.0.subnet_id", "aws_subnet.source", names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccApplicationStatusCheckConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_path.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccPreCheckApplicationStatusCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

	_, err := conn.DescribeApplicationStatusChecks(ctx, &ec2.DescribeApplicationStatusChecksInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckApplicationStatusCheckExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check", name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check", name, errors.New("ID not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindApplicationStatusCheckByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckApplicationStatusCheckDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_application_status_check" {
				continue
			}

			_, err := tfec2.FindApplicationStatusCheckByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Application Status Check", rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Application Status Check", rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccApplicationStatusCheckConfig_basic() string {
	return `
resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80
}
`
}

func testAccApplicationStatusCheckConfig_full(aggregation string, deviceIndex, failureThreshold, initializationGracePeriodSeconds int, ipVersion, path string, port int, protocol, statusCodeMatcher string, successThreshold, timeout int) string {
	return fmt.Sprintf(`
resource "aws_ec2_application_status_check" "test" {
  aggregation                         = %[1]q
  device_index                        = %[2]d
  failure_threshold                   = %[3]d
  initialization_grace_period_seconds = %[4]d
  interval                            = 60
  ip_scope                            = "private"
  ip_version                          = %[5]q
  path                                = %[6]q
  port                                = %[7]d
  protocol                            = %[8]q
  status_code_matcher                 = %[9]q
  success_threshold                   = %[10]d
  timeout                             = %[11]d
}
`, aggregation, deviceIndex, failureThreshold, initializationGracePeriodSeconds, ipVersion, path, port, protocol, statusCodeMatcher, successThreshold, timeout)
}

func testAccApplicationStatusCheckConfig_healthCheckPath(source, destination string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_subnet" "destination" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_security_group" "source" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "destination" {
  vpc_id = aws_vpc.test.id

  ingress {
    from_port       = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.source.id]
    to_port         = 443
  }
}

resource "aws_ec2_application_status_check" "test" {
  protocol = "https"
  port     = 443

  health_check_path {
    source {
      security_group_id = aws_security_group.%[1]s.id
      subnet_id         = aws_subnet.%[1]s.id
    }

    destination {
      security_group_id = aws_security_group.%[2]s.id
      subnet_id         = aws_subnet.%[2]s.id
    }
  }
}
`, source, destination)
}
