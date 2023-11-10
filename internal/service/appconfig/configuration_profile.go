// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_configuration_profile", name="Connection Profile")
// @Tags(identifierAttribute="arn")
func ResourceConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationProfileCreate,
		ReadWithoutTimeout:   resourceConfigurationProfileRead,
		UpdateWithoutTimeout: resourceConfigurationProfileUpdate,
		DeleteWithoutTimeout: resourceConfigurationProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"location_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"retrieval_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      configurationProfileTypeFreeform,
				ValidateFunc: validation.StringInSlice(ConfigurationProfileType_Values(), false),
			},
			"validator": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ValidateFunc: validation.Any(
								validation.StringIsJSON,
								verify.ValidARN,
							),
							DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appconfig.ValidatorType_Values(), false),
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConfigurationProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	appId := d.Get("application_id").(string)
	name := d.Get("name").(string)
	input := &appconfig.CreateConfigurationProfileInput{
		ApplicationId: aws.String(appId),
		LocationUri:   aws.String(d.Get("location_uri").(string)),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retrieval_role_arn"); ok {
		input.RetrievalRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	if v, ok := d.GetOk("validator"); ok && v.(*schema.Set).Len() > 0 {
		input.Validators = expandValidators(v.(*schema.Set).List())
	}

	profile, err := conn.CreateConfigurationProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Configuration Profile (%s) for Application (%s): %s", name, appId, err)
	}

	if profile == nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Configuration Profile (%s) for Application (%s): empty response", name, appId)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(profile.Id), aws.StringValue(profile.ApplicationId)))

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	confProfID, appID, err := ConfigurationProfileParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	}

	output, err := conn.GetConfigurationProfileWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] AppConfig Configuration Profile (%s) for Application (%s) not found, removing from state", confProfID, appID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s) for Application (%s): %s", confProfID, appID, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s) for Application (%s): empty response", confProfID, appID)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("configuration_profile_id", output.Id)
	d.Set("description", output.Description)
	d.Set("location_uri", output.LocationUri)
	d.Set("name", output.Name)
	d.Set("retrieval_role_arn", output.RetrievalRoleArn)
	d.Set("type", output.Type)

	if err := d.Set("validator", flattenValidators(output.Validators)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting validator: %s", err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s", appID, confProfID),
		Service:   "appconfig",
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceConfigurationProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		confProfID, appID, err := ConfigurationProfileParseID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Configuration Profile (%s): %s", d.Id(), err)
		}

		updateInput := &appconfig.UpdateConfigurationProfileInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			updateInput.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("retrieval_role_arn") {
			updateInput.RetrievalRoleArn = aws.String(d.Get("retrieval_role_arn").(string))
		}

		if d.HasChange("validator") {
			updateInput.Validators = expandValidators(d.Get("validator").(*schema.Set).List())
		}

		_, err = conn.UpdateConfigurationProfileWithContext(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Configuration Profile (%s) for Application (%s): %s", confProfID, appID, err)
		}
	}

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	confProfID, appID, err := ConfigurationProfileParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Configuration Profile: %s", d.Id())
	_, err = conn.DeleteConfigurationProfileWithContext(ctx, &appconfig.DeleteConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	})

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Configuration Profile (%s) for Application (%s): %s", confProfID, appID, err)
	}

	return diags
}

func ConfigurationProfileParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected ConfigurationProfileID:ApplicationID", id)
	}

	return parts[0], parts[1], nil
}

func expandValidator(tfMap map[string]interface{}) *appconfig.Validator {
	if tfMap == nil {
		return nil
	}

	validator := &appconfig.Validator{}

	// AppConfig API supports empty content
	if v, ok := tfMap["content"].(string); ok {
		validator.Content = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		validator.Type = aws.String(v)
	}

	return validator
}

func expandValidators(tfList []interface{}) []*appconfig.Validator {
	// AppConfig API requires a 0 length slice instead of a nil value
	// when updating from N validators to 0/nil validators
	validators := make([]*appconfig.Validator, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		validator := expandValidator(tfMap)

		if validator == nil {
			continue
		}

		validators = append(validators, validator)
	}

	return validators
}

func flattenValidator(validator *appconfig.Validator) map[string]interface{} {
	if validator == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := validator.Content; v != nil {
		tfMap["content"] = aws.StringValue(v)
	}

	if v := validator.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenValidators(validators []*appconfig.Validator) []interface{} {
	if len(validators) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, validator := range validators {
		if validator == nil {
			continue
		}

		tfList = append(tfList, flattenValidator(validator))
	}

	return tfList
}
