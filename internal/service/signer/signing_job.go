// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package signer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_signer_signing_job", name="Signing Job")
func resourceSigningJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningJobCreate,
		ReadWithoutTimeout:   resourceSigningJobRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"completed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestination: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucket: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"ignore_signing_job_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_invoker": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"profile_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"requested_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucket: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrVersion: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"signature_expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signed_object": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucket: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrKey: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSigningJobCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	profileName := d.Get("profile_name")
	input := signer.StartSigningJobInput{
		Destination: expandDestination(d.Get(names.AttrDestination).([]any)),
		ProfileName: aws.String(profileName.(string)),
		Source:      expandSource(d.Get(names.AttrSource).([]any)),
	}

	output, err := conn.StartSigningJob(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Signer Signing Job (%s): %s", profileName, err)
	}

	d.SetId(aws.ToString(output.JobId))

	const (
		timeout = 5 * time.Minute
	)
	_, err = waitSigningJobSucceeded(ctx, conn, d.Id(), timeout)

	if err != nil && !d.Get("ignore_signing_job_failure").(bool) {
		return sdkdiag.AppendErrorf(diags, "waiting for Signer Signing Job (%s) success: %s", d.Id(), err)
	}

	return append(diags, resourceSigningJobRead(ctx, d, meta)...)
}

func resourceSigningJobRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	output, err := findSigningJobByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Signer Signing Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer Signing Job (%s): %s", d.Id(), err)
	}

	d.Set("completed_at", aws.ToTime(output.CompletedAt).Format(time.RFC3339))
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).Format(time.RFC3339))
	d.Set("job_id", output.JobId)
	d.Set("job_invoker", output.JobInvoker)
	d.Set("job_owner", output.JobOwner)
	d.Set("platform_display_name", output.PlatformDisplayName)
	d.Set("platform_id", output.PlatformId)
	d.Set("profile_name", output.ProfileName)
	d.Set("profile_version", output.ProfileVersion)
	d.Set("requested_by", output.RequestedBy)
	if err := d.Set("revocation_record", flattenSigningJobRevocationRecord(output.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_record: %s", err)
	}
	signatureExpiresAt := ""
	if output.SignatureExpiresAt != nil {
		signatureExpiresAt = aws.ToTime(output.SignatureExpiresAt).Format(time.RFC3339)
	}
	d.Set("signature_expires_at", signatureExpiresAt)
	if err := d.Set("signed_object", flattenSignedObject(output.SignedObject)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signed_object: %s", err)
	}
	if err := d.Set(names.AttrSource, flattenSource(output.Source)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrStatusReason, output.StatusReason)

	return diags
}

func findSigningJobByID(ctx context.Context, conn *signer.Client, id string) (*signer.DescribeSigningJobOutput, error) {
	input := signer.DescribeSigningJobInput{
		JobId: aws.String(id),
	}

	output, err := conn.DescribeSigningJob(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusSigningJob(conn *signer.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSigningJobByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitSigningJobSucceeded(ctx context.Context, conn *signer.Client, id string, timeout time.Duration) (*signer.DescribeSigningJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.SigningStatusInProgress),
		Target:  enum.Slice(types.SigningStatusSucceeded),
		Refresh: statusSigningJob(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*signer.DescribeSigningJobOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func flattenSigningJobRevocationRecord(apiObject *types.SigningJobRevocationRecord) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if v := apiObject.Reason; v != nil {
		tfMap["reason"] = aws.ToString(v)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func flattenSource(apiObject *types.Source) []any {
	if apiObject == nil || apiObject.S3 == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"s3": flattenS3Source(apiObject.S3),
	}

	return []any{tfMap}
}

func flattenS3Source(apiObject *types.S3Source) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return []any{tfMap}
}

func expandSource(tfList []any) *types.Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	var apiObject *types.Source

	if v, ok := tfMap["s3"].([]any); ok && len(v) > 0 {
		apiObject = &types.Source{
			S3: expandS3Source(v),
		}
	}

	return apiObject
}

func expandS3Source(tfList []any) *types.S3Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.S3Source{}

	if v, ok := tfMap[names.AttrBucket].(string); ok {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandDestination(tfList []any) *types.Destination {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	var apiObject *types.Destination

	if v, ok := tfMap["s3"].([]any); ok && len(v) > 0 {
		apiObject = &types.Destination{
			S3: expandS3Destination(v),
		}
	}

	return apiObject
}

func expandS3Destination(tfList []any) *types.S3Destination {
	if tfList == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.S3Destination{}

	if _, ok := tfMap[names.AttrBucket]; ok {
		apiObject.BucketName = aws.String(tfMap[names.AttrBucket].(string))
	}

	if _, ok := tfMap[names.AttrPrefix]; ok {
		apiObject.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
	}

	return apiObject
}

func flattenSignedObject(apiObject *types.SignedObject) []any {
	if apiObject == nil || apiObject.S3 == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"s3": flattenS3SignedObject(apiObject.S3),
	}

	return []any{tfMap}
}

func flattenS3SignedObject(apiObject *types.S3SignedObject) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	return []any{tfMap}
}
