// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_endpoint_access", name="Endpoint Access")
func resourceEndpointAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointAccessCreate,
		ReadWithoutTimeout:   resourceEndpointAccessRead,
		UpdateWithoutTimeout: resourceEndpointAccessUpdate,
		DeleteWithoutTimeout: resourceEndpointAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 30),
			},
			"owner_account": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_interface": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAvailabilityZone: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrNetworkInterfaceID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"private_ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrSubnetID: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrVPCEndpointID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEndpointAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	endpointName := d.Get("endpoint_name").(string)
	input := &redshiftserverless.CreateEndpointAccessInput{
		EndpointName:  aws.String(endpointName),
		WorkgroupName: aws.String(d.Get("workgroup_name").(string)),
	}

	if v, ok := d.GetOk("owner_account"); ok {
		input.OwnerAccount = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateEndpointAccessWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Endpoint Access (%s): %s", endpointName, err)
	}

	d.SetId(aws.StringValue(output.Endpoint.EndpointName))

	if _, err := waitEndpointAccessActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Endpoint Access (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointAccessRead(ctx, d, meta)...)
}

func resourceEndpointAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	endpointAccess, err := findEndpointAccessByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Endpoint Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Endpoint Access (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAddress, endpointAccess.Address)
	d.Set(names.AttrARN, endpointAccess.EndpointArn)
	d.Set("endpoint_name", endpointAccess.EndpointName)
	d.Set("owner_account", d.Get("owner_account"))
	d.Set(names.AttrPort, endpointAccess.Port)
	d.Set(names.AttrSubnetIDs, aws.StringValueSlice(endpointAccess.SubnetIds))
	if err := d.Set("vpc_endpoint", []interface{}{flattenVPCEndpoint(endpointAccess.VpcEndpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_endpoint: %s", err)
	}
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(endpointAccess.VpcSecurityGroups, func(v *redshiftserverless.VpcSecurityGroupMembership) string {
		return aws.StringValue(v.VpcSecurityGroupId)
	}))
	d.Set("workgroup_name", endpointAccess.WorkgroupName)

	return diags
}

func resourceEndpointAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	input := &redshiftserverless.UpdateEndpointAccessInput{
		EndpointName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.UpdateEndpointAccessWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Endpoint Access (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointAccessActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Endpoint Access (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointAccessRead(ctx, d, meta)...)
}

func resourceEndpointAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Endpoint Access: %s", d.Id())
	_, err := conn.DeleteEndpointAccessWithContext(ctx, &redshiftserverless.DeleteEndpointAccessInput{
		EndpointName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Endpoint Access (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointAccessDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Endpoint Access (%s) delete: %s", d.Id(), err)
	}

	return diags
}
