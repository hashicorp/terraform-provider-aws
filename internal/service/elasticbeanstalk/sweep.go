//go:build sweep
// +build sweep

package elasticbeanstalk

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_elastic_beanstalk_application", &resource.Sweeper{
		Name:         "aws_elastic_beanstalk_application",
		Dependencies: []string{"aws_elastic_beanstalk_environment"},
		F:            sweepApplications,
	})

	resource.AddTestSweepers("aws_elastic_beanstalk_environment", &resource.Sweeper{
		Name: "aws_elastic_beanstalk_environment",
		F:    sweepEnvironments,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ElasticBeanstalkConn()

	resp, err := conn.DescribeApplicationsWithContext(ctx, &elasticbeanstalk.DescribeApplicationsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Elastic Beanstalk Application sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving beanstalk application: %w", err)
	}

	if len(resp.Applications) == 0 {
		log.Print("[DEBUG] No aws beanstalk applications to sweep")
		return nil
	}

	var errors error
	for _, bsa := range resp.Applications {
		applicationName := aws.StringValue(bsa.ApplicationName)
		_, err := conn.DeleteApplicationWithContext(ctx, &elasticbeanstalk.DeleteApplicationInput{
			ApplicationName: bsa.ApplicationName,
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "InvalidConfiguration.NotFound") || tfawserr.ErrCodeEquals(err, "ValidationError") {
				log.Printf("[DEBUG] beanstalk application %q not found", applicationName)
				continue
			}

			errors = multierror.Append(fmt.Errorf("error deleting Elastic Beanstalk Application %q: %w", applicationName, err))
		}
	}

	return errors
}

func sweepEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElasticBeanstalkConn()

	resp, err := conn.DescribeEnvironmentsWithContext(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
		IncludeDeleted: aws.Bool(false),
	})

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Elastic Beanstalk Environment sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving beanstalk environment: %w", err)
	}

	if len(resp.Environments) == 0 {
		log.Print("[DEBUG] No aws beanstalk environments to sweep")
		return nil
	}

	var errors error
	for _, bse := range resp.Environments {
		environmentName := aws.StringValue(bse.EnvironmentName)
		environmentID := aws.StringValue(bse.EnvironmentId)
		log.Printf("Trying to terminate (%s) (%s)", environmentName, environmentID)

		err := DeleteEnvironment(ctx, conn, environmentID, 5*time.Minute, 10*time.Second) //nolint:gomnd
		if err != nil {
			errors = multierror.Append(fmt.Errorf("error deleting Elastic Beanstalk Environment %q: %w", environmentID, err))
		}

		log.Printf("> Terminated (%s) (%s)", environmentName, environmentID)
	}

	return errors
}
