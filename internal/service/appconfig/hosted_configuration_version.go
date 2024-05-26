// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"
	"strconv"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_hosted_configuration_version")
func ResourceHostedConfigurationVersion() *schema.Resource {
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

func resourceHostedConfigurationVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID := d.Get(names.AttrApplicationID).(string)
	profileID := d.Get("configuration_profile_id").(string)

	input := &appconfig.CreateHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(profileID),
		Content:                []byte(d.Get(names.AttrContent).(string)),
		ContentType:            aws.String(d.Get(names.AttrContentType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHostedConfigurationVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig HostedConfigurationVersion for Application (%s): %s", appID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", aws.ToString(output.ApplicationId), aws.ToString(output.ConfigurationProfileId), output.VersionNumber))

	return append(diags, resourceHostedConfigurationVersionRead(ctx, d, meta)...)
}

func resourceHostedConfigurationVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID, confProfID, versionNumber, err := HostedConfigurationVersionParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int32(versionNumber),
	}

	output, err := conn.GetHostedConfigurationVersion(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Appconfig Hosted Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Hosted Configuration Version (%s): empty response", d.Id())
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set(names.AttrContent, string(output.Content))
	d.Set(names.AttrContentType, output.ContentType)
	d.Set(names.AttrDescription, output.Description)
	d.Set("version_number", output.VersionNumber)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s/hostedconfigurationversion/%d", appID, confProfID, versionNumber),
		Service:   "appconfig",
	}.String()

	d.Set(names.AttrARN, arn)

	return diags
}

func resourceHostedConfigurationVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID, confProfID, versionNumber, err := HostedConfigurationVersionParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Hosted Configuration Version: %s", d.Id())
	_, err = conn.DeleteHostedConfigurationVersion(ctx, &appconfig.DeleteHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int32(versionNumber),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	return diags
}

func HostedConfigurationVersionParseID(id string) (string, string, int32, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected ApplicationID/ConfigurationProfileID/VersionNumber", id)
	}

	version, err := strconv.ParseInt(parts[2], 0, 32)
	if err != nil {
		return "", "", 0, fmt.Errorf("parsing Hosted Configuration Version version_number: %w", err)
	}

	return parts[0], parts[1], int32(version), nil
}
