// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
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

// @SDKResource("aws_qbusiness_webexperience", name="Webexperience")
// @Tags(identifierAttribute="arn")
func ResourceWebexperience() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceWebexperienceCreate,
		ReadWithoutTimeout:   resourceWebexperienceRead,
		UpdateWithoutTimeout: resourceWebexperienceUpdate,
		DeleteWithoutTimeout: resourceWebexperienceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 36),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
				),
			},
			"arn": {
				Type:        schema.TypeString,
				Description: "The Amazon Resource Name (ARN) of the Amazon Q application.",
				Computed:    true,
			},
			"authentication_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"saml_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metadata_xml": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The SAML metadata document provided by your identity provider (IdP).",
										ValidateFunc: validation.StringLenBetween(1000, 1000000),
									},
									"iam_role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "ARN of an IAM role assumed by users when they authenticate into their Amazon Q web experience, containing the relevant Amazon Q permissions for conversing with Amazon Q,",
										ValidateFunc: verify.ValidARN,
									},
									"user_id_attribute": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The user attribute name in your IdP that maps to the user email.",
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"user_group_attribute": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "The group attribute name in your IdP that maps to user groups.",
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
			},
			"sample_propmpts_control_mode": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Sample prompts control mode for Amazon Q web experience.",
				ValidateDiagFunc: enum.Validate[types.WebExperienceSamplePromptsControlMode](),
			},
			"subtitle": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subtitle for Amazon Q web experience.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 500),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Title for Amazon Q web experience.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 500),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"webexperience_id": {
				Type:        schema.TypeString,
				Description: "Identifier of the Amazon Q web experience",
				Computed:    true,
			},
			"welcome_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customized welcome message for end users of an Amazon Q web experience",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 300),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWebexperienceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id := d.Get("application_id").(string)

	input := &qbusiness.CreateWebExperienceInput{
		ApplicationId:            aws.String(application_id),
		SamplePromptsControlMode: types.WebExperienceSamplePromptsControlMode(d.Get("sample_propmpts_control_mode").(string)),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("welcome_message"); ok {
		input.WelcomeMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subtitle"); ok {
		input.Subtitle = aws.String(v.(string))
	}

	if v, ok := d.GetOk("title"); ok {
		input.Title = aws.String(v.(string))
	}

	output, err := conn.CreateWebExperience(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness webexperience: %s", err)
	}

	d.SetId(application_id + "/" + aws.ToString(output.WebExperienceId))

	if _, err := waitWebexperienceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness webexperience (%s): waiting for completion: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 {
		input := &qbusiness.UpdateWebExperienceInput{
			ApplicationId:               aws.String(application_id),
			WebExperienceId:             output.WebExperienceId,
			AuthenticationConfiguration: expandAuthenticationConfiguration(v.([]interface{})),
		}

		_, err := conn.UpdateWebExperience(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating qbusiness webexperience (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebexperienceRead(ctx, d, meta)...)
}

func resourceWebexperienceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, webexperience_id, err := parseWebexperienceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness webexperience (%s): %s", d.Id(), err)
	}

	input := &qbusiness.UpdateWebExperienceInput{
		ApplicationId:   aws.String(application_id),
		WebExperienceId: aws.String(webexperience_id),
	}

	if d.HasChange("sample_propmpts_control_mode") {
		input.SamplePromptsControlMode = types.WebExperienceSamplePromptsControlMode(d.Get("sample_propmpts_control_mode").(string))
	}

	if d.HasChange("welcome_message") {
		input.WelcomeMessage = aws.String(d.Get("welcome_message").(string))
	}

	if d.HasChange("subtitle") {
		input.Subtitle = aws.String(d.Get("subtitle").(string))
	}

	if d.HasChange("title") {
		input.Title = aws.String(d.Get("title").(string))
	}

	if d.HasChange("authentication_configuration") {
		input.AuthenticationConfiguration = expandAuthenticationConfiguration(d.Get("authentication_configuration").([]interface{}))
	}

	_, err = conn.UpdateWebExperience(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness webexperience (%s): %s", d.Id(), err)
	}

	return append(diags, resourceWebexperienceRead(ctx, d, meta)...)
}

func resourceWebexperienceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	webex, err := FindWebexperienceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] qbusiness webexperience (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness webexperience: %s", err)
	}

	d.Set("arn", webex.WebExperienceArn)
	d.Set("webexperience_id", webex.WebExperienceId)
	d.Set("application_id", webex.ApplicationId)
	d.Set("sample_propmpts_control_mode", webex.SamplePromptsControlMode)
	d.Set("welcome_message", webex.WelcomeMessage)
	d.Set("subtitle", webex.Subtitle)
	d.Set("title", webex.Title)

	if webex.AuthenticationConfiguration != nil {
		if err := d.Set("authentication_configuration", flattenAuthenticationConfiguration(webex.AuthenticationConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting qbusiness webexperience authentication_configuration: %s", err)
		}
	}
	return diags
}

func resourceWebexperienceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, webexperience_id, err := parseWebexperienceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness webexperience (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteWebExperience(ctx, &qbusiness.DeleteWebExperienceInput{
		ApplicationId:   aws.String(application_id),
		WebExperienceId: aws.String(webexperience_id),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness webexperience (%s): %s", d.Id(), err)
	}

	if _, err := waitWebexperienceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness webexperience (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func parseWebexperienceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid webexperience ID: %s", id)
	}

	return parts[0], parts[1], nil
}

func FindWebexperienceByID(ctx context.Context, conn *qbusiness.Client, webexperience_id string) (*qbusiness.GetWebExperienceOutput, error) {
	application_d, webex_id, err := parseWebexperienceID(webexperience_id)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetWebExperienceInput{
		ApplicationId:   aws.String(application_d),
		WebExperienceId: aws.String(webex_id),
	}

	output, err := conn.GetWebExperience(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandAuthenticationConfiguration(l []interface{}) types.WebExperienceAuthConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["saml_configuration"]; ok {
		return expandSamlConfiguration(v.([]interface{}))
	}

	return nil
}

func expandSamlConfiguration(l []interface{}) *types.WebExperienceAuthConfigurationMemberSamlConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})

	saml_conf := types.SamlConfiguration{
		MetadataXML:     aws.String(m["metadata_xml"].(string)),
		RoleArn:         aws.String(m["iam_role_arn"].(string)),
		UserIdAttribute: aws.String(m["user_id_attribute"].(string)),
	}

	if v, ok := m["user_group_attribute"].(string); ok && v != "" {
		saml_conf.UserGroupAttribute = aws.String(v)
	}

	return &types.WebExperienceAuthConfigurationMemberSamlConfiguration{
		Value: saml_conf,
	}
}

func flattenAuthenticationConfiguration(authenticationConfiguration types.WebExperienceAuthConfiguration) []interface{} {
	if authenticationConfiguration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v, ok := authenticationConfiguration.(*types.WebExperienceAuthConfigurationMemberSamlConfiguration); ok {
		m["saml_configuration"] = flattenSamlConfiguration(v)
	}

	return []interface{}{m}
}

func flattenSamlConfiguration(samlConfiguration *types.WebExperienceAuthConfigurationMemberSamlConfiguration) []interface{} {
	if samlConfiguration == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"metadata_xml":         samlConfiguration.Value.MetadataXML,
			"iam_role_arn":         samlConfiguration.Value.RoleArn,
			"user_id_attribute":    samlConfiguration.Value.UserIdAttribute,
			"user_group_attribute": samlConfiguration.Value.UserGroupAttribute,
		},
	}
}
