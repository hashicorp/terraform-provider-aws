package controltower

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceControlCreate,
		ReadContext:   resourceControlRead,
		UpdateContext: resourceControlUpdate,
		DeleteContext: resourceControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"control_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn

	control_identifier := d.Get("control_identifier").(string)
	target_identifier := d.Get("target_identifier").(string)

	input := &controltower.EnableControlInput{
		ControlIdentifier: aws.String(control_identifier),
		TargetIdentifier:  aws.String(target_identifier),
	}

	log.Printf("[DEBUG] Enabling ControlTower Control: %#v", input)
	output, err := conn.EnableControlWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("Enabling ControlTower Control (%s): %s", control_identifier, err)
	}

	if _, err := waitControlCreated(ctx, conn, *output.OperationIdentifier); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ControlTower Control (%s) to be created: %w", d.Id(), err))
	}

	d.SetId(target_identifier)
	d.Set("control_identifier", control_identifier)

	return resourceControlRead(ctx, d, meta)
}

func resourceControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn

	log.Printf("[DEBUG] Reading ControlTower Control %s", d.Id())

	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: aws.String(d.Id()),
	}
	output, err := conn.ListEnabledControlsWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading ControlTower Control (%s): %w", d.Id(), err))
	}

	for _, c := range output.EnabledControls {
		if c == d.Get("control_identifier") {
			return nil
		}
	}

	// control identifier not found in response
	log.Printf("[WARN] ControlTower Control (%s) not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func resourceControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// conn := meta.(*conns.AWSClient).ControlTowerConn

	// control_identifier := d.Get("control_identifier").(string)
	// target_identifier := d.Get("target_identifier").(string)

	// input := &controltower.DisableControlInput{
	// 	ControlIdentifier: aws.String(control_identifier),
	// 	TargetIdentifier:  aws.String(target_identifier),
	// }

	// log.Printf("[DEBUG] Enabling ControlTower Control: %#v", input)
	// output, err := conn.DisableControlWithContext(ctx, input)

	// if err != nil {
	// 	return diag.Errorf("Enabling ControlTower Control (%s): %s", control_identifier, err)
	// }

	// d.SetId(aws.StringValue(target_identifier))

	// return resourceProfileRead(ctx, d, meta)
	return nil
}

func resourceControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// conn := meta.(*conns.AWSClient).ControlTowerConn
	return nil
}
