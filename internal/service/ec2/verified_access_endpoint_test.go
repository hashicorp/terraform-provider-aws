package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2VerifiedAccessEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_basic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					// resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					// resource.TestCheckResourceAttr(resourceName, "deletion_time", ""),
					// resource.TestCheckResourceAttr(resourceName, "description", ""),
					// resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					// acctest.CheckResourceAttrAccountID(resourceName, "owner"),
					// resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
					// resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					// resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_arn"),
					// resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_id"),
					// resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_instance_id"),
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

func TestAccEC2VerifiedAccessEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_endpoint" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Endpoint %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckVerifiedAccessEndpointExists(ctx context.Context, name string, verifiedaccessendpoint *types.VerifiedAccessEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*verifiedaccessendpoint = *output

		return nil
	}
}

func testAccVerifiedAccessEndpointConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`

	resource "aws_security_group" "test" {
		vpc_id = aws_vpc.test.id
		  
		tags = {
			Name = %[1]q
			}
	}

	resource "aws_network_interface" "test" {
	subnet_id = aws_subnet.test[0].id
	  
	tags = {
		Name = %[1]q
		}
	}

	resource "aws_lb" "test" {
		name               = %[1]q
		internal           = true
		load_balancer_type = "network"
		subnets            = aws_subnet.test[*].id
	  }

resource "aws_verifiedaccess_instance" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_trust_provider" "test" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "oidc"

  oidc_options {
    authorization_endpoint = "https://example.com/authorization_endpoint"
    client_id              = "s6BhdRkqt3"
    client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
    issuer                 = "https://example.com"
    scope                  = "test"
    token_endpoint         = "https://example.com/token_endpoint"
    user_info_endpoint     = "https://example.com/user_info_endpoint"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance_trust_provider_attachment" "test" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.test.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}

resource "aws_verifiedaccess_group" "test" {
	verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id
  }



`, rName))
}

func testAccVerifiedAccessEndpointConfig_basic(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_base(rName), fmt.Sprintf(`

	resource "aws_acm_certificate" "test" {
		certificate_body = "%[2]s"
		private_key      = "%[3]s"
	  
		tags = {
		  Name = %[1]q
		}
	  }

resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.example.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "load-balancer"
  load_balancer_options {
    load_balancer_arn = aws_lb.test.arn
    port              = 443
    protocol          = "https"
    subnet_ids        = [for subnet in aws_subnet.test : subnet.id]
  }
  security_group_ids       = [aws_security_group.example.id]
  verified_access_group_id = aws_verifiedaccess_group.example.id

  tags = {
    Name = %[1]q
}
`, rName, key, certificate))
}
