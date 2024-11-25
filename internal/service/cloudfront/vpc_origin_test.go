package cloudfront_test

import (
	"context"
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontVPCOrigin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_vpc_origin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCOriginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCOriginConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCOriginExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func testAccCheckVPCOriginExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Not yet implemented
		return nil
	}
}

func testAccCheckVPCOriginDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Not yet implemented
		return nil
	}
}

// FIXME: This resource has the right parts and the wrong configuration.
func testAccVPCOriginConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_security_group" "allow_tls" {
  name        = "allow_tls"
  description = "Allow TLS inbound traffic and all outbound traffic"
  vpc_id      = aws_vpc.main.id
}

resource "aws_vpc_security_group_ingress_rule" "allow_tls_ipv4" {
  security_group_id = aws_security_group.allow_tls.id
  cidr_ipv4         = aws_vpc.main.cidr_block
  from_port         = 443
  ip_protocol       = "tcp"
  to_port           = 443
}

resource "aws_vpc_security_group_egress_rule" "allow_all_traffic_ipv4" {
  security_group_id = aws_security_group.allow_tls.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1" # semantically equivalent to all ports
}

resource "aws_lb" "this" {
  name               = "test-lb-tf"
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.allow_tls.id]
  subnets            = [aws_subnet.a.id, aws_subnet.b.id, aws_subnet.c.id]
}

resource "aws_cloudfront_vpc_origin" "this" {
  vpc_origin_endpoint_config {
    name = "test2"
    origin_arn = aws_lb.this.arn
    http_port = 8080
    https_port = 8443
    origin_protocol_policy = "https-only"
    origin_ssl_protocols {
      items = ["TLSv1.2"]
      quantity = 1
    }
  }
}
`, rName)
}
