//go:build sweep
// +build sweep

package cloud9

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloud9_environment_ec2", &resource.Sweeper{
		Name: "aws_cloud9_environment_ec2",
		F:    sweepEnvironmentEC2s,
	})
}

func sweepEnvironmentEC2s(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Cloud9Conn()
	input := &cloud9.ListEnvironmentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListEnvironmentsPagesWithContext(ctx, input, func(page *cloud9.ListEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EnvironmentIds {
			r := ResourceEnvironmentEC2()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Cloud9 EC2 Environment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Cloud9 EC2 Environments (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cloud9 EC2 Environments (%s): %w", region, err)
	}

	return nil
}
