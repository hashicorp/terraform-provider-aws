// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package secretsmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.SecretsManagerConn(ctx)

	err = conn.ListSecretsPagesWithContext(ctx, &secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)

			log.Printf("[INFO] Deleting Secrets Manager Secret Policy: %s", name)
			input := &secretsmanager.DeleteResourcePolicyInput{
				SecretId: aws.String(name),
			}

			_, err := conn.DeleteResourcePolicyWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret Policy (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %w", err)
	}
	return nil
}

func sweepSecrets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SecretsManagerConn(ctx)

	err = conn.ListSecretsPagesWithContext(ctx, &secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)

			log.Printf("[INFO] Deleting Secrets Manager Secret: %s", name)
			input := &secretsmanager.DeleteSecretInput{
				ForceDeleteWithoutRecovery: aws.Bool(true),
				SecretId:                   aws.String(name),
			}

			_, err := conn.DeleteSecretWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %s", err)
	}
	return nil
}
