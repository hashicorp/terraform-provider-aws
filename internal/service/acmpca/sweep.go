// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.ACMPCAClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := acmpca.NewListCertificateAuthoritiesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ACM PCA Certificate Authority sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ACM PCA Certificate Authorities (%s): %w", region, err)
		}

		for _, v := range page.CertificateAuthorities {
			arn := aws.ToString(v.Arn)

			if v.Status == awstypes.CertificateAuthorityStatusDeleted {
				log.Printf("[INFO] Skipping ACM PCA Certificate Authority %s: Status=%s", arn, string(v.Status))
				continue
			}

			r := resourceCertificateAuthority()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("permanent_deletion_time_in_days", 7) //nolint:mnd // 7 days is the default value

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ACM PCA Certificate Authorities (%s): %w", region, err)
	}

	return nil
}
