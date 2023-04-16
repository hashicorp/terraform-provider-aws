package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessTrustProviderAttachment_basic(t *testing.T) {
	ctx := context.Background()
	var verifiedaccesstrustprovider ec2.VerifiedAccessTrustProviderCondensed
	description := sdkacctest.RandString(100)
	resourceName := "aws_verifiedaccess_trust_provider_attachment.test_01"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil, // Cannot check for destroy
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderAttachmentConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderAttachmentExists(ctx, resourceName, &verifiedaccesstrustprovider),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccVerifiedAccessTrustProviderAttachment_disappears(t *testing.T) {
	ctx := context.Background()
	var verifiedaccesstrustprovider ec2.VerifiedAccessTrustProviderCondensed
	description := sdkacctest.RandString(100)
	resourceName := "aws_verifiedaccess_trust_provider_attachment.test_01"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil, // Cannot check for destroy
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderAttachmentConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderAttachmentExists(ctx, resourceName, &verifiedaccesstrustprovider),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessTrustProviderAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessTrustProviderAttachmentExists(ctx context.Context, name string, verifiedaccesstrustprovider *ec2.VerifiedAccessTrustProviderCondensed) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProviderAttachment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProviderAttachment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		resp, err := tfec2.FindVerifiedAccessTrustProviderAttachment(ctx, conn, rs.Primary.Attributes["verified_access_trust_provider_id"], rs.Primary.Attributes["verified_access_instance_id"])

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessTrustProviderAttachment, rs.Primary.ID, err)
		}

		*verifiedaccesstrustprovider = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()
	ctx := context.Background()

	input := &ec2.DescribeVerifiedAccessInstancesInput{}
	_, err := conn.DescribeVerifiedAccessInstancesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessTrustProviderAttachmentConfig_baseConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test_01" {
  description = %[1]q
}
resource "aws_verifiedaccess_trust_provider" "test_01" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "iam-identity-center"
}
`, description)
}

func testAccVerifiedAccessTrustProviderAttachmentConfig_basic(description string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessTrustProviderAttachmentConfig_baseConfig(description), `
resource "aws_verifiedaccess_trust_provider_attachment" "test_01" {
  verified_access_instance_id       = aws_verifiedaccess_instance.test_01.id
  verified_access_trust_provider_id = aws_verifiedaccess_trust_provider.test_01.id
}
`)
}
