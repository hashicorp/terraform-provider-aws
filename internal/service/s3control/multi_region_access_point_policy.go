// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"account_id": {
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
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateS3MultiRegionAccessPointName,
						},
						"policy": {
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
	conn, err := ConnForMRAP(ctx, meta.(*conns.AWSClient))

	if err != nil {
		return diag.FromErr(err)
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	resourceID := MultiRegionAccessPointCreateResourceID(accountID, aws.StringValue(input.Details.Name))

	output, err := conn.PutMultiRegionAccessPointPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Multi-Region Access Point (%s) Policy: %s", resourceID, err)
	}

	d.SetId(resourceID)

	_, err = waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for S3 Multi-Region Access Point Policy (%s) create: %s", d.Id(), err)
	}

	return resourceMultiRegionAccessPointPolicyRead(ctx, d, meta)
}

func resourceMultiRegionAccessPointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := ConnForMRAP(ctx, meta.(*conns.AWSClient))

	if err != nil {
		return diag.FromErr(err)
	}

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	policyDocument, err := FindMultiRegionAccessPointPolicyDocumentByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Multi-Region Access Point Policy (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	if policyDocument != nil {
		var oldDetails map[string]interface{}
		if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			oldDetails = v.([]interface{})[0].(map[string]interface{})
		}

		if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointPolicyDocument(name, policyDocument, oldDetails)}); err != nil {
			return diag.Errorf("setting details: %s", err)
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

	return nil
}

func resourceMultiRegionAccessPointPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := ConnForMRAP(ctx, meta.(*conns.AWSClient))

	if err != nil {
		return diag.FromErr(err)
	}

	accountID, _, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.PutMultiRegionAccessPointPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Multi-Region Access Point Policy (%s): %s", d.Id(), err)
	}

	_, err = waitMultiRegionAccessPointRequestSucceeded(ctx, conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return diag.Errorf("waiting for S3 Multi-Region Access Point Policy (%s) update: %s", d.Id(), err)
	}

	return resourceMultiRegionAccessPointPolicyRead(ctx, d, meta)
}

func FindMultiRegionAccessPointPolicyDocumentByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID string, name string) (*s3control.MultiRegionAccessPointPolicyDocument, error) {
	input := &s3control.GetMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPointPolicyWithContext(ctx, input)

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

func expandPutMultiRegionAccessPointPolicyInput_(tfMap map[string]interface{}) *s3control.PutMultiRegionAccessPointPolicyInput_ {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PutMultiRegionAccessPointPolicyInput_{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["policy"].(string); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			policy = v
		}

		apiObject.Policy = aws.String(policy)
	}

	return apiObject
}

func flattenMultiRegionAccessPointPolicyDocument(name string, apiObject *s3control.MultiRegionAccessPointPolicyDocument, old map[string]interface{}) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["name"] = name

	if v := apiObject.Proposed; v != nil {
		if v := v.Policy; v != nil {
			policyToSet := aws.StringValue(v)
			if old != nil {
				if w, ok := old["policy"].(string); ok {
					var err error
					policyToSet, err = verify.PolicyToSet(w, aws.StringValue(v))

					if err != nil {
						policyToSet = aws.StringValue(v)
					}
				}
			}
			tfMap["policy"] = policyToSet
		}
	}

	return tfMap
}
