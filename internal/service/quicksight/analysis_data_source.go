package quicksight

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	schemahelper "github.com/hashicorp/terraform-provider-aws/internal/schema"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_analysis", name="Analysis")
func DataSourceAnalysis() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAnalysisRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"analysis_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"definition": schemahelper.DataSourcePropertyFromResourceProperty(quicksightschema.AnalysisDefinitionSchema()),
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_published_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": schemahelper.DataSourcePropertyFromResourceProperty(quicksightschema.ParametersSchema()),
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"principal": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"theme_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameAnalysis = "Analysis Data Source"
)

func dataSourceAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}
	analysisId := d.Get("analysis_id").(string)

	id := createAnalysisId(awsAccountId, analysisId)

	out, err := FindAnalysisByID(ctx, conn, id)

	if err != nil {
		return create.DiagError(names.QuickSight, create.ErrActionReading, ResNameAnalysis, d.Id(), err)
	}

	d.SetId(id)
	d.Set("arn", out.Arn)
	d.Set("aws_account_id", awsAccountId)
	d.Set("created_time", out.CreatedTime.Format(time.RFC3339))
	d.Set("last_updated_time", out.LastUpdatedTime.Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("status", out.Status)
	d.Set("analysis_id", out.AnalysisId)
	d.Set("theme_arn", out.ThemeArn)

	descResp, err := conn.DescribeAnalysisDefinitionWithContext(ctx, &quicksight.DescribeAnalysisDefinitionInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	})

	if err != nil {
		return diag.Errorf("describing QuickSight Analysis (%s) Definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenAnalysisDefinition(descResp.Definition)); err != nil {
		return diag.Errorf("setting definition: %s", err)
	}

	permsResp, err := conn.DescribeAnalysisPermissionsWithContext(ctx, &quicksight.DescribeAnalysisPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	})

	if err != nil {
		return diag.Errorf("describing QuickSight Analysis (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set("permissions", flattenPermissions(permsResp.Permissions)); err != nil {
		return diag.Errorf("setting permissions: %s", err)
	}

	return nil
}
