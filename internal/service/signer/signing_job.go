// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_signer_signing_job")
func ResourceSigningJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningJobCreate,
		ReadWithoutTimeout:   resourceSigningJobRead,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"completed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceSigningJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)
	profileName := d.Get("profile_name")
	source := d.Get(names.AttrSource).([]interface{})
	destination := d.Get(names.AttrDestination).([]interface{})

	startSigningJobInput := &signer.StartSigningJobInput{
		ProfileName: aws.String(profileName.(string)),
		Source:      expandSigningJobSource(source),
		Destination: expandSigningJobDestination(destination),
	}

	log.Printf("[DEBUG] Starting Signer Signing Job using profile name %q.", profileName)
	startSigningJobOutput, err := conn.StartSigningJob(ctx, startSigningJobInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Signing Job: %s", err)
	}

	jobId := aws.ToString(startSigningJobOutput.JobId)

	ignoreSigningJobFailure := d.Get("ignore_signing_job_failure").(bool)
	log.Printf("[DEBUG] Waiting for Signer Signing Job ID (%s) to complete.", jobId)

	waitInput := &signer.DescribeSigningJobInput{
		JobId: aws.String(jobId),
	}
	waiter := signer.NewSuccessfulSigningJobWaiter(conn)
	waitTime := 5 * time.Minute
	err = waiter.Wait(ctx, waitInput, waitTime)

	if err != nil {
		var rnr *types.ResourceNotFoundException
		if !errors.As(err, &rnr) || !ignoreSigningJobFailure {
			return sdkdiag.AppendErrorf(diags, "creating Signing Job: waiting for completion: %s", err)
		}
	}

	d.SetId(jobId)

	return append(diags, resourceSigningJobRead(ctx, d, meta)...)
}

func resourceSigningJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)
	jobId := d.Id()

	describeSigningJobOutput, err := findSigningJobByID(ctx, conn, jobId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Signer Signing Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer signing job (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_id", describeSigningJobOutput.JobId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job id: %s", err)
	}

	if err := d.Set("completed_at", aws.ToTime(describeSigningJobOutput.CompletedAt).Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job completed at: %s", err)
	}

	if err := d.Set(names.AttrCreatedAt, aws.ToTime(describeSigningJobOutput.CreatedAt).Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job created at: %s", err)
	}

	if err := d.Set("job_invoker", describeSigningJobOutput.JobInvoker); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job invoker: %s", err)
	}

	if err := d.Set("job_owner", describeSigningJobOutput.JobOwner); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job owner: %s", err)
	}

	if err := d.Set("platform_display_name", describeSigningJobOutput.PlatformDisplayName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job platform display name: %s", err)
	}

	if err := d.Set("platform_id", describeSigningJobOutput.PlatformId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job platform id: %s", err)
	}

	if err := d.Set("profile_name", describeSigningJobOutput.ProfileName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job profile name: %s", err)
	}

	if err := d.Set("profile_version", describeSigningJobOutput.ProfileVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job profile version: %s", err)
	}

	if err := d.Set("requested_by", describeSigningJobOutput.RequestedBy); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job requested by: %s", err)
	}

	if err := d.Set("revocation_record", flattenSigningJobRevocationRecord(describeSigningJobOutput.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job revocation record: %s", err)
	}

	signatureExpiresAt := ""
	if describeSigningJobOutput.SignatureExpiresAt != nil {
		signatureExpiresAt = aws.ToTime(describeSigningJobOutput.SignatureExpiresAt).Format(time.RFC3339)
	}
	if err := d.Set("signature_expires_at", signatureExpiresAt); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job requested by: %s", err)
	}

	if err := d.Set("signed_object", flattenSigningJobSignedObject(describeSigningJobOutput.SignedObject)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job signed object: %s", err)
	}

	if err := d.Set(names.AttrSource, flattenSigningJobSource(describeSigningJobOutput.Source)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job source: %s", err)
	}

	if err := d.Set(names.AttrStatus, describeSigningJobOutput.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job status: %s", err)
	}

	if err := d.Set(names.AttrStatusReason, describeSigningJobOutput.StatusReason); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing job status reason: %s", err)
	}

	return diags
}

func flattenSigningJobRevocationRecord(apiObject *types.SigningJobRevocationRecord) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Reason; v != nil {
		tfMap["reason"] = aws.ToString(v)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenSigningJobSource(apiObject *types.Source) []interface{} {
	if apiObject == nil || apiObject.S3 == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3": flattenSigningJobS3Source(apiObject.S3),
	}

	return []interface{}{tfMap}
}

func flattenSigningJobS3Source(apiObject *types.S3Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandSigningJobSource(tfList []interface{}) *types.Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var source *types.Source
	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		source = &types.Source{
			S3: expandSigningJobS3Source(v),
		}
	}

	return source
}

func expandSigningJobS3Source(tfList []interface{}) *types.S3Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	s3Source := &types.S3Source{}

	if v, ok := tfMap[names.AttrBucket].(string); ok {
		s3Source.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		s3Source.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok {
		s3Source.Version = aws.String(v)
	}

	return s3Source
}

func expandSigningJobDestination(tfList []interface{}) *types.Destination {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var destination *types.Destination
	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		destination = &types.Destination{
			S3: expandSigningJobS3Destination(v),
		}
	}

	return destination
}

func expandSigningJobS3Destination(tfList []interface{}) *types.S3Destination {
	if tfList == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	s3Destination := &types.S3Destination{}

	if _, ok := tfMap[names.AttrBucket]; ok {
		s3Destination.BucketName = aws.String(tfMap[names.AttrBucket].(string))
	}

	if _, ok := tfMap[names.AttrPrefix]; ok {
		s3Destination.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
	}

	return s3Destination
}

func flattenSigningJobSignedObject(apiObject *types.SignedObject) []interface{} {
	if apiObject == nil || apiObject.S3 == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3": flattenSigningJobS3SignedObject(apiObject.S3),
	}

	return []interface{}{tfMap}
}

func flattenSigningJobS3SignedObject(apiObject *types.S3SignedObject) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func findSigningJobByID(ctx context.Context, conn *signer.Client, id string) (*signer.DescribeSigningJobOutput, error) {
	in := &signer.DescribeSigningJobInput{
		JobId: aws.String(id),
	}

	out, err := conn.DescribeSigningJob(ctx, in)

	if err != nil {
		return nil, err
	}

	var nfe *types.ResourceNotFoundException
	if errors.As(err, &nfe) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
