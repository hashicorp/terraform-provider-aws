package datapipeline

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourcePipelineDefinition() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePipelineDefinitionRead,

		Schema: map[string]*schema.Schema{
			"parameter_object": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"string_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"parameter_value": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"string_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"pipeline_object": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ref_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"string_value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePipelineDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DataPipelineConn

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(pipelineID),
	}

	resp, err := conn.GetPipelineDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error getting DataPipeline Definition (%s): %s", pipelineID, err)
	}

	if err = d.Set("parameter_object", flattenPipelineDefinitionParameterObjects(resp.ParameterObjects)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	if err = d.Set("parameter_value", flattenPipelineDefinitionParameterValues(resp.ParameterValues)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	if err = d.Set("pipeline_object", flattenPipelineDefinitionObjects(resp.PipelineObjects)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", pipelineID, err)
	}
	d.SetId(pipelineID)

	return nil
}
