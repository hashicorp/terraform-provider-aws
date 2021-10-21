package prometheus

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceAlertManagerDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] AMP alert manager definition %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	req := &prometheusservice.CreateAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.CreateAlertManagerDefinitionWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus alert manager definition (%s): %w", d.Id(), err))
	}
	d.SetId(aws.StringValue(req.WorkspaceId))

	if _, err := waitAlertManagerDefinitionCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for alert manager definition (%s) to be created: %w", d.Id(), err))
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] AMP alert manager definition %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	req := &prometheusservice.PutAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutAlertManagerDefinitionWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus alert manager definition (%s): %w", d.Id(), err))
	}
	d.SetId(aws.StringValue(req.WorkspaceId))

	if _, err := waitAlertManagerDefinitionUpdated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for alert manager definition (%s) to be updated: %w", d.Id(), err))
	}

	return resourceAlertManagerDefinitionRead(ctx, d, meta)
}

func resourceAlertManagerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading AMP alert manager definition %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	details, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, &prometheusservice.DescribeAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Prometheus Alert Manager Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Alert Manager Definition (%s): %w", d.Id(), err))
	}

	if details == nil || details.AlertManagerDefinition == nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Alert Manager Definition (%s): empty response", d.Id()))
	}

	amd := details.AlertManagerDefinition

	d.Set("definition", string(amd.Data))
	d.Set("workspace_id", d.Id())

	return nil
}

func resourceAlertManagerDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting AMP alert manager definition %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	_, err := conn.DeleteAlertManagerDefinitionWithContext(ctx, &prometheusservice.DeleteAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Prometheus alert manager definition (%s): %w", d.Id(), err))
	}

	if _, err := waitAlertManagerDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for alert manager definition (%s) to be deleted: %w", d.Id(), err))
	}

	return nil
}

func waitAlertManagerDefinitionCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeCreating},
		Target:  []string{prometheusservice.AlertManagerDefinitionStatusCodeActive},
		Refresh: statusAlertManagerDefinitionCreated(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return v, err
	}

	return nil, err
}

func waitAlertManagerDefinitionUpdated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeUpdating},
		Target:  []string{prometheusservice.AlertManagerDefinitionStatusCodeActive},
		Refresh: statusAlertManagerDefinitionCreated(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return v, err
	}

	return nil, err
}

func waitAlertManagerDefinitionDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusAlertManagerDefinitionDeleted(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return v, err
	}

	return nil, err
}

func statusAlertManagerDefinitionCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeAlertManagerDefinitionInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, input)

		if err != nil {
			return output, resourceStatusFailed, err
		}

		if output == nil || output.AlertManagerDefinition == nil || output.AlertManagerDefinition.Status == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.AlertManagerDefinition, aws.StringValue(output.AlertManagerDefinition.Status.StatusCode), nil
	}
}

func statusAlertManagerDefinitionDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeAlertManagerDefinitionInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return output, resourceStatusUnknown, err
		}

		if output == nil || output.AlertManagerDefinition == nil || output.AlertManagerDefinition.Status == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.AlertManagerDefinition, aws.StringValue(output.AlertManagerDefinition.Status.StatusCode), nil
	}
}
