package prometheus

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
)

func ResourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"alert_manager_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	details, err := conn.DescribeWorkspaceWithContext(ctx, &prometheusservice.DescribeWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Prometheus Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Workspace (%s): %w", d.Id(), err))
	}

	if details == nil || details.Workspace == nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Workspace (%s): empty response", d.Id()))
	}

	ws := details.Workspace

	d.Set("alias", ws.Alias)
	d.Set("arn", ws.Arn)
	d.Set("prometheus_endpoint", ws.PrometheusEndpoint)

	am, err := conn.DescribeAlertManagerDefinition(&prometheusservice.DescribeAlertManagerDefinitionInput{
		WorkspaceId: ws.WorkspaceId,
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
			return nil
		}
	}
	d.Set("alert_manager_definition", string(am.AlertManagerDefinition.Data))

	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Updating AMP workspace %s", d.Id())

	req := &prometheusservice.UpdateWorkspaceAliasInput{
		WorkspaceId: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("alias"); ok {
		req.Alias = aws.String(v.(string))
	}
	conn := meta.(*conns.AWSClient).PrometheusConn
	if _, err := conn.UpdateWorkspaceAliasWithContext(ctx, req); err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}

	if d.HasChange("alert_manager_definition") {
		if _, err := waitWorkspaceActive(ctx, conn, d.Id()); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for Workspace (%s) to be ACTIVE: %w", d.Id(), err))
		}

		value, ok := d.GetOk("alert_manager_definition")
		if !ok {
			req := &prometheusservice.DeleteAlertManagerDefinitionInput{
				WorkspaceId: aws.String(d.Id()),
			}
			if _, err := conn.DeleteAlertManagerDefinitionWithContext(ctx, req); err != nil {
				return diag.FromErr(fmt.Errorf("error deleting Prometheus AlertManager (%s): %w", d.Id(), err))
			}

			if _, err := waitAlertManagerDeleted(ctx, conn, d.Id()); err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for AlertManager (%s) to be DELETED: %w", d.Id(), err))
			}
		} else {
			_, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, &prometheusservice.DescribeAlertManagerDefinitionInput{
				WorkspaceId: aws.String(d.Id()),
			})
			// we have an existing definition - we must update it
			if err == nil {
				req := &prometheusservice.PutAlertManagerDefinitionInput{
					WorkspaceId: aws.String(d.Id()),
					Data:        []byte(value.(string)),
				}
				if _, err := conn.PutAlertManagerDefinitionWithContext(ctx, req); err != nil {
					return diag.FromErr(fmt.Errorf("error updating Prometheus AlertManager (%s): %w\n%s\n", d.Id(), err, value.(string)))
				}
				// else we have to create a definition
			} else {
				req := &prometheusservice.CreateAlertManagerDefinitionInput{
					WorkspaceId: aws.String(d.Id()),
					Data:        []byte(value.(string)),
				}
				if _, err := conn.CreateAlertManagerDefinitionWithContext(ctx, req); err != nil {
					return diag.FromErr(fmt.Errorf("error creating Prometheus AlertManager (%s): %w\n%s\n", d.Id(), err, value.(string)))
				}
			}

			if _, err := waitAlertManagerActive(ctx, conn, d.Id()); err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for AlertManager (%s) to be ACTIVE: %w", d.Id(), err))
			}
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Creating AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	req := &prometheusservice.CreateWorkspaceInput{}
	if v, ok := d.GetOk("alias"); ok {
		req.Alias = aws.String(v.(string))
	}

	result, err := conn.CreateWorkspaceWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}
	d.SetId(aws.StringValue(result.WorkspaceId))

	if _, err := waitWorkspaceActive(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Workspace (%s) to be created: %w", d.Id(), err))
	}

	if am, ok := d.GetOk("alert_manager_definition"); ok {
		_, err = conn.CreateAlertManagerDefinitionWithContext(ctx, &prometheusservice.CreateAlertManagerDefinitionInput{
			WorkspaceId: result.WorkspaceId,
			Data:        []byte(am.(string)),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to create AlertManagerDefinition for workspace (%s) to be created: %w", d.Id(), err))
		}

		if _, err := waitAlertManagerActive(ctx, conn, d.Id()); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for AlertManager (%s) to be created: %w", d.Id(), err))
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).PrometheusConn

	_, err := conn.DeleteWorkspaceWithContext(ctx, &prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Prometheus Workspace (%s): %w", d.Id(), err))
	}

	// AlertManager is deleted when the workspace is deleted, so there's no need for a separate
	// DELETE call
	if _, err := waitWorkspaceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Workspace (%s) to be deleted: %w", d.Id(), err))
	}

	return nil
}
