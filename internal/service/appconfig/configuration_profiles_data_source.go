package appconfig

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceConfigurationProfiles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationProfilesRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration_profile_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameConfigurationProfiles = "Configuration Profiles Data Source"
)

func dataSourceConfigurationProfilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn()
	appId := d.Get("application_id").(string)

	out, err := findConfigurationProfileSummariesByApplication(ctx, conn, appId)
	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameConfigurationProfiles, appId, err)
	}

	d.SetId(appId)

	var configIds []*string
	for _, v := range out {
		configIds = append(configIds, v.Id)
	}

	d.Set("configuration_profile_ids", aws.StringValueSlice(configIds))

	return nil
}

func findConfigurationProfileSummariesByApplication(ctx context.Context, conn *appconfig.AppConfig, applicationId string) ([]*appconfig.ConfigurationProfileSummary, error) {
	var outputs []*appconfig.ConfigurationProfileSummary
	err := conn.ListConfigurationProfilesPagesWithContext(ctx, &appconfig.ListConfigurationProfilesInput{
		ApplicationId: aws.String(applicationId),
	}, func(output *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
		outputs = append(outputs, output.Items...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return outputs, nil
}
