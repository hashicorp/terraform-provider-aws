// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_configuration_profile", name="Configuration Profile")
// @Tags(identifierAttribute="arn")
func DataSourceConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationProfileRead,
		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"retrieval_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"validator": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrContent: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const (
	DSNameConfigurationProfile = "Configuration Profile Data Source"
)

func dataSourceConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appId := d.Get(names.AttrApplicationID).(string)
	profileId := d.Get("configuration_profile_id").(string)
	ID := fmt.Sprintf("%s:%s", profileId, appId)

	out, err := findConfigurationProfileByApplicationAndProfile(ctx, conn, appId, profileId)
	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameConfigurationProfile, ID, err)
	}

	d.SetId(ID)

	d.Set(names.AttrApplicationID, appId)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s", appId, profileId),
		Service:   "appconfig",
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set("configuration_profile_id", profileId)
	d.Set(names.AttrDescription, out.Description)
	d.Set("kms_key_identifier", out.KmsKeyIdentifier)
	d.Set("location_uri", out.LocationUri)
	d.Set(names.AttrName, out.Name)
	d.Set("retrieval_role_arn", out.RetrievalRoleArn)
	d.Set(names.AttrType, out.Type)

	if err := d.Set("validator", flattenValidators(out.Validators)); err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionSetting, DSNameConfigurationProfile, ID, err)
	}

	return diags
}

func findConfigurationProfileByApplicationAndProfile(ctx context.Context, conn *appconfig.Client, appId string, cpId string) (*appconfig.GetConfigurationProfileOutput, error) {
	res, err := conn.GetConfigurationProfile(ctx, &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appId),
		ConfigurationProfileId: aws.String(cpId),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
