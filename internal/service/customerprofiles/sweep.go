// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package customerprofiles

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_customerprofiles_domain", sweepDomains)
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input customerprofiles.ListDomainsInput
	conn := client.CustomerProfilesClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err := listDomainsPages(ctx, conn, &input, func(page *customerprofiles.ListDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := resourceDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
