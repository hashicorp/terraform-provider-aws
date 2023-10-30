package batch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceComputeEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{aws.String(d.Get("compute_environment_name").(string))},
	}
	desc, err := conn.DescribeComputeEnvironmentsWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): %s", d.Get("compute_environment_name").(string), err)
	}

	if l := len(desc.ComputeEnvironments); l == 0 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): empty response", d.Get("compute_environment_name").(string))
	} else if l > 1 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): too many results: wanted 1, got %d", d.Get("compute_environment_name").(string), l)
	}

	computeEnvironment := desc.ComputeEnvironments[0]
	d.SetId(aws.StringValue(computeEnvironment.ComputeEnvironmentArn))
	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("type", computeEnvironment.Type)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("state", computeEnvironment.State)

	if err := d.Set("tags", KeyValueTags(computeEnvironment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
