package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransferAgreement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedAgreement
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	resourceName := "aws_transfer_agreement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAgreement_basic(rName, baseDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "agreement_id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_role"),
					resource.TestCheckResourceAttrSet(resourceName, "base_directory"),
					resource.TestCheckResourceAttrSet(resourceName, "local_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "partner_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "serverid"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
			{
				Config: testAgreement_updated(rName, baseDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "agreement_id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_role"),
					resource.TestCheckResourceAttrSet(resourceName, "base_directory"),
					resource.TestCheckResourceAttrSet(resourceName, "local_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "partner_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "serverid"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccTransferAgreement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedAgreement
	resourceName := "aws_transfer_agreement.test"
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAgreement_basic(rName, baseDirectory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceAgreement(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAgreement_basic(rName string, baseDirectory string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
  {
	"Version": "2012-10-17",
	"Statement": [{
	  "Effect": "Allow",
	  "Principal": {
		"Service": "transfer.amazonaws.com"
	  },
	  "Action": "sts:AssumeRole"
	}]
  }
EOF
}
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
	 "Version":"2012-10-17",
	 "Statement":[
		{
		   "Sid":"AllowFullAccesstoS3",
		   "Effect":"Allow",
		   "Action":[
			  "s3:*"
		   ],
		   "Resource":"*"
		}
	 ]
}
POLICY
}
resource "aws_transfer_profile" "local" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "LOCAL"
}
resource "aws_transfer_profile" "partner" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "PARTNER"
}
resource "aws_transfer_server" "test" {}
resource "aws_transfer_agreement" "test" {
  access_role        = aws_iam_role.test.arn
  base_directory     = %[2]q
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  serverid           = aws_transfer_server.test.id
}
`, rName, baseDirectory)
}

func testAgreement_updated(rName string, baseDirectory string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
		{
		  "Version": "2012-10-17",
		  "Statement": [{
			"Effect": "Allow",
			"Principal": {
			  "Service": "transfer.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }]
		}
	  EOF
}
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
   "Version":"2012-10-17",
   "Statement":[
	  {
		 "Sid":"AllowFullAccesstoS3",
		 "Effect":"Allow",
		 "Action":[
			"s3:*"
		 ],
		 "Resource":"*"
      }
	]
}
POLICY
}
resource "aws_transfer_profile" "local" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "LOCAL"
}
resource "aws_transfer_profile" "partner" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "PARTNER"
}
resource "aws_transfer_server" "test" {}
resource "aws_transfer_agreement" "test" {
  access_role        = aws_iam_role.test.arn
  base_directory     = %[2]q
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  serverid           = aws_transfer_server.test.id
}
`, rName, baseDirectory)
}

func testAccCheckAgreementExists(ctx context.Context, n string, v *transfer.DescribedAgreement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Agreement ID is set")
		}

		agreementID, serverID, err := tftransfer.AccessParseResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Transfer Agreement ID: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindAgreementByID(ctx, conn, agreementID, serverID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAgreementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_agreement" {
				continue
			}
			agreementID, serverID, err := tftransfer.AccessParseResourceID(rs.Primary.ID)

			if err != nil {
				return fmt.Errorf("error parsing Transfer Agreement ID: %w", err)
			}

			_, err = tftransfer.FindAgreementByID(ctx, conn, agreementID, serverID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AS2 Agreement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
