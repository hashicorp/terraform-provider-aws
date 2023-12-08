// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package cur

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	cur "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	resource.AddTestSweepers("aws_cur_report_definition", &resource.Sweeper{
		Name: "aws_cur_report_definition",
		F:    sweepReportDefinitions,
	})
}

func sweepReportDefinitions(region string) error {
	ctx := sweep.Context(region)
	if region != names.USEast1RegionID {
		log.Printf("[WARN] Skipping Cost And Usage Report Definition sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CURConn(ctx)
	input := &cur.DescribeReportDefinitionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReportDefinitionsPagesWithContext(ctx, input, func(page *cur.DescribeReportDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportDefinition := range page.ReportDefinitions {
			r := ResourceReportDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(reportDefinition.ReportName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Cost And Usage Report Definition sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing Cost And Usage Report Definitions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cost And Usage Report Definitions (%s): %w", region, err)
	}

	return nil
}
