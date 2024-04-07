// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"
	"encoding/json"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ce_anomaly_monitor", name="Anomaly Monitor")
// @Tags(identifierAttribute="id")
func ResourceAnomalyMonitor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnomalyMonitorCreate,
		ReadWithoutTimeout:   resourceAnomalyMonitorRead,
		UpdateWithoutTimeout: resourceAnomalyMonitorUpdate,
		DeleteWithoutTimeout: resourceAnomalyMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitor_dimension": {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"monitor_specification"},
				ValidateDiagFunc: enum.Validate[awstypes.MonitorDimension](),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexache.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Monitor Name matching expression: [\\S\\s]*")),
			},
			"monitor_specification": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				ConflictsWith:    []string{"monitor_dimension"},
			},
			"monitor_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MonitorType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalyMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	input := &costexplorer.CreateAnomalyMonitorInput{
		AnomalyMonitor: &awstypes.AnomalyMonitor{
			MonitorName: aws.String(d.Get("name").(string)),
			MonitorType: awstypes.MonitorType(d.Get("monitor_type").(string)),
		},
		ResourceTags: getTagsIn(ctx),
	}
	switch awstypes.MonitorType(d.Get("monitor_type").(string)) {
	case awstypes.MonitorTypeDimensional:
		if v, ok := d.GetOk("monitor_dimension"); ok {
			input.AnomalyMonitor.MonitorDimension = awstypes.MonitorDimension(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "If Monitor Type is %s, dimension attrribute is required", awstypes.MonitorTypeDimensional)
		}
	case awstypes.MonitorTypeCustom:
		if v, ok := d.GetOk("monitor_specification"); ok {
			expression := awstypes.Expression{}

			if err := json.Unmarshal([]byte(v.(string)), &expression); err != nil {
				return sdkdiag.AppendErrorf(diags, "parsing specification: %s", err)
			}

			input.AnomalyMonitor.MonitorSpecification = &expression
		} else {
			return sdkdiag.AppendErrorf(diags, "If Monitor Type is %s, dimension attrribute is required", awstypes.MonitorTypeCustom)
		}
	}

	resp, err := conn.CreateAnomalyMonitor(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Anomaly Monitor: %s", err)
	}

	if resp == nil || resp.MonitorArn == nil {
		return sdkdiag.AppendErrorf(diags, "creating Cost Explorer Anomaly Monitor resource (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.ToString(resp.MonitorArn))

	return append(diags, resourceAnomalyMonitorRead(ctx, d, meta)...)
}

func resourceAnomalyMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	monitor, err := FindAnomalyMonitorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CE, create.ErrActionReading, ResNameAnomalyMonitor, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionReading, ResNameAnomalyMonitor, d.Id(), err)
	}

	if monitor.MonitorSpecification != nil {
		specificationToJson, err := json.Marshal(monitor.MonitorSpecification)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing specification response: %s", err)
		}
		specificationToSet, err := structure.NormalizeJsonString(string(specificationToJson))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Specification (%s) is invalid JSON: %s", specificationToSet, err)
		}

		d.Set("monitor_specification", specificationToSet)
	}

	d.Set("arn", monitor.MonitorArn)
	d.Set("monitor_dimension", monitor.MonitorDimension)
	d.Set("name", monitor.MonitorName)
	d.Set("monitor_type", monitor.MonitorType)

	return diags
}

func resourceAnomalyMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)
	requestUpdate := false

	input := &costexplorer.UpdateAnomalyMonitorInput{
		MonitorArn: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.MonitorName = aws.String(d.Get("name").(string))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.UpdateAnomalyMonitor(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.CE, create.ErrActionUpdating, ResNameAnomalyMonitor, d.Id(), err)
		}
	}

	return append(diags, resourceAnomalyMonitorRead(ctx, d, meta)...)
}

func resourceAnomalyMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	_, err := conn.DeleteAnomalyMonitor(ctx, &costexplorer.DeleteAnomalyMonitorInput{MonitorArn: aws.String(d.Id())})

	if err != nil && errs.IsA[*awstypes.UnknownMonitorException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionDeleting, ResNameAnomalyMonitor, d.Id(), err)
	}

	return diags
}
