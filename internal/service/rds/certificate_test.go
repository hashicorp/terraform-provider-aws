// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	rds_sdkv1 "github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSCertificate_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_Basic("rds-ca-rsa4096-g1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificate(ctx, resourceName, "rds-ca-rsa4096-g1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_identifier", "rds-ca-rsa4096-g1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_Basic("rds-ca-rsa4096-g1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificate(ctx, resourceName, "rds-ca-rsa4096-g1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_identifier", "rds-ca-rsa4096-g1"),
				),
			},
		},
	})
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		response, err := conn.DescribeCertificatesWithContext(ctx, &rds_sdkv1.DescribeCertificatesInput{})
		if err != nil {
			return err
		}

		for _, c := range (*response).Certificates {
			if aws.BoolValue(c.CustomerOverride) == true {
				return fmt.Errorf("RDS certificate customer override is not removed on resource destruction")
			}
		}

		return nil
	}
}

func testAccCheckCertificate(ctx context.Context, n string, certificate_identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		response, err := conn.DescribeCertificatesWithContext(ctx, &rds_sdkv1.DescribeCertificatesInput{})
		if err != nil {
			return err
		}

		if aws.StringValue((*response).DefaultCertificateForNewLaunches) != certificate_identifier {
			return fmt.Errorf("RDS certificate override is not in the expected state (%s)", certificate_identifier)
		}

		return nil
	}
}

func testAccCertificateConfig_Basic(certificate_identifier string) string {
	return fmt.Sprintf(`
resource "aws_rds_certificate" "test" {
  certificate_identifier = "%s"
}
	`, certificate_identifier)
}
