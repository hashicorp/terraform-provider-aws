// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_eks_addon", &resource.Sweeper{
		Name: "aws_eks_addon",
		F:    sweepAddons,
	})

	resource.AddTestSweepers("aws_eks_cluster", &resource.Sweeper{
		Name: "aws_eks_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_eks_addon",
			"aws_eks_fargate_profile",
			"aws_eks_node_group",
			"aws_emrcontainers_virtual_cluster",
		},
	})

	resource.AddTestSweepers("aws_eks_fargate_profile", &resource.Sweeper{
		Name: "aws_eks_fargate_profile",
		F:    sweepFargateProfiles,
	})

	resource.AddTestSweepers("aws_eks_identity_provider_config", &resource.Sweeper{
		Name: "aws_eks_identity_provider_config",
		F:    sweepIdentityProvidersConfig,
	})

	resource.AddTestSweepers("aws_eks_node_group", &resource.Sweeper{
		Name: "aws_eks_node_group",
		F:    sweepNodeGroups,
	})
}

func sweepAddons(region string) error {
	ctx := sweep.Context(region)
	sweepClient, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	client := sweepClient.EKSClient(ctx)

	paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Print(fmt.Errorf("[WARN] Skipping EKS Add-Ons sweep for %s: %w", region, err))
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListAddonsInput{
				ClusterName: &cluster,
			}

			paginator := eks.NewListAddonsPaginator(client, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*types.ResourceNotFoundException](err) {
					log.Print(fmt.Errorf("[WARN] Skipping cluster %s not found: %w", region, err))
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Add-Ons (%s): %w", region, err))
				}

				for _, addon := range page.Addons {
					r := ResourceAddon()
					d := r.Data(nil)
					d.SetId(AddonCreateResourceID(cluster, addon))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, sweepClient))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Add-Ons (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	sweepClient, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	sweepResources := make([]sweep.Sweepable, 0)

	client := sweepClient.EKSClient(ctx)

	paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Clusters sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EKS Clusters (%s): %w", region, err)
		}

		for _, cluster := range page.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(cluster)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, sweepClient))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping EKS Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepFargateProfiles(region string) error {
	ctx := sweep.Context(region)
	sweepClient, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	client := sweepClient.EKSClient(ctx)

	paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Fargate Profiles sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListFargateProfilesInput{
				ClusterName: &cluster,
			}

			paginator := eks.NewListFargateProfilesPaginator(client, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*types.ResourceNotFoundException](err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Fargate Profiles (%s): %w", region, err))
				}

				for _, profile := range page.FargateProfileNames {
					r := ResourceFargateProfile()
					d := r.Data(nil)
					d.SetId(FargateProfileCreateResourceID(cluster, profile))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, sweepClient))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Fargate Profiles (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepIdentityProvidersConfig(region string) error {
	ctx := sweep.Context(region)
	sweepClient, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	client := sweepClient.EKSClient(ctx)

	paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Print(fmt.Errorf("[WARN] Skipping EKS Identity Provider Configs sweep for %s: %w", region, err))
			return sweeperErrs // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListIdentityProviderConfigsInput{
				ClusterName: &cluster,
			}

			paginator := eks.NewListIdentityProviderConfigsPaginator(client, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*types.ResourceNotFoundException](err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Identity Provider Configs (%s): %w", region, err))
				}

				for _, identityProviderConfig := range page.IdentityProviderConfigs {
					r := ResourceIdentityProviderConfig()
					d := r.Data(nil)
					d.SetId(IdentityProviderConfigCreateResourceID(cluster, aws.ToString(identityProviderConfig.Name)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, sweepClient))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Identity Provider Configs (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNodeGroups(region string) error {
	ctx := sweep.Context(region)
	sweepClient, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	client := sweepClient.EKSClient(ctx)

	paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Node Groups sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListNodegroupsInput{
				ClusterName: &cluster,
			}

			paginator := eks.NewListNodegroupsPaginator(client, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}
				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*types.ResourceNotFoundException](err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Node Groups (%s): %w", region, err))
				}

				for _, nodeGroup := range page.Nodegroups {
					r := ResourceNodeGroup()
					d := r.Data(nil)
					d.SetId(NodeGroupCreateResourceID(cluster, nodeGroup))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, sweepClient))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Node Groups (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
