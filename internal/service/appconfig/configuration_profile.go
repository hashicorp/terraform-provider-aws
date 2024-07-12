// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_configuration_profile", name="Configuration Profile")
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
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			"kms_key_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.Any(
					verify.ValidARN,
					validation.StringLenBetween(1, 256)),
			},
			names.AttrName: {
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
			names.AttrType: {
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
						names.AttrContent: {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ValidateFunc: validation.Any(
								validation.StringIsJSON,
								verify.ValidARN,
							),
							DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ValidatorType](),
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
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appId := d.Get(names.AttrApplicationID).(string)
	name := d.Get(names.AttrName).(string)
	input := &appconfig.CreateConfigurationProfileInput{
		ApplicationId: aws.String(appId),
		LocationUri:   aws.String(d.Get("location_uri").(string)),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_identifier"); ok {
		input.KmsKeyIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retrieval_role_arn"); ok {
		input.RetrievalRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = aws.String(v.(string))
	}

	if v, ok := d.GetOk("validator"); ok && v.(*schema.Set).Len() > 0 {
		input.Validators = expandValidators(v.(*schema.Set).List())
	}

	profile, err := conn.CreateConfigurationProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Configuration Profile (%s) for Application (%s): %s", name, appId, err)
	}

	if profile == nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Configuration Profile (%s) for Application (%s): empty response", name, appId)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.ToString(profile.Id), aws.ToString(profile.ApplicationId)))

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	confProfID, appID, err := ConfigurationProfileParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	}

	output, err := conn.GetConfigurationProfile(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set("configuration_profile_id", output.Id)
	d.Set(names.AttrDescription, output.Description)
	d.Set("kms_key_identifier", output.KmsKeyIdentifier)
	d.Set("location_uri", output.LocationUri)
	d.Set(names.AttrName, output.Name)
	d.Set("retrieval_role_arn", output.RetrievalRoleArn)
	d.Set(names.AttrType, output.Type)

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
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceConfigurationProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		confProfID, appID, err := ConfigurationProfileParseID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Configuration Profile (%s): %s", d.Id(), err)
		}

		updateInput := &appconfig.UpdateConfigurationProfileInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
		}

		if d.HasChange(names.AttrDescription) {
			updateInput.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("kms_key_identifier") {
			updateInput.KmsKeyIdentifier = aws.String(d.Get("kms_key_identifier").(string))
		}

		if d.HasChange(names.AttrName) {
			updateInput.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("retrieval_role_arn") {
			updateInput.RetrievalRoleArn = aws.String(d.Get("retrieval_role_arn").(string))
		}

		if d.HasChange("validator") {
			updateInput.Validators = expandValidators(d.Get("validator").(*schema.Set).List())
		}

		_, err = conn.UpdateConfigurationProfile(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Configuration Profile (%s) for Application (%s): %s", confProfID, appID, err)
		}
	}

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	confProfID, appID, err := ConfigurationProfileParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Configuration Profile: %s", d.Id())
	_, err = conn.DeleteConfigurationProfile(ctx, &appconfig.DeleteConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func expandValidator(tfMap map[string]interface{}) awstypes.Validator {
	validator := awstypes.Validator{}

	// AppConfig API supports empty content
	if v, ok := tfMap[names.AttrContent].(string); ok {
		validator.Content = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		validator.Type = awstypes.ValidatorType(v)
	}

	return validator
}

func expandValidators(tfList []interface{}) []awstypes.Validator {
	// AppConfig API requires a 0 length slice instead of a nil value
	// when updating from N validators to 0/nil validators
	validators := make([]awstypes.Validator, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		validator := expandValidator(tfMap)
		validators = append(validators, validator)
	}

	return validators
}

func flattenValidator(validator awstypes.Validator) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := validator.Content; v != nil {
		tfMap[names.AttrContent] = aws.ToString(v)
	}

	tfMap[names.AttrType] = string(validator.Type)

	return tfMap
}

func flattenValidators(validators []awstypes.Validator) []interface{} {
	if len(validators) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, validator := range validators {
		tfList = append(tfList, flattenValidator(validator))
	}

	return tfList
}
