// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_instance_state")
func ResourceInstanceState() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceStateCreate,
		ReadWithoutTimeout:   resourceInstanceStateRead,
		UpdateWithoutTimeout: resourceInstanceStateUpdate,
		DeleteWithoutTimeout: resourceInstanceStateDelete,

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
			"instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"state": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{ec2.InstanceStateNameRunning, ec2.InstanceStateNameStopped}, false),
			},
		},
	}
}

func resourceInstanceStateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	instanceId := d.Get("instance_id").(string)

	instance, instanceErr := waitInstanceReady(ctx, conn, instanceId, d.Timeout(schema.TimeoutCreate))

	if instanceErr != nil {
		return create.AppendDiagError(diags, names.EC2, create.ErrActionReading, ResInstance, instanceId, instanceErr)
	}

	err := updateInstanceState(ctx, conn, instanceId, aws.StringValue(instance.State.Name), d.Get("state").(string), d.Get("force").(bool))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(d.Get("instance_id").(string))

	return append(diags, resourceInstanceStateRead(ctx, d, meta)...)
}

func resourceInstanceStateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	state, err := FindInstanceStateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.EC2, create.ErrActionReading, ResInstanceState, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.EC2, create.ErrActionReading, ResInstanceState, d.Id(), err)
	}

	d.Set("instance_id", d.Id())
	d.Set("state", state.Name)
	d.Set("force", d.Get("force").(bool))

	return diags
}

func resourceInstanceStateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	instance, instanceErr := waitInstanceReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if instanceErr != nil {
		return create.AppendDiagError(diags, names.EC2, create.ErrActionReading, ResInstance, aws.StringValue(instance.InstanceId), instanceErr)
	}

	if d.HasChange("state") {
		o, n := d.GetChange("state")
		err := updateInstanceState(ctx, conn, d.Id(), o.(string), n.(string), d.Get("force").(bool))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceInstanceStateRead(ctx, d, meta)...)
}

func resourceInstanceStateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] %s %s deleting an aws_ec2_instance_state resource only stops managing instance state, The Instance is left in its current state.: %s", names.EC2, ResInstanceState, d.Id())

	return nil // nosemgrep:ci.semgrep.pluginsdk.return-diags-not-nil
}

func updateInstanceState(ctx context.Context, conn *ec2.EC2, id string, currentState string, configuredState string, force bool) error {
	if currentState == configuredState {
		return nil
	}

	if configuredState == "stopped" {
		if err := stopInstance(ctx, conn, id, force, InstanceStopTimeout); err != nil {
			return err
		}
	}

	if configuredState == "running" {
		if err := startInstance(ctx, conn, id, InstanceStartTimeout); err != nil {
			return err
		}
	}

	return nil
}
