// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kendra_faq", name="FAQ")
// @Tags(identifierAttribute="arn")
func ResourceFaq() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFaqCreate,
		ReadWithoutTimeout:   resourceFaqRead,
		UpdateWithoutTimeout: resourceFaqUpdate,
		DeleteWithoutTimeout: resourceFaqDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"faq_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_format": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.FaqFileFormat](),
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			names.AttrLanguageCode: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(
						regexache.MustCompile(`[A-Za-z-]*`),
						"Must have alphanumeric characters or hyphens.",
					),
				),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(
						regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_path": {
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
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFaqCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kendra.CreateFaqInput{
		ClientToken: aws.String(id.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(name),
		RoleArn:     aws.String(d.Get(names.AttrRoleARN).(string)),
		S3Path:      expandS3Path(d.Get("s3_path").([]interface{})),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_format"); ok {
		input.FileFormat = types.FaqFileFormat(v.(string))
	}

	if v, ok := d.GetOk(names.AttrLanguageCode); ok {
		input.LanguageCode = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateFaq(ctx, input)
		},
		func(err error) (bool, error) {
			var validationException *types.ValidationException

			if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kendra Faq (%s): %s", name, err)
	}

	if outputRaw == nil {
		return sdkdiag.AppendErrorf(diags, "creating Kendra Faq (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateFaqOutput)

	id := aws.ToString(output.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitFaqCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Faq (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceFaqRead(ctx, d, meta)...)
}

func resourceFaqRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id, indexId, err := FaqParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := FindFaqByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Faq (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Kendra Faq (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/faq/%s", indexId, id),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("faq_id", resp.Id)
	d.Set("file_format", resp.FileFormat)
	d.Set("index_id", resp.IndexId)
	d.Set(names.AttrLanguageCode, resp.LanguageCode)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrRoleARN, resp.RoleArn)
	d.Set(names.AttrStatus, resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set("s3_path", flattenS3Path(resp.S3Path)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceFaqUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceFaqRead(ctx, d, meta)
}

func resourceFaqDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	log.Printf("[INFO] Deleting Kendra Faq %s", d.Id())

	id, indexId, err := FaqParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteFaq(ctx, &kendra.DeleteFaqInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kendra Faq (%s): %s", d.Id(), err)
	}

	if _, err := waitFaqDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kendra Faq (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitFaqCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeFaqOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.FaqStatusCreating, "PENDING_CREATION"), // API currently returns PENDING_CREATION instead of CREATING
		Target:                    enum.Slice(types.FaqStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusFaq(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeFaqOutput); ok {
		if output.Status == types.FaqStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitFaqDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeFaqOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.FaqStatusDeleting, "PENDING_DELETION"), // API currently returns PENDING_DELETION instead of DELETING
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusFaq(ctx, conn, id, indexId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeFaqOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func statusFaq(ctx context.Context, conn *kendra.Client, id, indexId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFaqByID(ctx, conn, id, indexId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func expandS3Path(tfList []interface{}) *types.S3Path {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Path{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func flattenS3Path(apiObject *types.S3Path) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		m[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		m[names.AttrKey] = aws.ToString(v)
	}

	return []interface{}{m}
}
