package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/prometheusservice/waiter"
)

func resourceAwsPrometheusWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsPrometheusWorkspaceCreate,
		ReadContext:   resourceAwsPrometheusWorkspaceRead,
		UpdateContext: resourceAwsPrometheusWorkspaceUpdate,
		DeleteContext: resourceAwsPrometheusWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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

func resourceAwsPrometheusWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusserviceconn

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

	if details.Workspace != nil {
		ws := details.Workspace
		if ws.Arn != nil {
			d.Set("arn", ws.Arn)
		}
		if ws.PrometheusEndpoint != nil {
			d.Set("prometheus_endpoint", ws.PrometheusEndpoint)
		}
		if ws.Alias != nil {
			d.Set("alias", ws.Alias)
		}
	}

	return nil
}

func resourceAwsPrometheusWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Updating AMP workspace %s", d.Id())

	req := &prometheusservice.UpdateWorkspaceAliasInput{
		WorkspaceId: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("alias"); ok {
		req.Alias = aws.String(v.(string))
	}
	conn := meta.(*AWSClient).prometheusserviceconn
	if _, err := conn.UpdateWorkspaceAliasWithContext(ctx, req); err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}

	return resourceAwsPrometheusWorkspaceRead(ctx, d, meta)
}

func resourceAwsPrometheusWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Creating AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusserviceconn

	req := &prometheusservice.CreateWorkspaceInput{}
	if v, ok := d.GetOk("alias"); ok {
		req.Alias = aws.String(v.(string))
	}

	result, err := conn.CreateWorkspaceWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}
	d.SetId(aws.StringValue(result.WorkspaceId))

	if _, err := waiter.WorkspaceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Workspace (%s) to be created: %w", d.Id(), err))
	}

	return resourceAwsPrometheusWorkspaceRead(ctx, d, meta)
}

func resourceAwsPrometheusWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusserviceconn

	_, err := conn.DeleteWorkspaceWithContext(ctx, &prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if _, err := waiter.WorkspaceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Workspace (%s) to be deleted: %w", d.Id(), err))
	}

	return diag.FromErr(err)
}
