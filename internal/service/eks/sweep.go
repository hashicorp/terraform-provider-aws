// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EKSClient(ctx)
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Add-On sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
			break
		}

		for _, v := range page.Clusters {
			clusterName := v
			input := &eks.ListAddonsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListAddonsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*types.ResourceNotFoundException](err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Add-Ons (%s): %w", region, err))
					break
				}

				for _, v := range page.Addons {
					r := resourceAddon()
					d := r.Data(nil)
					d.SetId(AddonCreateResourceID(clusterName, v))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Add-Ons (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EKSClient(ctx)
	input := &eks.ListClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EKS Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EKS Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepFargateProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EKSClient(ctx)
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Fargate Profile sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
			break
		}

		for _, v := range page.Clusters {
			clusterName := v
			input := &eks.ListFargateProfilesInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListFargateProfilesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

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
					break
				}

				for _, v := range page.FargateProfileNames {
					r := resourceFargateProfile()
					d := r.Data(nil)
					d.SetId(FargateProfileCreateResourceID(clusterName, v))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EKSClient(ctx)
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Identity Provider Config sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
			break
		}

		for _, v := range page.Clusters {
			clusterName := v
			input := &eks.ListIdentityProviderConfigsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListIdentityProviderConfigsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

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
					break
				}

				for _, v := range page.IdentityProviderConfigs {
					r := resourceIdentityProviderConfig()
					d := r.Data(nil)
					d.SetId(IdentityProviderConfigCreateResourceID(clusterName, aws.ToString(v.Name)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Identity Provider Configs (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNodeGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EKSClient(ctx)
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EKS Node Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
			break
		}

		for _, v := range page.Clusters {
			clusterName := v
			input := &eks.ListNodegroupsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListNodegroupsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

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
					break
				}

				for _, v := range page.Nodegroups {
					r := resourceNodeGroup()
					d := r.Data(nil)
					d.SetId(NodeGroupCreateResourceID(clusterName, v))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Node Groups (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
