// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoveryreadiness "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53RecoveryReadinessResourceSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_basic(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccRoute53RecoveryReadinessResourceSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_basic(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53recoveryreadiness.ResourceResourceSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessResourceSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_tags1(rName, cwArn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceSetConfig_tags2(rName, cwArn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccResourceSetConfig_tags1(rName, cwArn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessResourceSet_readinessScope(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_readinessScopes(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.readiness_scopes.#", acctest.Ct1),
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

func TestAccRoute53RecoveryReadinessResourceSet_basicDNSTargetResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	domainName := "myTestDomain.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()
	recordType := "A"
	recordSetId := "12345"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckResourceSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_basicDNSTarget(rName, domainName, hzArn, recordType, recordSetId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.domain_name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "resources.0.dns_target_resource.0.hosted_zone_arn"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.record_type", recordType),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.record_set_id", recordSetId),
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

func TestAccRoute53RecoveryReadinessResourceSet_dnsTargetResourceNLBTarget(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_dnsTargetNlbTarget(rName, hzArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "resources.0.dns_target_resource.0.target_resource.0.nlb_resource.0.arn"),
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

func TestAccRoute53RecoveryReadinessResourceSet_dnsTargetResourceR53Target(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()
	domainName := "my.target.domain"
	recordSetId := "987654321"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_dnsTargetR53Target(rName, hzArn, domainName, recordSetId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.target_resource.0.r53_resource.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.target_resource.0.r53_resource.0.record_set_id", recordSetId),
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

func TestAccRoute53RecoveryReadinessResourceSet_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53RecoveryReadinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig_timeout(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceSetExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "route53-recovery-readiness", regexache.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccCheckResourceSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53recoveryreadiness_resource_set" {
				continue
			}

			input := &route53recoveryreadiness.GetResourceSetInput{
				ResourceSetName: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetResourceSetWithContext(ctx, input)
			if err == nil {
				return fmt.Errorf("Route53RecoveryReadiness Resource Set (%s) not deleted", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckResourceSetExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

		input := &route53recoveryreadiness.GetResourceSetInput{
			ResourceSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetResourceSetWithContext(ctx, input)

		return err
	}
}

func testAccPreCheckResourceSet(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	input := &route53recoveryreadiness.ListResourceSetsInput{}

	_, err := conn.ListResourceSetsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccResourceSetConfig_NLB(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

data "aws_caller_identity" "current" {}
`, rName)
}

func testAccResourceSetConfig_basic(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }
}
`, rName, cwArn)
}

func testAccResourceSetConfig_tags1(rName, cwArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, cwArn, tagKey1, tagValue1)
}

func testAccResourceSetConfig_tags2(rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccResourceSetConfig_readinessScopes(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = "resource_set_test_cell"
}

resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn     = %[2]q
    readiness_scopes = [aws_route53recoveryreadiness_cell.test.arn]
  }
}
`, rName, cwArn)
}

func testAccResourceSetConfig_basicDNSTarget(rName, domainName, hzArn, recordType, recordSetId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = %[2]q
      hosted_zone_arn = %[3]q
      record_type     = %[4]q
      record_set_id   = %[5]q
    }
  }
}
`, rName, domainName, hzArn, recordType, recordSetId))
}

func testAccResourceSetConfig_dnsTargetNlbTarget(rName, hzArn string) string {
	return acctest.ConfigCompose(testAccResourceSetConfig_NLB(rName), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = "myTestDomain.test"
      hosted_zone_arn = %[2]q
      record_type     = "A"
      record_set_id   = "12345"

      target_resource {
        nlb_resource {
          arn = aws_lb.test.arn
        }
      }
    }
  }
}
`, rName, hzArn))
}

func testAccResourceSetConfig_dnsTargetR53Target(rName, hzArn, domainName, recordSetId string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = "myTestDomain.test"
      hosted_zone_arn = %[2]q
      record_type     = "A"
      record_set_id   = "12345"

      target_resource {
        r53_resource {
          domain_name   = %[3]q
          record_set_id = %[4]q
        }
      }
    }
  }
}
`, rName, hzArn, domainName, recordSetId)
}

func testAccResourceSetConfig_timeout(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  timeouts {
    delete = "10m"
  }
}
`, rName, cwArn)
}
