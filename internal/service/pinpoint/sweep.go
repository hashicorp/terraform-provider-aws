//go:build sweep
// +build sweep

package pinpoint

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_pinpoint_app", &resource.Sweeper{
		Name: "aws_pinpoint_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).PinpointConn()

	input := &pinpoint.GetAppsInput{}

	for {
		output, err := conn.GetAppsWithContext(ctx, input)
		if err != nil {
			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Pinpoint app sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Pinpoint apps: %s", err)
		}

		if len(output.ApplicationsResponse.Item) == 0 {
			log.Print("[DEBUG] No Pinpoint apps to sweep")
			return nil
		}

		for _, item := range output.ApplicationsResponse.Item {
			name := aws.StringValue(item.Name)

			log.Printf("[INFO] Deleting Pinpoint app %s", name)
			_, err := conn.DeleteAppWithContext(ctx, &pinpoint.DeleteAppInput{
				ApplicationId: item.Id,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Pinpoint app %s: %s", name, err)
			}
		}

		if output.ApplicationsResponse.NextToken == nil {
			break
		}
		input.Token = output.ApplicationsResponse.NextToken
	}

	return nil
}
