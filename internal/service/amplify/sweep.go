//go:build sweep
// +build sweep

package amplify

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_amplify_app", &resource.Sweeper{
		Name: "aws_amplify_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AmplifyConn
	input := &amplify.ListAppsInput{}
	var sweeperErrs *multierror.Error

	err = listAppsPages(conn, input, func(page *amplify.ListAppsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, app := range page.Apps {
			r := ResourceApp()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppId))
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
		log.Printf("[WARN] Skipping Amplify Apps sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Amplify Apps: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
