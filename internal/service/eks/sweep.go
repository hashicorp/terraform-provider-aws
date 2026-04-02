// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_eks_addon", sweepAddons)
	awsv2.Register("aws_eks_cluster", sweepClusters,
		"aws_eks_addon",
		"aws_eks_fargate_profile",
		"aws_eks_node_group",
		"aws_emrcontainers_virtual_cluster",
		"aws_prometheus_scraper",
	)
	awsv2.Register("aws_eks_fargate_profile", sweepFargateProfiles)
	awsv2.Register("aws_eks_identity_provider_config", sweepIdentityProvidersConfig)
	awsv2.Register("aws_eks_node_group", sweepNodeGroups)
}

func sweepAddons(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EKSClient(ctx)
	var input eks.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, clusterName := range page.Clusters {
			input := eks.ListAddonsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListAddonsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					break
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					break
				}

				if err != nil {
					return nil, err
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

	return sweepResources, nil
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EKSClient(ctx)
	var input eks.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			const (
				timeout = 15 * time.Minute
			)
			err := updateClusterDeletionProtection(ctx, conn, v, false, timeout)

			// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
			// ¯\_(ツ)_/¯
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFargateProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EKSClient(ctx)
	var input eks.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, clusterName := range page.Clusters {
			input := eks.ListFargateProfilesInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListFargateProfilesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					break
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					break
				}

				if err != nil {
					return nil, err
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

	return sweepResources, nil
}

func sweepIdentityProvidersConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EKSClient(ctx)
	var input eks.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, clusterName := range page.Clusters {
			input := eks.ListIdentityProviderConfigsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListIdentityProviderConfigsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					break
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					break
				}

				if err != nil {
					return nil, err
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

	return sweepResources, nil
}

func sweepNodeGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EKSClient(ctx)
	var input eks.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := eks.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, clusterName := range page.Clusters {
			input := eks.ListNodegroupsInput{
				ClusterName: aws.String(clusterName),
			}

			pages := eks.NewListNodegroupsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					break
				}

				// There are EKS clusters that are listed (and are in the AWS Console) but can't be found.
				// ¯\_(ツ)_/¯
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					break
				}

				if err != nil {
					return nil, err
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

	return sweepResources, nil
}
