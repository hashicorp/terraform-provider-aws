// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_parameter", name="Parameter")
// @Tags(identifierAttribute="id", resourceType="Parameter")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssm/types;awstypes;awstypes.Parameter")
func resourceParameter() *schema.Resource {
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
			names.AttrARN: {
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
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"insecure_value": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"insecure_value", names.AttrValue},
			},
			names.AttrKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ParameterTier](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old != "" {
						return awstypes.ParameterTier(new) == awstypes.ParameterTierIntelligentTiering
					}
					return false
				},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ParameterType](),
			},
			names.AttrValue: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ExactlyOneOf: []string{"insecure_value", names.AttrValue},
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			// Prevent the following error during tier update from Advanced to Standard:
			// ValidationException: This parameter uses the advanced-parameter tier. You can't downgrade a parameter from the advanced-parameter tier to the standard-parameter tier. If necessary, you can delete the advanced parameter and recreate it as a standard parameter.
			customdiff.ForceNewIfChange("tier", func(_ context.Context, old, new, meta interface{}) bool {
				return awstypes.ParameterTier(old.(string)) == awstypes.ParameterTierAdvanced && awstypes.ParameterTier(new.(string)) == awstypes.ParameterTierStandard
			}),
			customdiff.ComputedIf(names.AttrVersion, func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange(names.AttrValue)
			}),
			customdiff.ComputedIf(names.AttrValue, func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("insecure_value")
			}),
			customdiff.ComputedIf("insecure_value", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange(names.AttrValue)
			}),

			verify.SetTagsDiff,
		),
	}
}

func resourceParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	typ := awstypes.ParameterType(d.Get(names.AttrType).(string))
	value := d.Get(names.AttrValue).(string)
	if v, ok := d.Get("insecure_value").(string); ok && v != "" {
		value = v
	}
	input := &ssm.PutParameterInput{
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
		Name:           aws.String(name),
		Overwrite:      aws.Bool(shouldUpdateParameter(d)),
		Type:           typ,
		Value:          aws.String(value),
	}

	if v, ok := d.GetOk("data_type"); ok {
		input.DataType = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKeyID); ok && typ == awstypes.ParameterTypeSecureString {
		input.KeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tier"); ok {
		input.Tier = awstypes.ParameterTier(v.(string))
	}

	// AWS SSM Service only supports PutParameter requests with Tags
	// iff Overwrite is not provided or is false; in this resource's case,
	// the Overwrite value is always set in the paramInput so we check for the value
	tags := getTagsIn(ctx)
	if !aws.ToBool(input.Overwrite) {
		input.Tags = tags
	}

	_, err := conn.PutParameter(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeValidationException, "Tier is not supported") {
		log.Printf("[WARN] Creating SSM Parameter (%s): tier %q not supported, using default", name, d.Get("tier").(string))
		input.Tier = ""
		_, err = conn.PutParameter(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Parameter (%s): %s", name, err)
	}

	// Since the AWS SSM Service does not support PutParameter requests with
	// Tags and Overwrite set to true, we make an additional API call
	// to Update the resource's tags if necessary
	if len(tags) > 0 && input.Tags == nil {
		if err := createTags(ctx, conn, name, string(awstypes.ResourceTypeForTaggingParameter), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SSM Parameter (%s) tags: %s", name, err)
		}
	}

	d.SetId(name)

	return append(diags, resourceParameterRead(ctx, d, meta)...)
}

func resourceParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	const (
		// Maximum amount of time to wait for asynchronous validation on SSM Parameter creation.
		timeout = 2 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
			return findParameterByName(ctx, conn, d.Id(), true)
		},
		func(err error) (bool, error) {
			if d.IsNewResource() && tfresource.NotFound(err) && d.Get("data_type").(string) == "aws:ec2:image" {
				return true, err
			}

			return false, err
		},
	)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Parameter %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s): %s", d.Id(), err)
	}

	param := outputRaw.(*awstypes.Parameter)
	d.Set(names.AttrARN, param.ARN)
	d.Set(names.AttrName, param.Name)
	d.Set(names.AttrType, param.Type)
	d.Set(names.AttrVersion, param.Version)

	if _, ok := d.GetOk("insecure_value"); ok && param.Type != awstypes.ParameterTypeSecureString {
		d.Set("insecure_value", param.Value)
	} else {
		d.Set(names.AttrValue, param.Value)
	}

	if param.Type == awstypes.ParameterTypeSecureString && d.Get("insecure_value").(string) != "" {
		return sdkdiag.AppendErrorf(diags, "invalid configuration, cannot set type = %s and insecure_value", param.Type)
	}

	detail, err := findParameterMetadataByName(ctx, conn, d.Get(names.AttrName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Parameter %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Parameter metadata (%s): %s", d.Id(), err)
	}

	d.Set("allowed_pattern", detail.AllowedPattern)
	d.Set("data_type", detail.DataType)
	d.Set(names.AttrDescription, detail.Description)
	d.Set(names.AttrKeyID, detail.KeyId)
	d.Set("tier", detail.Tier)

	return diags
}

func resourceParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if d.HasChangesExcept("overwrite", names.AttrTags, names.AttrTagsAll) {
		typ := awstypes.ParameterType(d.Get(names.AttrType).(string))
		value := d.Get(names.AttrValue).(string)
		if v, ok := d.Get("insecure_value").(string); ok && v != "" {
			value = v
		}
		input := &ssm.PutParameterInput{
			AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
			Name:           aws.String(d.Id()),
			Overwrite:      aws.Bool(shouldUpdateParameter(d)),
			Tier:           awstypes.ParameterTier(d.Get("tier").(string)),
			Type:           typ,
			Value:          aws.String(value),
		}

		if d.HasChange("data_type") {
			input.DataType = aws.String(d.Get("data_type").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrKeyID) && typ == awstypes.ParameterTypeSecureString {
			input.KeyId = aws.String(d.Get(names.AttrKeyID).(string))
		}

		// Retrieve the value set in the config directly to counteract the DiffSuppressFunc above.
		if v := d.GetRawConfig().GetAttr("tier"); v.IsKnown() && !v.IsNull() {
			input.Tier = awstypes.ParameterTier(v.AsString())
		}

		_, err := conn.PutParameter(ctx, input)

		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Tier is not supported") {
			log.Printf("[WARN] Creating SSM Parameter (%s): tier %q not supported, using default", d.Id(), d.Get("tier").(string))
			input.Tier = ""
			_, err = conn.PutParameter(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Parameter (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceParameterRead(ctx, d, meta)...)
}

func resourceParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Parameter: %s", d.Id())
	_, err := conn.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ParameterNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Parameter (%s): %s", d.Id(), err)
	}

	return diags
}

func findParameterByName(ctx context.Context, conn *ssm.Client, name string, withDecryption bool) (*awstypes.Parameter, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(withDecryption),
	}

	output, err := conn.GetParameter(ctx, input)

	if errs.IsA[*awstypes.ParameterNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Parameter == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Parameter, nil
}

func findParameterMetadataByName(ctx context.Context, conn *ssm.Client, name string) (*awstypes.ParameterMetadata, error) {
	input := &ssm.DescribeParametersInput{
		ParameterFilters: []awstypes.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Equals"),
				Values: []string{name},
			},
		},
	}

	return findParameterMetadata(ctx, conn, input)
}

func findParameterMetadata(ctx context.Context, conn *ssm.Client, input *ssm.DescribeParametersInput) (*awstypes.ParameterMetadata, error) {
	output, err := findParametersMetadata(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findParametersMetadata(ctx context.Context, conn *ssm.Client, input *ssm.DescribeParametersInput) ([]awstypes.ParameterMetadata, error) {
	var output []awstypes.ParameterMetadata

	pages := ssm.NewDescribeParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Parameters...)
	}

	return output, nil
}

func shouldUpdateParameter(d *schema.ResourceData) bool {
	// If the user has specified a preference, return their preference.
	if v := d.GetRawConfig().GetAttr("overwrite"); v.IsKnown() && !v.IsNull() {
		return v.True()
	}

	// Since the user has not specified a preference, obey lifecycle rules
	// if it is not a new resource, otherwise overwrite should be set to false.
	return !d.IsNewResource()
}
