//go:build sweep
// +build sweep

package appsync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_appsync_graphql_api", &resource.Sweeper{
		Name: "aws_appsync_graphql_api",
		F:    sweepGraphQLAPIs,
	})
}

func sweepGraphQLAPIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppSyncConn

	input := &appsync.ListGraphqlApisInput{}

	for {
		output, err := conn.ListGraphqlApis(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppSync GraphQL API sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving AppSync GraphQL APIs: %s", err)
		}

		for _, graphAPI := range output.GraphqlApis {
			id := aws.StringValue(graphAPI.ApiId)
			input := &appsync.DeleteGraphqlApiInput{
				ApiId: graphAPI.ApiId,
			}

			log.Printf("[INFO] Deleting AppSync GraphQL API %s", id)
			_, err := conn.DeleteGraphqlApi(input)

			if err != nil {
				return fmt.Errorf("error deleting AppSync GraphQL API (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}
