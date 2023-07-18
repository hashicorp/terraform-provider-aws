// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for asynchronous validation on SSM Parameter creation.
	parameterCreationValidationTimeout = 2 * time.Minute
)

// @SDKResource("aws_ssm_parameter", name="Parameter")
// @Tags(identifierAttribute="id", resourceType="Parameter")
func ResourceParameter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterCreate,
		ReadWithoutTimeout:   resourceParameterRead,
		UpdateWithoutTimeout: resourceParameterUpdate,
		DeleteWithoutTimeout: resourceParameterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"data_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws:ec2:image",
					"aws:ssm:integration",
					"text",
				}, false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"insecure_value": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"insecure_value", "value"},
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"overwrite": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: "this attribute has been deprecated",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ssm.ParameterTier_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old != "" {
						return new == ssm.ParameterTierIntelligentTiering
					}
					return false
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ssm.ParameterType_Values(), false),
			},
			"value": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ExactlyOneOf: []string{"insecure_value", "value"},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			// Prevent the following error during tier update from Advanced to Standard:
			// ValidationException: This parameter uses the advanced-parameter tier. You can't downgrade a parameter from the advanced-parameter tier to the standard-parameter tier. If necessary, you can delete the advanced parameter and recreate it as a standard parameter.
			customdiff.ForceNewIfChange("tier", func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) == ssm.ParameterTierAdvanced && new.(string) == ssm.ParameterTierStandard
			}),
			customdiff.ComputedIf("version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("value")
			}),
			customdiff.ComputedIf("value", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("insecure_value")
			}),
			customdiff.ComputedIf("insecure_value", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("value")
			}),

			verify.SetTagsDiff,
		),
	}
}

func resourceParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	name := d.Get("name").(string)

	value := d.Get("value").(string)
	if v, ok := d.Get("insecure_value").(string); ok && v != "" {
		value = v
	}

	input := &ssm.PutParameterInput{
		Name:           aws.String(name),
		Type:           aws.String(d.Get("type").(string)),
		Value:          aws.String(value),
		Overwrite:      aws.Bool(ShouldUpdateParameter(d)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if v, ok := d.GetOk("tier"); ok {
		input.Tier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_type"); ok {
		input.DataType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if keyID, ok := d.GetOk("key_id"); ok && d.Get("type").(string) == ssm.ParameterTypeSecureString {
		input.SetKeyId(keyID.(string))
	}

	// AWS SSM Service only supports PutParameter requests with Tags
	// iff Overwrite is not provided or is false; in this resource's case,
	// the Overwrite value is always set in the paramInput so we check for the value
	tags := getTagsIn(ctx)
	if !aws.BoolValue(input.Overwrite) {
		input.Tags = tags
	}

	_, err := conn.PutParameterWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "Tier is not supported") {
		log.Printf("[WARN] Creating SSM Parameter (%s): tier %q not supported, using default", name, d.Get("tier").(string))
		input.Tier = nil
		_, err = conn.PutParameterWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Parameter (%s): %s", name, err)
	}

	// Since the AWS SSM Service does not support PutParameter requests with
	// Tags and Overwrite set to true, we make an additional API call
	// to Update the resource's tags if necessary
	if len(tags) > 0 && input.Tags == nil {
		if err := createTags(ctx, conn, name, ssm.ResourceTypeForTaggingParameter, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SSM Parameter (%s) tags: %s", name, err)
		}
	}

	d.SetId(name)

	return append(diags, resourceParameterRead(ctx, d, meta)...)
}

func resourceParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	input := &ssm.GetParameterInput{
		Name:           aws.String(d.Id()),
		WithDecryption: aws.Bool(true),
	}

	var resp *ssm.GetParameterOutput
	err := retry.RetryContext(ctx, parameterCreationValidationTimeout, func() *retry.RetryError {
		var err error
		resp, err = conn.GetParameterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) && d.IsNewResource() && d.Get("data_type").(string) == "aws:ec2:image" {
			return retry.RetryableError(fmt.Errorf("reading SSM Parameter (%s) after creation: this can indicate that the provided parameter value could not be validated by SSM", d.Id()))
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetParameterWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) && !d.IsNewResource() {
		log.Printf("[WARN] SSM Parameter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s): %s", d.Id(), err)
	}

	param := resp.Parameter
	d.Set("arn", param.ARN)
	name := aws.StringValue(param.Name)
	d.Set("name", name)
	d.Set("type", param.Type)
	d.Set("version", param.Version)

	if _, ok := d.GetOk("insecure_value"); ok && aws.StringValue(param.Type) != ssm.ParameterTypeSecureString {
		d.Set("insecure_value", param.Value)
	} else {
		d.Set("value", param.Value)
	}

	if aws.StringValue(param.Type) == ssm.ParameterTypeSecureString && d.Get("insecure_value").(string) != "" {
		return sdkdiag.AppendErrorf(diags, "invalid configuration, cannot set type = %s and insecure_value", aws.StringValue(param.Type))
	}

	describeParamsInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Equals"),
				Values: []*string{aws.String(name)},
			},
		},
	}
	describeResp, err := conn.DescribeParametersWithContext(ctx, describeParamsInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing SSM parameter (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && (describeResp == nil || len(describeResp.Parameters) == 0 || describeResp.Parameters[0] == nil) {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	detail := describeResp.Parameters[0]
	d.Set("key_id", detail.KeyId)
	d.Set("description", detail.Description)
	d.Set("tier", detail.Tier)
	d.Set("allowed_pattern", detail.AllowedPattern)
	d.Set("data_type", detail.DataType)

	return diags
}

func resourceParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	if d.HasChangesExcept("overwrite", "tags", "tags_all") {
		value := d.Get("value").(string)

		if v, ok := d.Get("insecure_value").(string); ok && v != "" {
			value = v
		}
		paramInput := &ssm.PutParameterInput{
			Name:           aws.String(d.Get("name").(string)),
			Type:           aws.String(d.Get("type").(string)),
			Tier:           aws.String(d.Get("tier").(string)),
			Value:          aws.String(value),
			Overwrite:      aws.Bool(ShouldUpdateParameter(d)),
			AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
		}

		// Retrieve the value set in the config directly to counteract the DiffSuppressFunc above
		tier := d.GetRawConfig().GetAttr("tier")
		if tier.IsKnown() && !tier.IsNull() {
			paramInput.Tier = aws.String(tier.AsString())
		}

		if d.HasChange("data_type") {
			paramInput.DataType = aws.String(d.Get("data_type").(string))
		}

		if d.HasChange("description") {
			paramInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("key_id") && d.Get("type").(string) == ssm.ParameterTypeSecureString {
			paramInput.SetKeyId(d.Get("key_id").(string))
		}

		_, err := conn.PutParameterWithContext(ctx, paramInput)

		if tfawserr.ErrMessageContains(err, "ValidationException", "Tier is not supported") {
			log.Printf("[WARN] Updating SSM Parameter (%s): tier %q not supported, using default", d.Get("name").(string), d.Get("tier").(string))
			paramInput.Tier = nil
			_, err = conn.PutParameterWithContext(ctx, paramInput)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Parameter (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceParameterRead(ctx, d, meta)...)
}

func resourceParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	_, err := conn.DeleteParameterWithContext(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Parameter (%s): %s", d.Id(), err)
	}

	return diags
}

func ShouldUpdateParameter(d *schema.ResourceData) bool {
	// If the user has specified a preference, return their preference
	if value, ok := d.GetOkExists("overwrite"); ok {
		return value.(bool)
	}

	// Since the user has not specified a preference, obey lifecycle rules
	// if it is not a new resource, otherwise overwrite should be set to false.
	return !d.IsNewResource()
}
