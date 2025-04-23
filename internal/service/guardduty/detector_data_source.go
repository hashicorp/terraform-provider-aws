// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_guardduty_detector", name="Detector")
// @Tags
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(tagsIdentifierAttribute="arn")
func dataSourceDetector() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDetectorRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"features": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_configuration": {
							Computed: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrStatus: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrServiceRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceDetectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get(names.AttrID).(string)

	if detectorID == "" {
		output, err := FindDetector(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading this account's single GuardDuty Detector: %s", err)
		}

		detectorID = aws.ToString(output)
	}

	gdo, err := findDetectorByID(ctx, conn, detectorID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): %s", detectorID, err)
	}

	d.SetId(detectorID)
	if gdo.Features != nil {
		if err := d.Set("features", flattenDetectorFeatureConfigurationResults(gdo.Features)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting features: %s", err)
		}
	} else {
		d.Set("features", nil)
	}
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("detector/%s", detectorID),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("finding_publishing_frequency", gdo.FindingPublishingFrequency)
	d.Set(names.AttrServiceRoleARN, gdo.ServiceRole)
	d.Set(names.AttrStatus, gdo.Status)

	setTagsOut(ctx, gdo.Tags)

	return diags
}
