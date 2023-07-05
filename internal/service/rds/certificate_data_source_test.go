// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSCertificateDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCertificatePreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "data.aws_rds_certificate.latest", "id"),
				),
			},
		},
	})
}

func TestAccRDSCertificateDataSource_latestValidTill(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCertificatePreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateDataSourceConfig_latestValidTill(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "rds", regexp.MustCompile(`cert:rds-ca-[0-9]{4}`)),
					resource.TestCheckResourceAttr(dataSourceName, "certificate_type", "CA"),
					resource.TestCheckResourceAttr(dataSourceName, "customer_override", "false"),
					resource.TestCheckNoResourceAttr(dataSourceName, "customer_override_valid_till"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`rds-ca-[0-9]{4}`)),
					resource.TestMatchResourceAttr(dataSourceName, "thumbprint", regexp.MustCompile(`[0-9a-f]+`)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_from", regexp.MustCompile(acctest.RFC3339RegexPattern)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_till", regexp.MustCompile(acctest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func testAccCertificatePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeCertificatesInput{}

	_, err := conn.DescribeCertificatesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCertificateDataSourceConfig_id() string {
	return `
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

data "aws_rds_certificate" "test" {
  id = data.aws_rds_certificate.latest.id
}
`
}

func testAccCertificateDataSourceConfig_latestValidTill() string {
	return `
data "aws_rds_certificate" "test" {
  latest_valid_till = true
}
`
}
