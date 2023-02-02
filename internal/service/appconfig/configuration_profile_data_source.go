package appconfig

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationProfileRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"retrieval_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"validator": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
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
	conn := meta.(*conns.AWSClient).AppConfigConn()

	appId := d.Get("application_id").(string)
	profileId := d.Get("configuration_profile_id").(string)
	ID := fmt.Sprintf("%s:%s", profileId, appId)

	out, err := findConfigurationProfileByApplicationAndProfile(ctx, conn, appId, profileId)
	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameConfigurationProfile, ID, err)
	}

	d.SetId(ID)

	d.Set("application_id", appId)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s", appId, profileId),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)
	d.Set("configuration_profile_id", profileId)
	d.Set("description", out.Description)
	d.Set("location_uri", out.LocationUri)
	d.Set("name", out.Name)
	d.Set("retrieval_role_arn", out.RetrievalRoleArn)
	d.Set("type", out.Type)

	if err := d.Set("validator", flattenValidators(out.Validators)); err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionSetting, DSNameConfigurationProfile, ID, err)
	}

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameConfigurationProfile, ID, err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionSetting, DSNameConfigurationProfile, ID, err)
	}

	return nil
}

func findConfigurationProfileByApplicationAndProfile(ctx context.Context, conn *appconfig.AppConfig, appId string, cpId string) (*appconfig.GetConfigurationProfileOutput, error) {
	res, err := conn.GetConfigurationProfileWithContext(ctx, &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appId),
		ConfigurationProfileId: aws.String(cpId),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
