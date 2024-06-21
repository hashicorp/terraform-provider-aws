// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_account_vdm_attributes", name="Account VDM Attributes")
func ResourceAccountVDMAttributes() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountVDMAttributesUpdate,
		ReadWithoutTimeout:   resourceAccountVDMAttributesRead,
		UpdateWithoutTimeout: resourceAccountVDMAttributesUpdate,
		DeleteWithoutTimeout: resourceAccountVDMAttributesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"dashboard_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"engagement_metrics": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.FeatureStatus](),
						},
					},
				},
			},
			"guardian_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"optimized_shared_delivery": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.FeatureStatus](),
						},
					},
				},
			},
			"vdm_enabled": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.FeatureStatus](),
			},
		},
	}
}

const (
	ResNameAccountVDMAttributes = "Account VDM Attributes"
)

func resourceAccountVDMAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.PutAccountVdmAttributesInput{
		VdmAttributes: &types.VdmAttributes{
			VdmEnabled: types.FeatureStatus(d.Get("vdm_enabled").(string)),
		},
	}

	if v, ok := d.GetOk("dashboard_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.VdmAttributes.DashboardAttributes = expandDashboardAttributes(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("guardian_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.VdmAttributes.GuardianAttributes = expandGuardianAttributes(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.PutAccountVdmAttributes(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameAccountVDMAttributes, "", err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameAccountVDMAttributes, "", errors.New("empty output"))
	}

	if d.IsNewResource() {
		d.SetId("ses-account-vdm-attributes")
	}

	return append(diags, resourceAccountVDMAttributesRead(ctx, d, meta)...)
}

func resourceAccountVDMAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindAccountVDMAttributes(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 AccountVDMAttributes (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameAccountVDMAttributes, d.Id(), err)
	}

	if out.DashboardAttributes != nil {
		if err := d.Set("dashboard_attributes", []interface{}{flattenDashboardAttributes(out.DashboardAttributes)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameAccountVDMAttributes, d.Id(), err)
		}
	}

	if out.GuardianAttributes != nil {
		if err := d.Set("guardian_attributes", []interface{}{flattenGuardianAttributes(out.GuardianAttributes)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameAccountVDMAttributes, d.Id(), err)
		}
	}

	d.Set("vdm_enabled", out.VdmEnabled)

	return diags
}

func resourceAccountVDMAttributesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 AccountVDMAttributes %s", d.Id())

	_, err := conn.PutAccountVdmAttributes(ctx, &sesv2.PutAccountVdmAttributesInput{
		VdmAttributes: &types.VdmAttributes{
			VdmEnabled: "DISABLED",
		},
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameAccountVDMAttributes, d.Id(), err)
	}

	return diags
}

func FindAccountVDMAttributes(ctx context.Context, conn *sesv2.Client) (*types.VdmAttributes, error) {
	in := &sesv2.GetAccountInput{}
	out, err := conn.GetAccount(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.VdmAttributes == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.VdmAttributes, nil
}

func expandDashboardAttributes(tfMap map[string]interface{}) *types.DashboardAttributes {
	if tfMap == nil {
		return nil
	}

	a := &types.DashboardAttributes{}

	if v, ok := tfMap["engagement_metrics"].(string); ok && v != "" {
		a.EngagementMetrics = types.FeatureStatus(v)
	}

	return a
}

func expandGuardianAttributes(tfMap map[string]interface{}) *types.GuardianAttributes {
	if tfMap == nil {
		return nil
	}

	a := &types.GuardianAttributes{}

	if v, ok := tfMap["optimized_shared_delivery"].(string); ok && v != "" {
		a.OptimizedSharedDelivery = types.FeatureStatus(v)
	}

	return a
}

func flattenDashboardAttributes(apiObject *types.DashboardAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"engagement_metrics": string(apiObject.EngagementMetrics),
	}

	return m
}

func flattenGuardianAttributes(apiObject *types.GuardianAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"optimized_shared_delivery": string(apiObject.OptimizedSharedDelivery),
	}

	return m
}
