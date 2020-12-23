package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsPrometheusWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPrometheusWorkspaceCreate,
		Read:   resourceAwsPrometheusWorkspaceRead,
		Update: resourceAwsPrometheusWorkspaceUpdate,
		Delete: resourceAwsPrometheusWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceAwsPrometheusWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Reading AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusconn

	details, err := conn.DescribeWorkspace(&prometheusservice.DescribeWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	d.Set("arn", aws.StringValue(details.Workspace.Arn))
	d.Set("prometheus_endpoint", aws.StringValue(details.Workspace.PrometheusEndpoint))
	d.Set("status", aws.StringValue(details.Workspace.Status.StatusCode))
	if details.Workspace.Alias != nil {
		d.Set("alias", aws.StringValue(details.Workspace.Alias))
	}

	return nil
}

func resourceAwsPrometheusWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Updating AMP workspace %s", d.Id())

	// today, AMP only supports updating the workspace alias
	if !d.HasChange("alias") {
		return nil
	}

	req := &prometheusservice.UpdateWorkspaceAliasInput{
		WorkspaceId: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("alias"); ok {
		req.SetAlias(v.(string))
	}
	conn := meta.(*AWSClient).prometheusconn
	if _, err := conn.UpdateWorkspaceAlias(req); err != nil {
		return err
	}

	return resourceAwsPrometheusWorkspaceRead(d, meta)
}

func resourceAwsPrometheusWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Creating AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusconn

	req := &prometheusservice.CreateWorkspaceInput{}
	if v, ok := d.GetOk("alias"); ok {
		req.SetAlias(v.(string))
	}

	result, err := conn.CreateWorkspace(req)
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(result.WorkspaceId))

	return resourceAwsPrometheusWorkspaceRead(d, meta)
}

func resourceAwsPrometheusWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting AMP workspace %s", d.Id())
	conn := meta.(*AWSClient).prometheusconn

	_, err := conn.DeleteWorkspace(&prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})
	return err
}
