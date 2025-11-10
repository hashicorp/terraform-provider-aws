// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationscontacts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_notificationscontacts_email_contact", sweepEmailContacts)
}

func sweepEmailContacts(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NotificationsContactsClient(ctx)
	var input notificationscontacts.ListEmailContactsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := notificationscontacts.NewListEmailContactsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.EmailContacts {
			sweepResources = append(sweepResources, framework.NewSweepResource(newEmailContactResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
