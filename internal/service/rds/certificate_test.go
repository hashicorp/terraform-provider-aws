// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSCertificate_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_rds_certificate.test"
}

func TestAccRDSCertificateConfig_Basic(certificate_identifier bool) string {
	return fmt.Sprintf(`
resource "aws_rds_certificate" "test" {
  certificate_identifier = %[1]t 
}
	`, certificate_identifier)
}
