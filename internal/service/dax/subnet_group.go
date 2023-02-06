package dax

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetGroupCreate,
		ReadWithoutTimeout:   resourceSubnetGroupRead,
		UpdateWithoutTimeout: resourceSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceSubnetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.CreateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Get("name").(string)),
		SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateSubnetGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DAX Subnet Group (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(d.Get("name").(string))
	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	resp, err := conn.DescribeSubnetGroupsWithContext(ctx, &dax.DescribeSubnetGroupsInput{
		SubnetGroupNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dax.ErrCodeSubnetGroupNotFoundFault) {
			log.Printf("[WARN] DAX Subnet Group %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DAX Subnet Group (%s): %s", d.Id(), err)
	}
	sg := resp.SubnetGroups[0]

	d.Set("name", sg.SubnetGroupName)
	d.Set("description", sg.Description)
	subnetIDs := make([]*string, 0, len(sg.Subnets))
	for _, v := range sg.Subnets {
		subnetIDs = append(subnetIDs, v.SubnetIdentifier)
	}
	d.Set("subnet_ids", flex.FlattenStringList(subnetIDs))
	d.Set("vpc_id", sg.VpcId)
	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.UpdateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("subnet_ids") {
		input.SubnetIds = flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set))
	}

	_, err := conn.UpdateSubnetGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DAX Subnet Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.DeleteSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteSubnetGroupWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dax.ErrCodeSubnetGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DAX Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}
