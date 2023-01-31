package amp

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAlertManagerDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAlertManagerDefinitionCreate,
		ReadWithoutTimeout:   resourceAlertManagerDefinitionRead,
		UpdateWithoutTimeout: resourceAlertManagerDefinitionUpdate,
		DeleteWithoutTimeout: resourceAlertManagerDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	conn := meta.(*conns.AWSClient).AMPConn()

	workspaceID := d.Get("workspace_id").(string)
	input := &prometheusservice.CreateAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	_, err := conn.CreateAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Prometheus Alert Manager Definition (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)

	if _, err := waitAlertManagerDefinitionCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Prometheus Alert Manager Definition (%s) create: %s", d.Id(), err)
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()

	amd, err := FindAlertManagerDefinitionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Alert Manager Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	d.Set("definition", string(amd.Data))
	d.Set("workspace_id", d.Id())

	return nil
}

func resourceAlertManagerDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()

	input := &prometheusservice.PutAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionUpdated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Prometheus Alert Manager Definition (%s) update: %s", d.Id(), err)
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()

	log.Printf("[DEBUG] Deleting Prometheus Alert Manager Definition: (%s)", d.Id())
	_, err := conn.DeleteAlertManagerDefinitionWithContext(ctx, &prometheusservice.DeleteAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Prometheus Alert Manager Definition (%s) delete: %s", d.Id(), err)
	}

	return nil
}
