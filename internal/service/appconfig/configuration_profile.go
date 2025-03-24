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
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

type configurationProfileType string

const (
	configurationProfileTypeFeatureFlags configurationProfileType = "AWS.AppConfig.FeatureFlags"
	configurationProfileTypeFreeform     configurationProfileType = "AWS.Freeform"
)

func (configurationProfileType) Values() []configurationProfileType {
	return []configurationProfileType{
		configurationProfileTypeFeatureFlags,
		configurationProfileTypeFreeform,
	}
}

// @SDKResource("aws_appconfig_configuration_profile", name="Configuration Profile")
// @Tags(identifierAttribute="arn")
func resourceConfigurationProfile() *schema.Resource {
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
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"retrieval_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          configurationProfileTypeFreeform,
				ValidateDiagFunc: enum.Validate[configurationProfileType](),
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
	}
}

func resourceConfigurationProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID := d.Get(names.AttrApplicationID).(string)
	name := d.Get(names.AttrName).(string)
	input := appconfig.CreateConfigurationProfileInput{
		ApplicationId: aws.String(appID),
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

	output, err := conn.CreateConfigurationProfile(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Configuration Profile (%s) for Application (%s): %s", name, appID, err)
	}

	d.SetId(configurationProfileCreateResourceID(aws.ToString(output.Id), appID))

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	configurationProfileID, applicationID, err := configurationProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findConfigurationProfileByTwoPartKey(ctx, conn, applicationID, configurationProfileID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Configuration Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrARN, configurationProfileARN(ctx, meta.(*conns.AWSClient), applicationID, configurationProfileID))
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

	return diags
}

func resourceConfigurationProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		configurationProfileID, applicationID, err := configurationProfileParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := appconfig.UpdateConfigurationProfileInput{
			ApplicationId:          aws.String(applicationID),
			ConfigurationProfileId: aws.String(configurationProfileID),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("kms_key_identifier") {
			input.KmsKeyIdentifier = aws.String(d.Get("kms_key_identifier").(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("retrieval_role_arn") {
			input.RetrievalRoleArn = aws.String(d.Get("retrieval_role_arn").(string))
		}

		if d.HasChange("validator") {
			input.Validators = expandValidators(d.Get("validator").(*schema.Set).List())
		}

		_, err = conn.UpdateConfigurationProfile(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Configuration Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationProfileRead(ctx, d, meta)...)
}

func resourceConfigurationProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	configurationProfileID, applicationID, err := configurationProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Configuration Profile: %s", d.Id())
	input := appconfig.DeleteConfigurationProfileInput{
		ApplicationId:          aws.String(applicationID),
		ConfigurationProfileId: aws.String(configurationProfileID),
	}
	_, err = conn.DeleteConfigurationProfile(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Configuration Profile (%s): %s", d.Id(), err)
	}

	return diags
}

const configurationProfileResourceIDSeparator = ":"

func configurationProfileCreateResourceID(configurationProfileID, applicationID string) string {
	parts := []string{configurationProfileID, applicationID}
	id := strings.Join(parts, configurationProfileResourceIDSeparator)

	return id
}

func configurationProfileParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, configurationProfileResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ConfigurationProfileID%[2]sApplicationID", id, configurationProfileResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findConfigurationProfileByTwoPartKey(ctx context.Context, conn *appconfig.Client, applicationID, configurationProfileID string) (*appconfig.GetConfigurationProfileOutput, error) {
	input := appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(applicationID),
		ConfigurationProfileId: aws.String(configurationProfileID),
	}

	return findConfigurationProfile(ctx, conn, &input)
}

func findConfigurationProfile(ctx context.Context, conn *appconfig.Client, input *appconfig.GetConfigurationProfileInput) (*appconfig.GetConfigurationProfileOutput, error) {
	output, err := conn.GetConfigurationProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandValidator(tfMap map[string]any) awstypes.Validator {
	apiObject := awstypes.Validator{}

	// AppConfig API supports empty content
	if v, ok := tfMap[names.AttrContent].(string); ok {
		apiObject.Content = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.ValidatorType(v)
	}

	return apiObject
}

func expandValidators(tfList []any) []awstypes.Validator {
	// AppConfig API requires a 0 length slice instead of a nil value
	// when updating from N apiObjects to 0/nil apiObjects
	apiObjects := make([]awstypes.Validator, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandValidator(tfMap))
	}

	return apiObjects
}

func flattenValidator(apiObject awstypes.Validator) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Content; v != nil {
		tfMap[names.AttrContent] = aws.ToString(v)
	}

	tfMap[names.AttrType] = string(apiObject.Type)

	return tfMap
}

func flattenValidators(apiObjects []awstypes.Validator) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, validator := range apiObjects {
		tfList = append(tfList, flattenValidator(validator))
	}

	return tfList
}

func configurationProfileARN(ctx context.Context, c *conns.AWSClient, applicationID, configurationProfileID string) string {
	return c.RegionalARN(ctx, "appconfig", "application/"+applicationID+"/configurationprofile/"+configurationProfileID)
}
