// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package acmpca

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func init() {
	resource.AddTestSweepers("aws_acmpca_certificate_authority", &resource.Sweeper{
		Name: "aws_acmpca_certificate_authority",
		F:    sweepCertificateAuthorities,
	})
}

func sweepCertificateAuthorities(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ACMPCAConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &acmpca.ListCertificateAuthoritiesInput{}

	err = conn.ListCertificateAuthoritiesPagesWithContext(ctx, input, func(page *acmpca.ListCertificateAuthoritiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.CertificateAuthorities {
			if item == nil {
				continue
			}

			if aws.StringValue(item.Status) == acmpca.CertificateAuthorityStatusDeleted {
				continue
			}

			arn := aws.StringValue(item.Arn)

			r := ResourceCertificateAuthority()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("permanent_deletion_time_in_days", 7)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing ACM PCA Certificate Authorities: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping ACM PCA Certificate Authorities for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping ACM PCA Certificate Authorities sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
