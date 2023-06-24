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

func TestAccTransferProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedProfile
	resourceName := "aws_transfer_as2_profile.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testProfile_basic(rName, certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "as2_id", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_id"),
					resource.TestCheckResourceAttr(resourceName, "profile_type", "LOCAL"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testProfile_updated(rName, certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "as2_id", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_id"),
					resource.TestCheckResourceAttr(resourceName, "profile_type", "LOCAL"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccTransferProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedProfile
	resourceName := "aws_transfer_as2_profile.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())
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
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testProfile_basic(rName, certificate, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testProfile_basic(rName string, certificate string, key string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate = %[2]q
  private_key = %[3]q
  usage       = "SIGNING"
}
resource "aws_transfer_as2_profile" "test" {
  as2_id          = %[1]q
  certificate_ids = [aws_transfer_as2_certificate.test.certificate_id]
  profile_type    = "LOCAL"
}
`, rName, certificate, key)
}

func testProfile_updated(rName string, certificate string, key string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate = %[2]q
  private_key = %[3]q
  usage       = "SIGNING"
}
resource "aws_transfer_as2_profile" "test" {
  as2_id          = %[1]q
  certificate_ids = [aws_transfer_as2_certificate.test.certificate_id]
  profile_type    = "LOCAL"
}
`, rName, certificate, key)
}

func testAccCheckProfileExists(ctx context.Context, n string, v *transfer.DescribedProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Profile ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindProfileByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_as2_profile" {
				continue
			}

			_, err := tftransfer.FindProfileByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AS2 Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
