//go:build sweep
// +build sweep

package cur

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	cur "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cur_report_definition", &resource.Sweeper{
		Name: "aws_cur_report_definition",
		F:    sweepReportDefinitions,
	})
}

func sweepReportDefinitions(region string) error {
	c, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	client := c.(*conns.AWSClient)
	if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), sweep.Partition(region), cur.ServiceName); ok {
		_, ok := rs[region]
		if !ok {
			log.Printf("[WARN] Skipping Cost and Usage Report Definitions sweep for %s: not supported in this region", region)
			return nil
		}
	}

	conn := client.CURConn

	input := &cur.DescribeReportDefinitionsInput{}
	var sweeperErrs *multierror.Error
	err = conn.DescribeReportDefinitionsPages(input, func(page *cur.DescribeReportDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportDefinition := range page.ReportDefinitions {
			r := ResourceReportDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(reportDefinition.ReportName))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Cost And Usage Report Definitions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Cost And Usage Report Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
