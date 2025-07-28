// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_dedicated_ip_assignment", name="Dedicated IP Assignment")
func resourceDedicatedIPAssignment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDedicatedIPAssignmentCreate,
		ReadWithoutTimeout:   resourceDedicatedIPAssignmentRead,
		DeleteWithoutTimeout: resourceDedicatedIPAssignmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"destination_pool_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	}
}

const (
	resNameDedicatedIPAssignment = "Dedicated IP Assignment"
)

func resourceDedicatedIPAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	ip, destinationPoolName := d.Get("ip").(string), d.Get("destination_pool_name").(string)
	id := dedicatedIPAssignmentCreateResourceID(ip, destinationPoolName)
	input := &sesv2.PutDedicatedIpInPoolInput{
		DestinationPoolName: aws.String(destinationPoolName),
		Ip:                  aws.String(ip),
	}

	_, err := conn.PutDedicatedIpInPool(ctx, input)

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameDedicatedIPAssignment, d.Get("ip").(string), err)
	}

	d.SetId(id)

	return append(diags, resourceDedicatedIPAssignmentRead(ctx, d, meta)...)
}

func resourceDedicatedIPAssignmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	ip, destinationPoolName, err := dedicatedIPAssignmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findDedicatedIPByTwoPartKey(ctx, conn, ip, destinationPoolName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 DedicatedIPAssignment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameDedicatedIPAssignment, d.Id(), err)
	}

	d.Set("destination_pool_name", out.PoolName)
	d.Set("ip", out.Ip)

	return diags
}

func resourceDedicatedIPAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	ip, _, err := dedicatedIPAssignmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting SESV2 DedicatedIPAssignment %s", d.Id())
	const (
		// defaultDedicatedPoolName contains the name of the standard pool managed by AWS
		// where dedicated IP addresses with an assignment are stored
		//
		// When an assignment resource is removed from state, the delete function will re-assign
		// the relevant IP to this pool.
		defaultDedicatedPoolName = "ses-default-dedicated-pool"
	)
	_, err = conn.PutDedicatedIpInPool(ctx, &sesv2.PutDedicatedIpInPoolInput{
		Ip:                  aws.String(ip),
		DestinationPoolName: aws.String(defaultDedicatedPoolName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameDedicatedIPAssignment, d.Id(), err)
	}

	return diags
}

const dedicatedIPAssignmentResourceIDSeparator = ","

func dedicatedIPAssignmentCreateResourceID(ip, destinationPoolName string) string {
	parts := []string{ip, destinationPoolName}
	id := strings.Join(parts, dedicatedIPAssignmentResourceIDSeparator)

	return id
}

func dedicatedIPAssignmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, dedicatedIPAssignmentResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected IP%[2]sDESTINATION_POOL_NAME", id, dedicatedIPAssignmentResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDedicatedIPByTwoPartKey(ctx context.Context, conn *sesv2.Client, ip, destinationPoolName string) (*types.DedicatedIp, error) {
	input := &sesv2.GetDedicatedIpInput{
		Ip: aws.String(ip),
	}

	output, err := findDedicatedIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if aws.ToString(output.PoolName) != destinationPoolName {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDedicatedIP(ctx context.Context, conn *sesv2.Client, input *sesv2.GetDedicatedIpInput) (*types.DedicatedIp, error) {
	output, err := conn.GetDedicatedIp(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DedicatedIp == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DedicatedIp, nil
}
