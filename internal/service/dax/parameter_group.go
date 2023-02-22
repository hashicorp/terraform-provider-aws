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
)

func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

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
				ForceNew: true,
			},
			"parameters": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.CreateParameterGroupInput{
		ParameterGroupName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateParameterGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DAX Parameter Group (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(d.Get("name").(string))

	if len(d.Get("parameters").(*schema.Set).List()) > 0 {
		return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
	}
	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	resp, err := conn.DescribeParameterGroupsWithContext(ctx, &dax.DescribeParameterGroupsInput{
		ParameterGroupNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dax.ErrCodeParameterGroupNotFoundFault) {
			log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DAX Parameter Group (%s): %s", d.Id(), err)
	}

	if len(resp.ParameterGroups) == 0 {
		log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	pg := resp.ParameterGroups[0]

	paramresp, err := conn.DescribeParametersWithContext(ctx, &dax.DescribeParametersInput{
		ParameterGroupName: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dax.ErrCodeParameterGroupNotFoundFault) {
			log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DAX Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set("name", pg.ParameterGroupName)
	desc := pg.Description
	// default description is " "
	if desc != nil && aws.StringValue(desc) == " " {
		*desc = ""
	}
	d.Set("description", desc)
	d.Set("parameters", flattenParameterGroupParameters(paramresp.Parameters))
	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.UpdateParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	if d.HasChange("parameters") {
		input.ParameterNameValues = expandParameterGroupParameterNameValue(
			d.Get("parameters").(*schema.Set).List(),
		)
	}

	_, err := conn.UpdateParameterGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DAX Parameter Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXConn()

	input := &dax.DeleteParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteParameterGroupWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dax.ErrCodeParameterGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DAX Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}
