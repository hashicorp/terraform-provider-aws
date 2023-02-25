//go:build sweep
// +build sweep

package amplify

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AmplifyConn()
	sweepResources := make([]sweep.Sweepable, 0)

	input := &amplify.ListAppsInput{}
	err = listAppsPages(ctx, conn, input, func(page *amplify.ListAppsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, app := range page.Apps {
			r := ResourceApp()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Amplify App sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing Amplify Apps: %w", err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Amplify Apps (%s): %w", region, err)
	}

	return nil
}
