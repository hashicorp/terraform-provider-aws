// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_secretsmanager_secret_policy", &resource.Sweeper{
		Name: "aws_secretsmanager_secret_policy",
		F:    sweepSecretPolicies,
	})

	resource.AddTestSweepers("aws_secretsmanager_secret", &resource.Sweeper{
		Name: "aws_secretsmanager_secret",
		F:    sweepSecrets,
	})
}

func sweepSecretPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SecretsManagerClient(ctx)
	input := &secretsmanager.ListSecretsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := secretsmanager.NewListSecretsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Secrets Manager Secrets (%s): %w", region, err)
		}

		for _, v := range page.SecretList {
			arn := aws.ToString(v.ARN)

			if owningService := aws.ToString(v.OwningService); owningService != "" {
				log.Printf("[INFO] Skipping Secrets Manager Secret %s: OwningService=%s", arn, owningService)
				continue
			}

			r := resourceSecretPolicy()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Secrets Manager Secret Policies (%s): %w", region, err)
	}

	return nil
}

func sweepSecrets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SecretsManagerClient(ctx)
	input := &secretsmanager.ListSecretsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := secretsmanager.NewListSecretsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Secrets Manager Secrets (%s): %w", region, err)
		}

		for _, v := range page.SecretList {
			arn := aws.ToString(v.ARN)

			if owningService := aws.ToString(v.OwningService); owningService != "" {
				log.Printf("[INFO] Skipping Secrets Manager Secret %s: OwningService=%s", arn, owningService)
				continue
			}

			r := resourceSecret()
			d := r.Data(nil)
			d.SetId(arn)
			// Refresh replicas.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping Secrets Manager Secret %s: %s", arn, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Secrets Manager Secrets (%s): %w", region, err)
	}

	return nil
}
