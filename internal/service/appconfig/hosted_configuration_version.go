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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_hosted_configuration_version", name="Hosted Configuration Version")
func resourceHostedConfigurationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedConfigurationVersionCreate,
		ReadWithoutTimeout:   resourceHostedConfigurationVersionRead,
		DeleteWithoutTimeout: resourceHostedConfigurationVersionDelete,

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			names.AttrContent: {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			names.AttrContentType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"version_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceHostedConfigurationVersionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	configurationProfileID := d.Get("configuration_profile_id").(string)
	input := &appconfig.CreateHostedConfigurationVersionInput{
		ApplicationId:          aws.String(applicationID),
		ConfigurationProfileId: aws.String(configurationProfileID),
		Content:                []byte(d.Get(names.AttrContent).(string)),
		ContentType:            aws.String(d.Get(names.AttrContentType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHostedConfigurationVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Hosted Configuration Version for Application (%s): %s", applicationID, err)
	}

	d.SetId(hostedConfigurationVersionCreateResourceID(aws.ToString(output.ApplicationId), aws.ToString(output.ConfigurationProfileId), output.VersionNumber))

	return append(diags, resourceHostedConfigurationVersionRead(ctx, d, meta)...)
}

func resourceHostedConfigurationVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID, configurationProfileID, versionNumber, err := hostedConfigurationVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findHostedConfigurationVersionByThreePartKey(ctx, conn, applicationID, configurationProfileID, versionNumber)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Hosted Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrARN, hostedConfigurationVersionARN(ctx, meta.(*conns.AWSClient), applicationID, configurationProfileID, versionNumber))
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set(names.AttrContent, string(output.Content))
	d.Set(names.AttrContentType, output.ContentType)
	d.Set(names.AttrDescription, output.Description)
	d.Set("version_number", output.VersionNumber)

	return diags
}

func resourceHostedConfigurationVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID, configurationProfileID, versionNumber, err := hostedConfigurationVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Hosted Configuration Version: %s", d.Id())
	input := appconfig.DeleteHostedConfigurationVersionInput{
		ApplicationId:          aws.String(applicationID),
		ConfigurationProfileId: aws.String(configurationProfileID),
		VersionNumber:          aws.Int32(versionNumber),
	}
	_, err = conn.DeleteHostedConfigurationVersion(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	return diags
}

const hostedConfigurationVersionResourceIDSeparator = "/"

func hostedConfigurationVersionCreateResourceID(applicationID, configurationProfileID string, versionNumber int32) string {
	parts := []string{applicationID, configurationProfileID, flex.Int32ValueToStringValue(versionNumber)}
	id := strings.Join(parts, hostedConfigurationVersionResourceIDSeparator)

	return id
}

func hostedConfigurationVersionParseResourceID(id string) (string, string, int32, error) {
	parts := strings.Split(id, hostedConfigurationVersionResourceIDSeparator)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected ApplicationID%[2]sConfigurationProfileID%[2]sVersionNumber", id, hostedConfigurationVersionResourceIDSeparator)
	}

	return parts[0], parts[1], flex.StringValueToInt32Value(parts[2]), nil
}

func findHostedConfigurationVersionByThreePartKey(ctx context.Context, conn *appconfig.Client, applicationID, configurationProfileID string, versionNumber int32) (*appconfig.GetHostedConfigurationVersionOutput, error) {
	input := appconfig.GetHostedConfigurationVersionInput{
		ApplicationId:          aws.String(applicationID),
		ConfigurationProfileId: aws.String(configurationProfileID),
		VersionNumber:          aws.Int32(versionNumber),
	}

	return findHostedConfigurationVersion(ctx, conn, &input)
}

func findHostedConfigurationVersion(ctx context.Context, conn *appconfig.Client, input *appconfig.GetHostedConfigurationVersionInput) (*appconfig.GetHostedConfigurationVersionOutput, error) {
	output, err := conn.GetHostedConfigurationVersion(ctx, input)

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

func hostedConfigurationVersionARN(ctx context.Context, c *conns.AWSClient, applicationID, configurationProfileID string, versionNumber int32) string {
	return c.RegionalARN(ctx, "appconfig", "application/"+applicationID+"/configurationprofile/"+configurationProfileID+"/hostedconfigurationversion/"+flex.Int32ValueToStringValue(versionNumber))
}
