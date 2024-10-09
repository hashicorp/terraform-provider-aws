// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_multi_region_access_point_policy")
func resourceMultiRegionAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMultiRegionAccessPointPolicyCreate,
		ReadWithoutTimeout:   resourceMultiRegionAccessPointPolicyRead,
		UpdateWithoutTimeout: resourceMultiRegionAccessPointPolicyUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"details": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateS3MultiRegionAccessPointName,
						},
						names.AttrPolicy: {
							Type:                  schema.TypeString,
							Required:              true,
							ValidateFunc:          validation.StringIsJSON,
							DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
							DiffSuppressOnRefresh: true,
							StateFunc: func(v interface{}) string {
								json, _ := structure.NormalizeJsonString(v)
								return json
							},
						},
					},
				},
			},
			"established": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"proposed": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMultiRegionAccessPointPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	id := MultiRegionAccessPointCreateResourceID(accountID, aws.ToString(input.Details.Name))

	output, err := conn.PutMultiRegionAccessPointPolicy(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Multi-Region Access Point (%s) Policy: %s", id, err)
	}

	d.SetId(id)

	if _, err := waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.ToString(output.RequestTokenARN), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Multi-Region Access Point Policy (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMultiRegionAccessPointPolicyRead(ctx, d, meta)...)
}

func resourceMultiRegionAccessPointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := findMultiRegionAccessPointPolicyDocumentByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Multi-Region Access Point Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	if policyDocument != nil {
		var oldDetails map[string]interface{}
		if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			oldDetails = v.([]interface{})[0].(map[string]interface{})
		}

		if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointPolicyDocument(name, policyDocument, oldDetails)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting details: %s", err)
		}
	} else {
		d.Set("details", nil)
	}
	if v := policyDocument.Established; v != nil {
		d.Set("established", v.Policy)
	} else {
		d.Set("established", nil)
	}
	if v := policyDocument.Proposed; v != nil {
		d.Set("proposed", v.Policy)
	} else {
		d.Set("proposed", nil)
	}

	return diags
}

func resourceMultiRegionAccessPointPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, _, err := MultiRegionAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.PutMultiRegionAccessPointPolicy(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Multi-Region Access Point Policy (%s): %s", d.Id(), err)
	}

	if _, err := waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.ToString(output.RequestTokenARN), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Multi-Region Access Point Policy (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceMultiRegionAccessPointPolicyRead(ctx, d, meta)...)
}

func findMultiRegionAccessPointPolicyDocumentByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*types.MultiRegionAccessPointPolicyDocument, error) {
	input := &s3control.GetMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPointPolicy(ctx, input, func(o *s3control.Options) {
		// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
		o.Region = names.USWest2RegionID
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}

func expandPutMultiRegionAccessPointPolicyInput_(tfMap map[string]interface{}) *types.PutMultiRegionAccessPointPolicyInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PutMultiRegionAccessPointPolicyInput{}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPolicy].(string); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			policy = v
		}

		apiObject.Policy = aws.String(policy)
	}

	return apiObject
}

func flattenMultiRegionAccessPointPolicyDocument(name string, apiObject *types.MultiRegionAccessPointPolicyDocument, old map[string]interface{}) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrName] = name

	if v := apiObject.Proposed; v != nil {
		if v := v.Policy; v != nil {
			policyToSet := aws.ToString(v)
			if old != nil {
				if w, ok := old[names.AttrPolicy].(string); ok {
					var err error
					policyToSet, err = verify.PolicyToSet(w, aws.ToString(v))

					if err != nil {
						policyToSet = aws.ToString(v)
					}
				}
			}
			tfMap[names.AttrPolicy] = policyToSet
		}
	}

	return tfMap
}
