package appconfig

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentsRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"environment_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameEnvironments = "Environments Data Source"
)

func dataSourceEnvironmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn()
	appID := d.Get("application_id").(string)

	out, err := findEnvironmentsByApplication(ctx, conn, appID)
	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, DSNameEnvironments, appID, err)
	}

	d.SetId(appID)

	var environmentIds []*string
	for _, v := range out {
		environmentIds = append(environmentIds, v.Id)
	}
	d.Set("environment_ids", aws.StringValueSlice(environmentIds))

	return nil
}

func findEnvironmentsByApplication(ctx context.Context, conn *appconfig.AppConfig, appId string) ([]*appconfig.Environment, error) {
	var outputs []*appconfig.Environment
	err := conn.ListEnvironmentsPagesWithContext(ctx, &appconfig.ListEnvironmentsInput{
		ApplicationId: aws.String(appId),
	}, func(output *appconfig.ListEnvironmentsOutput, lastPage bool) bool {
		outputs = append(outputs, output.Items...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return outputs, nil
}
