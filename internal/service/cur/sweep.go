// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	cur "github.com/aws/aws-sdk-go-v2/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
	conn := client.CURClient(ctx)
	input := &cur.DescribeReportDefinitionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cur.NewDescribeReportDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cost And Usage Report Definition sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Cost And Usage Report Definitions (%s): %w", region, err)
		}

		for _, v := range page.ReportDefinitions {
			r := resourceReportDefinition()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ReportName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cost And Usage Report Definitions (%s): %w", region, err)
	}

	return nil
}
