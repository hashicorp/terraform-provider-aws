package elbv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccELBV2LoadBalancerDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	dataSourceName2 := "data.aws_lb.alb_test_with_name"
	dataSourceName3 := "data.aws_lb.alb_test_with_tags"
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAcclbBasicDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancerDataSource_outpost(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAcclbOutpostDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.0.outpost_id", resourceName, "subnet_mapping.0.outpost_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancerDataSource_backwardsCompatibility(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName1 := "data.aws_alb.alb_test_with_arn"
	dataSourceName2 := "data.aws_alb.alb_test_with_name"
	dataSourceName3 := "data.aws_alb.alb_test_with_tags"
	resourceName := "aws_alb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAcclbBackwardsCompatibilityDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName1, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "access_logs.#", resourceName, "access_logs.#"),
				),
			},
		},
	})
}

func testAcclbBasicDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  desync_mitigation_mode = "defensive"

  tags = {
    Name   = %[1]q
    Config = "Basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = aws_vpc.test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}

data "aws_lb" "alb_test_with_arn" {
  arn = aws_lb.test.arn
}

data "aws_lb" "alb_test_with_name" {
  name = aws_lb.test.name
}

data "aws_lb" "alb_test_with_tags" {
  tags = aws_lb.test.tags
}
`, rName))
}

func testAcclbOutpostDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = [aws_subnet.test.id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name   = %[1]q
    Config = "Outposts"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}

data "aws_lb" "alb_test_with_arn" {
  arn = aws_lb.test.arn
}
`, rName)
}

func testAcclbBackwardsCompatibilityDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_alb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name   = %[1]q
    Config = "BackwardsCompatibility"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = aws_vpc.test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}

data "aws_alb" "alb_test_with_arn" {
  arn = aws_alb.test.arn
}

data "aws_alb" "alb_test_with_name" {
  name = aws_alb.test.name
}

data "aws_alb" "alb_test_with_tags" {
  tags = aws_alb.test.tags
}
`, rName))
}
