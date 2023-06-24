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

func TestAccConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedConnector
	resourceName := "aws_transfer_as2_connector.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testConnector_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "access_role"),
					resource.TestCheckResourceAttr(resourceName, "as2_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.compression"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.encryption_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.local_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.mdn_response"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.partner_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testConnector_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "access_role"),
					resource.TestCheckResourceAttr(resourceName, "as2_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.compression"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.encryption_algorithm"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.local_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.mdn_response"),
					resource.TestCheckResourceAttrSet(resourceName, "as2_config.0.partner_profile_id"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedConnector
	resourceName := "aws_transfer_as2_connector.test"
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
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testConnector_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testConnector_basic(rName string) string {
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

resource "aws_transfer_as2_profile" "local" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "LOCAL"
}
resource "aws_transfer_as2_profile" "partner" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "PARTNER"
}

resource "aws_transfer_as2_connector" "test" {
  access_role = aws_iam_role.test.arn
  as2_config {
	compression           = "DISABLED"
	encryption_algorithm  = "AES128_CBC"
    message_subject       = %[1]q
    local_profile_id      = aws_transfer_as2_profile.local.profile_id
	mdn_response          = "NONE"
	mdn_signing_algorithm = "NONE"
    partner_profile_id    = aws_transfer_as2_profile.partner.profile_id
	signing_algorithm     = "NONE"
  }
  url = "http://www.test.com"
}
`, rName)
}

func testConnector_updated(rName string) string {
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

resource "aws_transfer_as2_profile" "local" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "LOCAL"
}
resource "aws_transfer_as2_profile" "partner" {
  as2_id = %[1]q
  #certificate_ids = ["xxx"]
  profile_type = "PARTNER"
}

resource "aws_transfer_as2_connector" "test" {
  access_role = aws_iam_role.test.arn
  as2_config {
	compression           = "DISABLED"
	encryption_algorithm  = "AES128_CBC"
    message_subject       = %[1]q
    local_profile_id      = aws_transfer_as2_profile.local.profile_id
	mdn_response          = "NONE"
	mdn_signing_algorithm = "NONE"
    partner_profile_id    = aws_transfer_as2_profile.partner.profile_id
	signing_algorithm     = "NONE"
  }
  url = "http://www.test.com"
}
`, rName)
}

func testAccCheckConnectorExists(ctx context.Context, n string, v *transfer.DescribedConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn()

		output, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_as2_connector" {
				continue
			}

			_, err := tftransfer.FindConnectorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AS2 Connector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
