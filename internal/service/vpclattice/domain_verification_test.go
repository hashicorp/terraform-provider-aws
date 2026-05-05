// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeDomainVerification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domainVerification vpclattice.GetDomainVerificationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := fmt.Sprintf("%s.example.com", rName)
	resourceName := "aws_vpclattice_domain_verification.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainVerificationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainVerificationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainVerificationExists(ctx, t, resourceName, &domainVerification),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`domainverification/dv-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "txt_record_name"),
					resource.TestCheckResourceAttrSet(resourceName, "txt_record_value"),
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

func TestAccVPCLatticeDomainVerification_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domainVerification vpclattice.GetDomainVerificationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := fmt.Sprintf("%s.example.com", rName)
	resourceName := "aws_vpclattice_domain_verification.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainVerificationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainVerificationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainVerificationExists(ctx, t, resourceName, &domainVerification),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfvpclattice.ResourceDomainVerification, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainVerificationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_domain_verification" {
				continue
			}

			_, err := tfvpclattice.FindDomainVerificationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPCLattice Domain Verification %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainVerificationExists(ctx context.Context, t *testing.T, n string, v *vpclattice.GetDomainVerificationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindDomainVerificationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDomainVerificationConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_domain_verification" "test" {
  domain_name = %[1]q
}
`, domainName)
}
