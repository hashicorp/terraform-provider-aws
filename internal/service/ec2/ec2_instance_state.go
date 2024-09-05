// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_instance_state", name="Instance State")
func resourceInstanceState() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceStateCreate,
		ReadWithoutTimeout:   resourceInstanceStateRead,
		UpdateWithoutTimeout: resourceInstanceStateUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrState: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.InstanceStateNameRunning, awstypes.InstanceStateNameStopped), false),
			},
		},
	}
}

func resourceInstanceStateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	instance, err := waitInstanceReady(ctx, conn, instanceID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) ready: %s", instanceID, err)
	}

	if err := updateInstanceState(ctx, conn, instanceID, string(instance.State.Name), d.Get(names.AttrState).(string), d.Get("force").(bool)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(instanceID)

	return append(diags, resourceInstanceStateRead(ctx, d, meta)...)
}

func resourceInstanceStateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	state, err := findInstanceStateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Instance State %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance State (%s): %s", d.Id(), err)
	}

	d.Set("force", d.Get("force").(bool))
	d.Set(names.AttrInstanceID, d.Id())
	d.Set(names.AttrState, state.Name)

	return diags
}

func resourceInstanceStateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if _, err := waitInstanceReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) ready: %s", d.Id(), err)
	}

	if d.HasChange(names.AttrState) {
		o, n := d.GetChange(names.AttrState)

		if err := updateInstanceState(ctx, conn, d.Id(), o.(string), n.(string), d.Get("force").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceInstanceStateRead(ctx, d, meta)...)
}

func updateInstanceState(ctx context.Context, conn *ec2.Client, id string, currentState string, configuredState string, force bool) error {
	if currentState == configuredState {
		return nil
	}

	if configuredState == "stopped" {
		if err := stopInstance(ctx, conn, id, force, instanceStopTimeout); err != nil {
			return err
		}
	}

	if configuredState == "running" {
		if err := startInstance(ctx, conn, id, false, instanceStartTimeout); err != nil {
			return err
		}
	}

	return nil
}
