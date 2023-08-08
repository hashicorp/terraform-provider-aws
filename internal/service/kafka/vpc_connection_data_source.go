package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_vpc_connection", name="Vpc Connection")
func DataSourceVpcConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVpcConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"authentication": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"client_subnets": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameVPCConnection = "VPC Connection Data Source"
)

func dataSourceVpcConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	arn := d.Get("arn").(string)
	out, err := FindVPCConnectionByARN(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameVPCConnection, arn, err)
	}

	d.SetId(aws.ToString(out.VpcConnectionArn))

	d.Set("arn", out.VpcConnectionArn)
	d.Set("authentication", out.Authentication)
	d.Set("vpc_id", out.VpcId)
	d.Set("target_cluster_arn", out.TargetClusterArn)

	if err := d.Set("client_subnets", flex.FlattenStringValueSet(out.Subnets)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionSetting, ResNameVPCConnection, d.Id(), err)...)
	}
	if err := d.Set("security_groups", flex.FlattenStringValueSet(out.SecurityGroups)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionSetting, ResNameVPCConnection, d.Id(), err)...)
	}

	return nil
}
