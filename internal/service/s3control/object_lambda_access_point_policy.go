// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_object_lambda_access_point_policy")
func resourceObjectLambdaAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObjectLambdaAccessPointPolicyCreate,
		ReadWithoutTimeout:   resourceObjectLambdaAccessPointPolicyRead,
		UpdateWithoutTimeout: resourceObjectLambdaAccessPointPolicyUpdate,
		DeleteWithoutTimeout: resourceObjectLambdaAccessPointPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPolicy: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceObjectLambdaAccessPointPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	name := d.Get(names.AttrName).(string)
	id := ObjectLambdaAccessPointCreateResourceID(accountID, name)
	input := &s3control.PutAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyForObjectLambda(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Object Lambda Access Point (%s) Policy: %s", id, err)
	}

	d.SetId(id)

	return resourceObjectLambdaAccessPointPolicyRead(ctx, d, meta)
}

func resourceObjectLambdaAccessPointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, status, err := findObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Object Lambda Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("has_public_access_policy", status.IsPublic)
	d.Set(names.AttrName, name)

	if policy != "" {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}

	return nil
}

func resourceObjectLambdaAccessPointPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3control.PutAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicyForObjectLambda(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	return resourceObjectLambdaAccessPointPolicyRead(ctx, d, meta)
}

func resourceObjectLambdaAccessPointPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Object Lambda Access Point Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointPolicyForObjectLambda(ctx, &s3control.DeleteAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Object Lambda Access Point Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func findObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (string, *types.PolicyStatus, error) {
	inputGAPPFOL := &s3control.GetAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	outputGAPPFOL, err := conn.GetAccessPointPolicyForObjectLambda(ctx, inputGAPPFOL)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputGAPPFOL,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if outputGAPPFOL == nil {
		return "", nil, tfresource.NewEmptyResultError(inputGAPPFOL)
	}

	policy := aws.ToString(outputGAPPFOL.Policy)

	if policy == "" {
		return "", nil, tfresource.NewEmptyResultError(inputGAPPFOL)
	}

	inputGAPPSFOL := &s3control.GetAccessPointPolicyStatusForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	outputGAPPSFOL, err := conn.GetAccessPointPolicyStatusForObjectLambda(ctx, inputGAPPSFOL)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputGAPPSFOL,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if outputGAPPSFOL == nil || outputGAPPSFOL.PolicyStatus == nil {
		return "", nil, tfresource.NewEmptyResultError(inputGAPPSFOL)
	}

	return policy, outputGAPPSFOL.PolicyStatus, nil
}
