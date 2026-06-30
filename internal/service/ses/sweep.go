// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_ses_configuration_set", sweepConfigurationSets)
	awsv2.Register("aws_ses_domain_identity", sweepIdentities(awstypes.IdentityTypeDomain))
	awsv2.Register("aws_ses_email_identity", sweepIdentities(awstypes.IdentityTypeEmailAddress))
	awsv2.Register("aws_ses_receipt_rule_set", sweepReceiptRuleSets)
}

func sweepConfigurationSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SESClient(ctx)
	var input ses.ListConfigurationSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listConfigurationSetsPages(ctx, conn, &input, func(page *ses.ListConfigurationSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ConfigurationSets {
			r := resourceConfigurationSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepIdentities(identityType awstypes.IdentityType) sweep.SweeperFn {
	return func(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
		conn := client.SESClient(ctx)
		input := ses.ListIdentitiesInput{
			IdentityType: identityType,
		}
		sweepResources := make([]sweep.Sweepable, 0)

		pages := ses.NewListIdentitiesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return nil, err
			}

			for _, v := range page.Identities {
				var r *schema.Resource
				switch identityType {
				case awstypes.IdentityTypeDomain:
					r = resourceDomainIdentity()
				case awstypes.IdentityTypeEmailAddress:
					r = resourceEmailIdentity()
				default:
					continue
				}
				d := r.Data(nil)
				d.SetId(v)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}

		return sweepResources, nil
	}
}

func sweepReceiptRuleSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SESClient(ctx)
	var input ses.ListReceiptRuleSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	// You cannot delete the receipt rule set that is currently active.
	r := resourceActiveReceiptRuleSet()
	d := r.Data(nil)
	d.SetId("unknown")
	err := sweep.NewSweepResource(r, d, client).Delete(ctx)
	if err != nil {
		return nil, err
	}

	err = listReceiptRuleSetsPages(ctx, conn, &input, func(page *ses.ListReceiptRuleSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleSets {
			r := resourceReceiptRuleSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
