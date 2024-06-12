// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_bcmdataexports_export", &resource.Sweeper{
		Name: "aws_bcmdataexports_export",
		F:    sweepExports,
	})
}

func sweepExports(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.BCMDataExportsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &bcmdataexports.ListExportsInput{}

	pages := bcmdataexports.NewListExportsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping BCM Data Exports export sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving BCM Data Exports Export: %w", err)
		}

		for _, b := range page.Exports {
			id := aws.ToString(b.ExportArn)

			log.Printf("[INFO] Deleting AuditManager Assessment: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceExport, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Assessments for %s: %w", region, err)
	}

	return nil
}
