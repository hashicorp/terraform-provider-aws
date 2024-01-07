// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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

	paginator := secretsmanager.NewListSecretsPaginator(conn, &secretsmanager.ListSecretsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("error retrieving Secrets Manager Secrets: %w", err)
		}

		if page != nil {
			if len(page.SecretList) == 0 {
				log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			}

			for _, secret := range page.SecretList {
				name := aws.StringValue(secret.Name)

				log.Printf("[INFO] Deleting Secrets Manager Secret Policy: %s", name)
				input := &secretsmanager.DeleteResourcePolicyInput{
					SecretId: aws.String(name),
				}

				_, err := conn.DeleteResourcePolicy(ctx, input)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
						continue
					}
					log.Printf("[ERROR] Failed to delete Secrets Manager Secret Policy (%s): %s", name, err)
				}
			}
		}
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

	paginator := secretsmanager.NewListSecretsPaginator(conn, &secretsmanager.ListSecretsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("error retrieving Secrets Manager Secrets: %w", err)
		}

		if page != nil {
			if len(page.SecretList) == 0 {
				log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			}

			for _, secret := range page.SecretList {
				name := aws.StringValue(secret.Name)

				log.Printf("[INFO] Deleting Secrets Manager Secret: %s", name)
				input := &secretsmanager.DeleteSecretInput{
					ForceDeleteWithoutRecovery: aws.Bool(true),
					SecretId:                   aws.String(name),
				}

				_, err := conn.DeleteSecret(ctx, input)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
						continue
					}
					log.Printf("[ERROR] Failed to delete Secrets Manager Secret (%s): %s", name, err)
				}
			}
		}
	}
	return nil
}
