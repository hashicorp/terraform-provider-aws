package guardduty

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceDetector() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDetectorRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDetectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn()

	detectorId := d.Get("id").(string)

	if detectorId == "" {
		input := &guardduty.ListDetectorsInput{}

		resp, err := conn.ListDetectorsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing GuardDuty Detectors: %s", err)
		}

		if resp == nil || len(resp.DetectorIds) == 0 {
			return sdkdiag.AppendErrorf(diags, "no GuardDuty Detectors found")
		}
		if len(resp.DetectorIds) > 1 {
			return sdkdiag.AppendErrorf(diags, "multiple GuardDuty Detectors found; please use the `id` argument to look up a single detector")
		}

		detectorId = aws.StringValue(resp.DetectorIds[0])
	}

	getInput := &guardduty.GetDetectorInput{
		DetectorId: aws.String(detectorId),
	}

	getResp, err := conn.GetDetectorWithContext(ctx, getInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): %s", detectorId, err)
	}

	if getResp == nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): empty result", detectorId)
	}

	d.SetId(detectorId)
	d.Set("status", getResp.Status)
	d.Set("service_role_arn", getResp.ServiceRole)
	d.Set("finding_publishing_frequency", getResp.FindingPublishingFrequency)

	return diags
}
