//go:build sweep
// +build sweep

package eks

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_eks_addon", &resource.Sweeper{
		Name: "aws_eks_addon",
		F:    sweepAddon,
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

func sweepAddon(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	ctx := context.TODO()
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListAddonsInput{
				ClusterName: cluster,
			}

			err := conn.ListAddonsPagesWithContext(ctx, input, func(page *eks.ListAddonsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, addon := range page.Addons {
					r := ResourceAddon()
					d := r.Data(nil)
					d.SetId(AddonCreateResourceID(aws.StringValue(cluster), aws.StringValue(addon)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Add-Ons (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping EKS Add-Ons sweep for %s: %w", region, err))
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Add-Ons (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(cluster))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Clusters sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EKS Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EKS Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepFargateProfiles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListFargateProfilesInput{
				ClusterName: cluster,
			}

			err := conn.ListFargateProfilesPages(input, func(page *eks.ListFargateProfilesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, profile := range page.FargateProfileNames {
					r := ResourceFargateProfile()
					d := r.Data(nil)
					d.SetId(FargateProfileCreateResourceID(aws.StringValue(cluster), aws.StringValue(profile)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Fargate Profiles (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Fargate Profiles sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Fargate Profiles (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepIdentityProvidersConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	ctx := context.TODO()
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPagesWithContext(ctx, input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListIdentityProviderConfigsInput{
				ClusterName: cluster,
			}

			err := conn.ListIdentityProviderConfigsPagesWithContext(ctx, input, func(page *eks.ListIdentityProviderConfigsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, identityProviderConfig := range page.IdentityProviderConfigs {
					r := ResourceIdentityProviderConfig()
					d := r.Data(nil)
					d.SetId(IdentityProviderConfigCreateResourceID(aws.StringValue(cluster), aws.StringValue(identityProviderConfig.Name)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Identity Provider Configs (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping EKS Identity Provider Configs sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Identity Provider Configs (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNodeGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EKSConn
	input := &eks.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			input := &eks.ListNodegroupsInput{
				ClusterName: cluster,
			}

			err := conn.ListNodegroupsPages(input, func(page *eks.ListNodegroupsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, nodeGroup := range page.Nodegroups {
					r := ResourceNodeGroup()
					d := r.Data(nil)
					d.SetId(NodeGroupCreateResourceID(aws.StringValue(cluster), aws.StringValue(nodeGroup)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Node Groups (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Node Groups sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EKS Node Groups (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
