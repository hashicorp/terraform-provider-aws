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

func DataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_arn": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"alarm_role_arn": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameEnvironment = "Environment Data Source"
)

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn()

	appID := d.Get("application_id").(string)
	envID := d.Get("environment_id").(string)
	ID := fmt.Sprintf("%s:%s", envID, appID)

	out, err := findEnvironmentByApplicationAndEnvironment(ctx, conn, appID, envID)
	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	d.SetId(ID)

	d.Set("application_id", appID)
	d.Set("environment_id", envID)
	d.Set("description", out.Description)
	d.Set("name", out.Name)
	d.Set("state", out.State)

	if err := d.Set("monitor", flattenEnvironmentMonitors(out.Monitors)); err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/environment/%s", appID, envID),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionSetting, DSNameEnvironment, ID, err)
	}

	return nil
}

func findEnvironmentByApplicationAndEnvironment(ctx context.Context, conn *appconfig.AppConfig, appId string, envId string) (*appconfig.GetEnvironmentOutput, error) {
	res, err := conn.GetEnvironmentWithContext(ctx, &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appId),
		EnvironmentId: aws.String(envId),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
