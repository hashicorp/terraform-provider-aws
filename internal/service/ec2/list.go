// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// DescribeInstances is an "All-Or-Some" call.
func listInstances(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) iter.Seq2[[]awstypes.Instance, error] {
	return func(yield func([]awstypes.Instance, error) bool) {
		pages := ec2.NewDescribeInstancesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing EC2 Instances: %w", err))
				return
			}

			for _, v := range page.Reservations {
				if !yield(v.Instances, nil) {
					return
				}
			}
		}
	}
}

func listRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.RouteTable, error] {
	return func(yield func(awstypes.RouteTable, error) bool) {
		pages := ec2.NewDescribeRouteTablesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.RouteTable](), fmt.Errorf("listing EC2 Route Tables: %w", err))
				return
			}

			for _, v := range page.RouteTables {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listSecondaryNetworks(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecondaryNetworksInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.SecondaryNetwork, error] {
	return func(yield func(awstypes.SecondaryNetwork, error) bool) {
		pages := ec2.NewDescribeSecondaryNetworksPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.SecondaryNetwork](), fmt.Errorf("listing EC2 Secondary Networks: %w", err))
				return
			}

			for _, v := range page.SecondaryNetworks {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listSecondarySubnets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecondarySubnetsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.SecondarySubnet, error] {
	return func(yield func(awstypes.SecondarySubnet, error) bool) {
		pages := ec2.NewDescribeSecondarySubnetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.SecondarySubnet](), fmt.Errorf("listing EC2 Secondary Subnets: %w", err))
				return
			}

			for _, v := range page.SecondarySubnets {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listSecurityGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.SecurityGroup, error] {
	return func(yield func(awstypes.SecurityGroup, error) bool) {
		pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.SecurityGroup](), fmt.Errorf("listing EC2 Security Groups: %w", err))
				return
			}

			for _, v := range page.SecurityGroups {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listSecurityGroupRules(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput, filter tfslices.Predicate[awstypes.SecurityGroupRule], optFns ...func(*ec2.Options)) iter.Seq2[awstypes.SecurityGroupRule, error] {
	return func(yield func(awstypes.SecurityGroupRule, error) bool) {
		pages := ec2.NewDescribeSecurityGroupRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.SecurityGroupRule](), fmt.Errorf("listing EC2 Security Group Rules: %w", err))
				return
			}

			for _, v := range page.SecurityGroupRules {
				if filter != nil && !filter(v) {
					continue
				}

				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listSubnets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.Subnet, error] {
	return func(yield func(awstypes.Subnet, error) bool) {
		pages := ec2.NewDescribeSubnetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.Subnet](), fmt.Errorf("listing EC2 Subnets: %w", err))
				return
			}

			for _, v := range page.Subnets {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listTransitGatewayMeteringPolicies(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayMeteringPoliciesInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.TransitGatewayMeteringPolicy, error] {
	return func(yield func(awstypes.TransitGatewayMeteringPolicy, error) bool) {
		err := describeTransitGatewayMeteringPoliciesPages(ctx, conn, input, func(page *ec2.DescribeTransitGatewayMeteringPoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, v := range page.TransitGatewayMeteringPolicies {
				if !yield(v, nil) {
					return false
				}
			}

			return !lastPage
		}, optFns...)
		if err != nil {
			yield(inttypes.Zero[awstypes.TransitGatewayMeteringPolicy](), fmt.Errorf("listing EC2 Transit Gateway Metering Policies: %w", err))
			return
		}
	}
}

func listVPCs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.Vpc, error] {
	return func(yield func(awstypes.Vpc, error) bool) {
		pages := ec2.NewDescribeVpcsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.Vpc](), fmt.Errorf("listing EC2 VPCs: %w", err))
				return
			}

			for _, v := range page.Vpcs {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listVPCEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.VpcEndpoint, error] {
	return func(yield func(awstypes.VpcEndpoint, error) bool) {
		pages := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(inttypes.Zero[awstypes.VpcEndpoint](), fmt.Errorf("listing EC2 VPC Endpoints: %w", err))
				return
			}

			for _, item := range page.VpcEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
