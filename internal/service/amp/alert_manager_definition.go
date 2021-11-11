package amp

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAlertManagerDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertManagerDefinitionCreate,
		ReadContext:   resourceAlertManagerDefinitionRead,
		UpdateContext: resourceAlertManagerDefinitionUpdate,
		DeleteContext: resourceAlertManagerDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAlertManagerDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	workspaceID := d.Get("workspace_id").(string)
	input := &prometheusservice.CreateAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Creating Prometheus Alert Manager Definition: %s", input)
	_, err := conn.CreateAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus Alert Manager Definition (%s): %w", workspaceID, err))
	}

	d.SetId(workspaceID)

	if _, err := waitAlertManagerDefinitionCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Alert Manager Definition (%s) create: %w", d.Id(), err))
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	input := &prometheusservice.PutAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Updating Prometheus Alert Manager Definition: %s", input)
	_, err := conn.PutAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus Alert Manager Definition (%s): %w", d.Id(), err))
	}

	if _, err := waitAlertManagerDefinitionUpdated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Alert Manager Definition (%s) update: %w", d.Id(), err))
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	amd, err := FindAlertManagerDefinitionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Alert Manager Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Alert Manager Definition (%s): %w", d.Id(), err))
	}

	d.Set("definition", string(amd.Data))
	d.Set("workspace_id", d.Id())

	return nil
}

func resourceAlertManagerDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	log.Printf("[DEBUG] Deleting Prometheus Alert Manager Definition: (%s)", d.Id())
	_, err := conn.DeleteAlertManagerDefinitionWithContext(ctx, &prometheusservice.DeleteAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Prometheus Alert Manager Definition (%s): %w", d.Id(), err))
	}

	if _, err := waitAlertManagerDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Alert Manager Definition (%s) delete: %w", d.Id(), err))
	}

	return nil
}
