//go:build sweep
// +build sweep

package lambda

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_lambda_function", &resource.Sweeper{
		Name: "aws_lambda_function",
		F:    sweepFunctions,
	})

	resource.AddTestSweepers("aws_lambda_layer", &resource.Sweeper{
		Name: "aws_lambda_layer",
		F:    sweepLayerVersions,
	})
}

func sweepFunctions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).LambdaConn

	resp, err := conn.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Function sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Lambda functions: %s", err)
	}

	if len(resp.Functions) == 0 {
		log.Print("[DEBUG] No aws lambda functions to sweep")
		return nil
	}

	for _, f := range resp.Functions {
		_, err := conn.DeleteFunction(
			&lambda.DeleteFunctionInput{
				FunctionName: f.FunctionName,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func sweepLayerVersions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).LambdaConn
	resp, err := conn.ListLayers(&lambda.ListLayersInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Layer sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Lambda layers: %s", err)
	}

	if len(resp.Layers) == 0 {
		log.Print("[DEBUG] No aws lambda layers to sweep")
		return nil
	}

	for _, l := range resp.Layers {
		versionResp, err := conn.ListLayerVersions(&lambda.ListLayerVersionsInput{
			LayerName: l.LayerName,
		})
		if err != nil {
			return fmt.Errorf("Error retrieving versions for lambda layer: %s", err)
		}

		for _, v := range versionResp.LayerVersions {
			_, err := conn.DeleteLayerVersion(&lambda.DeleteLayerVersionInput{
				LayerName:     l.LayerName,
				VersionNumber: v.Version,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
