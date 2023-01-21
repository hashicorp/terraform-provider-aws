package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocalGatewayRouteTableVPCAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationCreate,
		ReadWithoutTimeout:   resourceLocalGatewayRouteTableVPCAssociationRead,
		UpdateWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationUpdate,
		DeleteWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"local_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLocalGatewayRouteTableVPCAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &ec2.CreateLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableId: aws.String(d.Get("local_gateway_route_table_id").(string)),
		TagSpecifications:        tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeLocalGatewayRouteTableVpcAssociation),
		VpcId:                    aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateLocalGatewayRouteTableVpcAssociationWithContext(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Local Gateway Route Table VPC Association: %s", err)
	}

	d.SetId(aws.StringValue(output.LocalGatewayRouteTableVpcAssociation.LocalGatewayRouteTableVpcAssociationId))

	if _, err := WaitLocalGatewayRouteTableVPCAssociationAssociated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Local Gateway Route Table VPC Association (%s) to associate: %s", d.Id(), err)
	}

	return append(diags, resourceLocalGatewayRouteTableVPCAssociationRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteTableVPCAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	association, err := GetLocalGatewayRouteTableVPCAssociation(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Route Table VPC Association (%s): %s", d.Id(), err)
	}

	if association == nil {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeAssociated {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) status (%s), removing from state", d.Id(), aws.StringValue(association.State))
		d.SetId("")
		return diags
	}

	d.Set("local_gateway_id", association.LocalGatewayId)
	d.Set("local_gateway_route_table_id", association.LocalGatewayRouteTableId)

	tags := KeyValueTags(association.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	d.Set("vpc_id", association.VpcId)

	return diags
}

func resourceLocalGatewayRouteTableVPCAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Local Gateway Route Table VPC Association (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocalGatewayRouteTableVPCAssociationRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteTableVPCAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableVpcAssociationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteLocalGatewayRouteTableVpcAssociationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, "InvalidLocalGatewayRouteTableVpcAssociationID.NotFound") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Local Gateway Route Table VPC Association (%s): %s", d.Id(), err)
	}

	if _, err := WaitLocalGatewayRouteTableVPCAssociationDisassociated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Local Gateway Route Table VPC Association (%s) to disassociate: %s", d.Id(), err)
	}

	return diags
}

func GetLocalGatewayRouteTableVPCAssociation(ctx context.Context, conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
		LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
	}

	output, err := conn.DescribeLocalGatewayRouteTableVpcAssociationsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, fmt.Errorf("empty response")
	}

	var association *ec2.LocalGatewayRouteTableVpcAssociation

	for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
		if outputAssociation == nil {
			continue
		}

		if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
			association = outputAssociation
			break
		}
	}

	return association, nil
}
