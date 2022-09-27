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
		DeleteContext: resourceControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"control_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
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

	d.SetId(target_identifier)

	if _, err := waitControl(ctx, conn, *output.OperationIdentifier); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ControlTower Control (%s) to be created: %w", d.Id(), err))
	}

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
		if aws.StringValue(c.ControlIdentifier) == d.Get("control_identifier") {
			return nil
		}
	}

	// control identifier not found in response
	if !d.IsNewResource() {
		log.Printf("[WARN] ControlTower Control (%s) not found, removing from state", d.Id())
		d.SetId("")
	}
	return nil
}

func resourceControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn

	log.Printf("[DEBUG] Disabling ControlTower Control %s", d.Id())

	control_identifier := d.Get("control_identifier").(string)
	target_identifier := d.Get("target_identifier").(string)

	input := &controltower.DisableControlInput{
		ControlIdentifier: aws.String(control_identifier),
		TargetIdentifier:  aws.String(target_identifier),
	}

	output, err := conn.DisableControlWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error disabling ControlTower Control (%s): %s", control_identifier, err)
	}

	if _, err := waitControl(ctx, conn, *output.OperationIdentifier); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ControlTower Control (%s) to disable: %w", d.Id(), err))
	}

	return nil
}
