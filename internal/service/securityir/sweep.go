// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir

import (
	"context"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityir"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityir/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_securityir_membership", sweepMemberships)
}

func sweepMemberships(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SecurityIRClient(ctx)
	region := client.Region(ctx)
	var input securityir.ListMembershipsInput
	var sweepResources []sweep.Sweepable

	for item, err := range listMemberships(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if item.MembershipStatus != awstypes.MembershipStatusActive || string(item.Region) != region {
			continue
		}

		id := aws.ToString(item.MembershipId)
		membership, err := findMembershipByID(ctx, conn, id)
		if retry.NotFound(err) {
			continue
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		// Only cancel memberships created by acceptance tests.
		if !strings.HasPrefix(aws.ToString(membership.MembershipName), sweep.ResourcePrefix) {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newMembershipResource, client,
			framework.NewAttribute(names.AttrID, id),
		))
	}

	return sweepResources, nil
}
