package amp

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRuleGroupNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRuleGroupNamespaceCreate,
		ReadContext:   resourceRuleGroupNamespaceRead,
		UpdateContext: resourceRuleGroupNamespaceUpdate,
		DeleteContext: resourceRuleGroupNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRuleGroupNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	workspaceID := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	input := &prometheusservice.CreateRuleGroupsNamespaceInput{
		Name:        aws.String(name),
		Data:        []byte(d.Get("data").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	log.Printf("[DEBUG] Creating Prometheus Rule Group Namespace: %s", input)
	output, err := conn.CreateRuleGroupsNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus Rule Group Namespace (%s) for workspace (%s): %w", name, workspaceID, err))
	}

	d.SetId(aws.StringValue(output.Arn))

	if _, err := waitRuleGroupNamespaceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Rule Group Namespace (%s) create: %w", d.Id(), err))
	}

	return resourceRuleGroupNamespaceRead(ctx, d, meta)
}

func resourceRuleGroupNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	input := &prometheusservice.PutRuleGroupsNamespaceInput{
		Name:        aws.String(d.Get("name").(string)),
		Data:        []byte(d.Get("data").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Updating Prometheus Rule Group Namespace: %s", input)
	_, err := conn.PutRuleGroupsNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus Rule Group Namespace (%s): %w", d.Id(), err))
	}

	if _, err := waitRuleGroupNamespaceUpdated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Rule Group Namespace (%s) update: %w", d.Id(), err))
	}

	return resourceRuleGroupNamespaceRead(ctx, d, meta)
}

func resourceRuleGroupNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	rgn, err := FindRuleGroupNamespaceByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Rule Group Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Prometheus Rule Group Namespace (%s): %w", d.Id(), err))
	}

	d.Set("data", string(rgn.Data))
	d.Set("name", rgn.Name)
	_, workspaceID, err := nameAndWorkspaceIDFromRuleGroupNamespaceARN(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("workspace_id", workspaceID)

	return nil
}

func resourceRuleGroupNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	log.Printf("[DEBUG] Deleting Prometheus Rule Group Namespace: (%s)", d.Id())
	_, err := conn.DeleteRuleGroupsNamespaceWithContext(ctx, &prometheusservice.DeleteRuleGroupsNamespaceInput{
		Name:        aws.String(d.Get("name").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Prometheus Rule Group Namespace (%s): %w", d.Id(), err))
	}

	if _, err := waitRuleGroupNamespaceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Rule Group Namespace (%s) delete: %w", d.Id(), err))
	}

	return nil
}
