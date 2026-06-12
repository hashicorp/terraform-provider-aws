// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_devopsagent_private_connection", sweepPrivateConnections)
}

func sweepPrivateConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DevOpsAgentClient(ctx)
	input := devopsagent.ListPrivateConnectionsInput{}
	var sweepResources []sweep.Sweepable

	out, err := conn.ListPrivateConnections(ctx, &input)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, nil
	}

	for _, v := range out.PrivateConnections {
		sweepResources = append(sweepResources, framework.NewSweepResource(newPrivateConnectionResource, client,
			framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
		)
	}

	return sweepResources, nil
}
