// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_inspector_assessment_target", sweepAssesmentTargets)
}

func sweepAssesmentTargets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input inspector.ListAssessmentTargetsInput
	conn := client.InspectorClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := inspector.NewListAssessmentTargetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentTargetArns {
			r := resourceAssessmentTarget()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
