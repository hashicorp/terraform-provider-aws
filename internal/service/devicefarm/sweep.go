//go:build sweep
// +build sweep

package devicefarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_devicefarm_project", &resource.Sweeper{
		Name: "aws_devicefarm_project",
		F:    sweepProjects,
	})
}

func sweepProjects(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DeviceFarm

	input := &devicefarm.ListProjectsInput{}
	for {
		output, err := conn.ListProjectsPages(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DeviceFarm Project sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DeviceFarm Projects: %s", err)
		}

		if len(output.Projects) == 0 {
			log.Print("[DEBUG] No DeviceFarm Projects to sweep")
			return nil
		}

		for _, project := range output.Projects {
			arn := aws.StringValue(project.Arn)

			log.Printf("[INFO] Deleting DeviceFarm Project: %s", arn)
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(arn)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DeviceFarm Project (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}
