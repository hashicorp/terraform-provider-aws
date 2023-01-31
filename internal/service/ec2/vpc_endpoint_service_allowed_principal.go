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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCEndpointServiceAllowedPrincipal() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointServiceAllowedPrincipalCreate,
		ReadWithoutTimeout:   resourceVPCEndpointServiceAllowedPrincipalRead,
		DeleteWithoutTimeout: resourceVPCEndpointServiceAllowedPrincipalDelete,

		Schema: map[string]*schema.Schema{
			"principal_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointServiceAllowedPrincipalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	principalARN := d.Get("principal_arn").(string)

	_, err := conn.ModifyVpcEndpointServicePermissionsWithContext(ctx, &ec2.ModifyVpcEndpointServicePermissionsInput{
		AddAllowedPrincipals: aws.StringSlice([]string{principalARN}),
		ServiceId:            aws.String(serviceID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying EC2 VPC Endpoint Service (%s) permissions: %s", serviceID, err)
	}

	d.SetId(fmt.Sprintf("a-%s%d", serviceID, create.StringHashcode(principalARN)))

	return append(diags, resourceVPCEndpointServiceAllowedPrincipalRead(ctx, d, meta)...)
}

func resourceVPCEndpointServiceAllowedPrincipalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	principalARN := d.Get("principal_arn").(string)

	err := FindVPCEndpointServicePermissionExists(ctx, conn, serviceID, principalARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Endpoint Service Allowed Principal %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoint Service (%s) Allowed Principal (%s): %s", serviceID, principalARN, err)
	}

	return diags
}

func resourceVPCEndpointServiceAllowedPrincipalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	principalARN := d.Get("principal_arn").(string)

	_, err := conn.ModifyVpcEndpointServicePermissionsWithContext(ctx, &ec2.ModifyVpcEndpointServicePermissionsInput{
		RemoveAllowedPrincipals: aws.StringSlice([]string{principalARN}),
		ServiceId:               aws.String(serviceID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying EC2 VPC Endpoint Service (%s) permissions: %s", serviceID, err)
	}

	return diags
}
