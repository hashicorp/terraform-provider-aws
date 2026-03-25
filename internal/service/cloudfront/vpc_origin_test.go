// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontVPCOrigin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcOrigin awstypes.VpcOrigin
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_vpc_origin.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCOriginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "vpcorigin/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceName, names.AttrID, aws.ToString(vpcOrigin.Id))(s)
					},
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_origin_endpoint_config.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.http_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.https_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_protocol_policy", "http-only"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.quantity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccCloudFrontVPCOrigin_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcOrigin awstypes.VpcOrigin
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_vpc_origin.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCOriginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceVPCOrigin, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontVPCOrigin_update(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcOrigin awstypes.VpcOrigin
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_vpc_origin.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCOriginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "vpcorigin/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_origin_endpoint_config.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.http_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.https_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_protocol_policy", "http-only"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.quantity", "1"),
				),
			},
			{
				Config: testAccVPCOriginConfig_httpsOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "cloudfront", "vpcorigin/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_origin_endpoint_config.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.http_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.https_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_protocol_policy", "https-only"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.items.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_origin_endpoint_config.0.origin_ssl_protocols.0.quantity", "2"),
				),
			},
		},
	})
}

func TestAccCloudFrontVPCOrigin_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcOrigin awstypes.VpcOrigin
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_vpc_origin.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCOriginConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccVPCOriginConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCOriginConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, t, resourceName, &vpcOrigin),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckVPCOriginExists(ctx context.Context, t *testing.T, n string, v *awstypes.VpcOrigin) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindVPCOriginByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.VpcOrigin

		return nil
	}
}

func testAccCheckVPCOriginDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_vpc_origin" {
				continue
			}

			_, err := tfcloudfront.FindVPCOriginByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront VPC Origin %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccVPCOriginConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName))
}

func testAccVPCOriginConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCOriginConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_vpc_origin" "test" {
  vpc_origin_endpoint_config {
    name                   = %[1]q
    arn                    = aws_lb.test.arn
    http_port              = 8080
    https_port             = 8443
    origin_protocol_policy = "http-only"

    origin_ssl_protocols {
      items    = ["TLSv1.2"]
      quantity = 1
    }
  }
}
`, rName))
}

func testAccVPCOriginConfig_httpsOnly(rName string) string {
	return acctest.ConfigCompose(testAccVPCOriginConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_vpc_origin" "test" {
  vpc_origin_endpoint_config {
    name                   = %[1]q
    arn                    = aws_lb.test.arn
    http_port              = 8080
    https_port             = 8443
    origin_protocol_policy = "https-only"

    origin_ssl_protocols {
      items    = ["TLSv1.2", "TLSv1.1"]
      quantity = 2
    }
  }
}
`, rName))
}

func testAccVPCOriginConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCOriginConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_vpc_origin" "test" {
  vpc_origin_endpoint_config {
    name                   = %[1]q
    arn                    = aws_lb.test.arn
    http_port              = 8080
    https_port             = 8443
    origin_protocol_policy = "http-only"

    origin_ssl_protocols {
      items    = ["TLSv1.2"]
      quantity = 1
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccVPCOriginConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCOriginConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudfront_vpc_origin" "test" {
  vpc_origin_endpoint_config {
    name                   = %[1]q
    arn                    = aws_lb.test.arn
    http_port              = 8080
    https_port             = 8443
    origin_protocol_policy = "http-only"

    origin_ssl_protocols {
      items    = ["TLSv1.2"]
      quantity = 1
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
