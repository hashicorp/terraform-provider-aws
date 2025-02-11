// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_verifiedpermissions_policy_store", &resource.Sweeper{
		Name: "aws_verifiedpermissions_policy_store",
		F:    sweepPolicyStores,
	})
}

func sweepPolicyStores(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.VerifiedPermissionsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &verifiedpermissions.ListPolicyStoresInput{}

	pages := verifiedpermissions.NewListPolicyStoresPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping VerifiedPermissions Policy Stores sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving VerifiedPermissions Policy Stores: %w", err)
		}

		for _, store := range page.PolicyStores {
			id := aws.ToString(store.PolicyStoreId)
			log.Printf("[INFO] Deleting VerifiedPermissions Policy Store: %s", id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourcePolicyStore, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping VerifiedPermissions Policy Stores for %s: %w", region, err)
	}

	return nil
}
