// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIsCloudFrontDistributionARN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		arn      string
		expected bool
	}{
		{
			name:     "standard AWS partition",
			arn:      "arn:aws:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: true,
		},
		{
			name:     "AWS GovCloud partition",
			arn:      "arn:aws-us-gov:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: true,
		},
		{
			name:     "AWS China partition",
			arn:      "arn:aws-cn:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: true,
		},
		{
			name:     "ISOB partition",
			arn:      "arn:isob:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: true,
		},
		{
			name:     "unknown future partition",
			arn:      "arn:aws-new-region:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: true,
		},
		{
			name:     "ALB ARN",
			arn:      "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188", //lintignore:AWSAT003,AWSAT005
			expected: false,
		},
		{
			name:     "CloudFront origin access identity",
			arn:      "arn:aws:cloudfront::123456789012:origin-access-identity/E12345678901234", //lintignore:AWSAT005
			expected: false,
		},
		{
			name:     "not an ARN",
			arn:      "not-an-arn",
			expected: false,
		},
		{
			name:     "invalid ARN format",
			arn:      "arn:aws:cloudfront:123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.arn, func(t *testing.T) {
			t.Parallel()
			result := tfwafv2.IsCloudFrontDistributionARN(tt.arn)
			if result != tt.expected {
				t.Errorf("isCloudFrontDistributionARN(%q) = %v, want %v", tt.arn, result, tt.expected)
			}
		})
	}
}

func TestCloudFrontDistributionIDFromARN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		arn         string
		expectedID  string
		expectError bool
	}{
		{
			name:       "standard AWS CloudFront ARN",
			arn:        "arn:aws:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expectedID: "E12345678901234",
		},
		{
			name:       "GovCloud CloudFront ARN",
			arn:        "arn:aws-us-gov:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expectedID: "E12345678901234",
		},
		{
			name:       "China CloudFront ARN",
			arn:        "arn:aws-cn:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expectedID: "E12345678901234",
		},
		{
			name:       "ISOB CloudFront ARN",
			arn:        "arn:isob:cloudfront::123456789012:distribution/E12345678901234", //lintignore:AWSAT005
			expectedID: "E12345678901234",
		},
		{
			name:        "invalid ARN - no slash",
			arn:         "invalid-arn",
			expectError: true,
		},
		{
			name:        "invalid ARN - missing distribution ID",
			arn:         "arn:aws:cloudfront::123456789012:distribution", //lintignore:AWSAT005
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.arn, func(t *testing.T) {
			t.Parallel()
			id, err := tfwafv2.CloudFrontDistributionIDFromARN(tt.arn)
			if tt.expectError {
				if err == nil {
					t.Errorf("cloudFrontDistributionIDFromARN(%q) expected error, got nil", tt.arn)
				}
			} else {
				if err != nil {
					t.Errorf("cloudFrontDistributionIDFromARN(%q) unexpected error: %v", tt.arn, err)
				}
				if id != tt.expectedID {
					t.Errorf("cloudFrontDistributionIDFromARN(%q) = %q, want %q", tt.arn, id, tt.expectedID)
				}
			}
		})
	}
}

func TestAccWAFV2WebACLDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	datasourceName := "data.aws_wafv2_web_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebACLDataSourceConfig_nonExistent(name),
				ExpectError: regexache.MustCompile(`WAFv2 WebACL not found`),
			},
			{
				Config: testAccWebACLDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, datasourceName, names.AttrARN, "wafv2", regexache.MustCompile(fmt.Sprintf("regional/webacl/%v/.+$", name))),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrScope, resourceName, names.AttrScope),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLDataSource_resource(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	datasourceName := "data.aws_wafv2_web_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLDataSourceConfig_resource(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, datasourceName, names.AttrARN, "wafv2", regexache.MustCompile(fmt.Sprintf("regional/webacl/%v/.+$", name))),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrScope, resourceName, names.AttrScope),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLDataSource_resourceNotFound(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebACLDataSourceConfig_resourceNotFound(name),
				ExpectError: regexache.MustCompile(`WAFv2 WebACL not found for`),
			},
		},
	})
}

func TestAccWAFV2WebACLDataSource_cloudfront(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	datasourceName := "data.aws_wafv2_web_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckWAFV2CloudFrontScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLDataSourceConfig_cloudfront(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrScope, "CLOUDFRONT"),
				),
			},
		},
	})
}

func testAccWebACLDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = aws_wafv2_web_acl.test.name
  scope = "REGIONAL"
}
`, name)
}

func testAccWebACLDataSourceConfig_nonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = "tf-acc-test-does-not-exist"
  scope = "REGIONAL"
}
`, name)
}

func testAccWebACLDataSourceConfig_resource(name string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = false
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_association" "test" {
  resource_arn = aws_lb.test.arn
  web_acl_arn  = aws_wafv2_web_acl.test.arn
}

data "aws_wafv2_web_acl" "test" {
  resource_arn = aws_lb.test.arn
  scope        = "REGIONAL"

  depends_on = [aws_wafv2_web_acl_association.test]
}
`, name)
}

func testAccWebACLDataSourceConfig_resourceNotFound(name string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = false
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_wafv2_web_acl" "test" {
  resource_arn = aws_lb.test.arn
  scope        = "REGIONAL"
}
`, name)
}

func testAccWebACLDataSourceConfig_cloudfront(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "CLOUDFRONT"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_cloudfront_distribution" "test" {
  web_acl_id = aws_wafv2_web_acl.test.arn

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false
      cookies {
        forward = "all"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

data "aws_wafv2_web_acl" "test" {
  resource_arn = aws_cloudfront_distribution.test.arn
  scope        = "CLOUDFRONT"
}
`, name)
}
