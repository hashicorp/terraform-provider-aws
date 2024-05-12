// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_conformance_pack", name="Conformance Pack")
func resourceConformancePack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConformancePackPut,
		ReadWithoutTimeout:   resourceConformancePackRead,
		UpdateWithoutTimeout: resourceConformancePackPut,
		DeleteWithoutTimeout: resourceConformancePackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_s3_bucket": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"delivery_s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"input_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 60,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and hyphen characters")),
			},
			"template_body": {
				Type:                  schema.TypeString,
				Optional:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONOrYAMLDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					verify.ValidStringIsJSONOrYAML,
				),
				AtLeastOneOf: []string{"template_body", "template_s3_uri"},
			},
			"template_s3_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
				),
				AtLeastOneOf: []string{"template_s3_uri", "template_body"},
			},
		},
	}
}

func resourceConformancePackPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutConformancePackInput{
		ConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("input_parameter"); ok && v.(*schema.Set).Len() > 0 {
		input.ConformancePackInputParameters = expandConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutConformancePack(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Conformance Pack (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	const (
		timeout = 5 * time.Minute
	)
	if _, err := waitConformancePackCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Conformance Pack (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConformancePackRead(ctx, d, meta)...)
}

func resourceConformancePackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	pack, err := findConformancePackByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Conformance Pack (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, pack.ConformancePackArn)
	d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket)
	d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix)
	if err = d.Set("input_parameter", flattenConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_parameter: %s", err)
	}
	d.Set(names.AttrName, pack.ConformancePackName)

	return diags
}

func resourceConformancePackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 5 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Conformance Pack: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteConformancePack(ctx, &configservice.DeleteConformancePackInput{
			ConformancePackName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchConformancePackException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Conformance Pack (%s): %s", d.Id(), err)
	}

	if _, err := waitConformancePackDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Conformance Pack (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConformancePackByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConformancePackDetail, error) {
	input := &configservice.DescribeConformancePacksInput{
		ConformancePackNames: []string{name},
	}

	return findConformancePack(ctx, conn, input)
}

func findConformancePack(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConformancePacksInput) (*types.ConformancePackDetail, error) {
	output, err := findConformancePacks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConformancePacks(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConformancePacksInput) ([]types.ConformancePackDetail, error) {
	var output []types.ConformancePackDetail

	pages := configservice.NewDescribeConformancePacksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchConformancePackException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ConformancePackDetails...)
	}

	return output, nil
}

func findConformancePackStatusByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConformancePackStatusDetail, error) {
	input := &configservice.DescribeConformancePackStatusInput{
		ConformancePackNames: []string{name},
	}

	return findConformancePackStatus(ctx, conn, input)
}

func findConformancePackStatus(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConformancePackStatusInput) (*types.ConformancePackStatusDetail, error) {
	output, err := findConformancePackStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConformancePackStatuses(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConformancePackStatusInput) ([]types.ConformancePackStatusDetail, error) {
	var output []types.ConformancePackStatusDetail

	pages := configservice.NewDescribeConformancePackStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchConformancePackException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ConformancePackStatusDetails...)
	}

	return output, nil
}

func statusConformancePack(ctx context.Context, conn *configservice.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConformancePackStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConformancePackState), err
	}
}

func waitConformancePackCreated(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.ConformancePackStatusDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConformancePackStateCreateInProgress),
		Target:  enum.Slice(types.ConformancePackStateCreateComplete),
		Refresh: statusConformancePack(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ConformancePackStatusDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConformancePackStatusReason)))

		return output, err
	}

	return nil, err
}

func waitConformancePackDeleted(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.ConformancePackStatusDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConformancePackStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusConformancePack(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ConformancePackStatusDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConformancePackStatusReason)))

		return output, err
	}

	return nil, err
}

func expandConformancePackInputParameters(tfList []interface{}) []types.ConformancePackInputParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ConformancePackInputParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.ConformancePackInputParameter{}

		if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
			apiObject.ParameterName = aws.String(v)
		}

		if v, ok := tfMap["parameter_value"].(string); ok && v != "" {
			apiObject.ParameterValue = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConformancePackInputParameters(apiObjects []types.ConformancePackInputParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"parameter_name":  aws.ToString(apiObject.ParameterName),
			"parameter_value": aws.ToString(apiObject.ParameterValue),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
