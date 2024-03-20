// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
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
	input := &acmpca.ListCertificateAuthoritiesInput{}
	conn := client.ACMPCAConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListCertificateAuthoritiesPagesWithContext(ctx, input, func(page *acmpca.ListCertificateAuthoritiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CertificateAuthorities {
			arn := aws.StringValue(v.Arn)

			if status := aws.StringValue(v.Status); status == acmpca.CertificateAuthorityStatusDeleted {
				log.Printf("[INFO] Skipping ACM PCA Certificate Authority %s: Status=%s", arn, status)
				continue
			}

			r := ResourceCertificateAuthority()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("permanent_deletion_time_in_days", 7) //nolint:gomnd

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ACM PCA Certificate Authority sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ACM PCA Certificate Authorities (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ACM PCA Certificate Authorities (%s): %w", region, err)
	}

	return nil
}
