// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_dedicated_ip_assignment")
func ResourceDedicatedIPAssignment() *schema.Resource {
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
	ResNameDedicatedIPAssignment = "Dedicated IP Assignment"
)

func resourceDedicatedIPAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.PutDedicatedIpInPoolInput{
		Ip:                  aws.String(d.Get("ip").(string)),
		DestinationPoolName: aws.String(d.Get("destination_pool_name").(string)),
	}

	_, err := conn.PutDedicatedIpInPool(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameDedicatedIPAssignment, d.Get("ip").(string), err)
	}

	id := toID(d.Get("ip").(string), d.Get("destination_pool_name").(string))
	d.SetId(id)

	return append(diags, resourceDedicatedIPAssignmentRead(ctx, d, meta)...)
}

func resourceDedicatedIPAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindDedicatedIPAssignmentByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 DedicatedIPAssignment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameDedicatedIPAssignment, d.Id(), err)
	}

	d.Set("ip", out.Ip)
	d.Set("destination_pool_name", out.PoolName)

	return diags
}

func resourceDedicatedIPAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)
	ip, _ := splitID(d.Id())

	log.Printf("[INFO] Deleting SESV2 DedicatedIPAssignment %s", d.Id())
	_, err := conn.PutDedicatedIpInPool(ctx, &sesv2.PutDedicatedIpInPoolInput{
		Ip:                  aws.String(ip),
		DestinationPoolName: aws.String(defaultDedicatedPoolName),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameDedicatedIPAssignment, d.Id(), err)
	}

	return diags
}

const (
	// defaultDedicatedPoolName contains the name of the standard pool managed by AWS
	// where dedicated IP addresses with an assignment are stored
	//
	// When an assignment resource is removed from state, the delete function will re-assign
	// the relevant IP to this pool.
	defaultDedicatedPoolName = "ses-default-dedicated-pool"
)

// ErrIncorrectPoolAssignment is returned when an IP is assigned to a pool different
// from what is specified in state
var ErrIncorrectPoolAssignment = errors.New("incorrect pool assignment")

func FindDedicatedIPAssignmentByID(ctx context.Context, conn *sesv2.Client, id string) (*types.DedicatedIp, error) {
	ip, destinationPoolName := splitID(id)

	in := &sesv2.GetDedicatedIpInput{
		Ip: aws.String(ip),
	}
	out, err := conn.GetDedicatedIp(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DedicatedIp == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	if out.DedicatedIp.PoolName == nil || aws.ToString(out.DedicatedIp.PoolName) != destinationPoolName {
		return nil, &retry.NotFoundError{
			LastError:   ErrIncorrectPoolAssignment,
			LastRequest: in,
		}
	}

	return out.DedicatedIp, nil
}

func toID(ip, destinationPoolName string) string {
	return fmt.Sprintf("%s,%s", ip, destinationPoolName)
}

func splitID(id string) (string, string) {
	items := strings.Split(id, ",")
	if len(items) == 2 {
		return items[0], items[1]
	}
	return "", ""
}
