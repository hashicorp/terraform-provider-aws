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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
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
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	appID := d.Get("application_id").(string)
	profileID := d.Get("configuration_profile_id").(string)

	input := &appconfig.CreateHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(profileID),
		Content:                []byte(d.Get("content").(string)),
		ContentType:            aws.String(d.Get("content_type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHostedConfigurationVersionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig HostedConfigurationVersion for Application (%s): %s", appID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", aws.StringValue(output.ApplicationId), aws.StringValue(output.ConfigurationProfileId), aws.Int64Value(output.VersionNumber)))

	return append(diags, resourceHostedConfigurationVersionRead(ctx, d, meta)...)
}

func resourceHostedConfigurationVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	appID, confProfID, versionNumber, err := HostedConfigurationVersionParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int64(int64(versionNumber)),
	}

	output, err := conn.GetHostedConfigurationVersionWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
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

	d.Set("application_id", output.ApplicationId)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set("content", string(output.Content))
	d.Set("content_type", output.ContentType)
	d.Set("description", output.Description)
	d.Set("version_number", output.VersionNumber)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s/hostedconfigurationversion/%d", appID, confProfID, versionNumber),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)

	return diags
}

func resourceHostedConfigurationVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	appID, confProfID, versionNumber, err := HostedConfigurationVersionParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting AppConfig Hosted Configuration Version: %s", d.Id())
	_, err = conn.DeleteHostedConfigurationVersionWithContext(ctx, &appconfig.DeleteHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int64(int64(versionNumber)),
	})

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Hosted Configuration Version (%s): %s", d.Id(), err)
	}

	return diags
}

func HostedConfigurationVersionParseID(id string) (string, string, int, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected ApplicationID/ConfigurationProfileID/VersionNumber", id)
	}

	version, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", "", 0, fmt.Errorf("parsing Hosted Configuration Version version_number: %w", err)
	}

	return parts[0], parts[1], version, nil
}
