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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceCreate,
		ReadWithoutTimeout:   resourceWorkspaceRead,
		UpdateWithoutTimeout: resourceWorkspaceUpdate,
		DeleteWithoutTimeout: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

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
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				ValidateFunc: verify.ValidARN,
				Optional:     true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).AMPConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	details, err := conn.DescribeWorkspaceWithContext(ctx, &prometheusservice.DescribeWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) && !d.IsNewResource() {
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

	loggingConfig, err := conn.DescribeLoggingConfigurationWithContext(ctx, &prometheusservice.DescribeLoggingConfigurationInput{WorkspaceId: aws.String(d.Id())})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
			d.Set("cloudwatch_log_group_arn", "")
		} else {
			return diag.FromErr(fmt.Errorf("error reading Prometheus logging coniguration for workspace (%s): %w", d.Id(), err))
		}
	} else {
		d.Set("cloudwatch_log_group_arn", loggingConfig.LoggingConfiguration.LogGroupArn)
	}

	tags, err := ListTags(conn, *ws.Arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for Prometheus Workspace (%s): %s", d.Id(), err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}
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
	conn := meta.(*conns.AWSClient).AMPConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Prometheus WorkSpace (%s) tags: %s", d.Id(), err))
		}
	}

	if _, err := conn.UpdateWorkspaceAliasWithContext(ctx, req); err != nil {
		return diag.FromErr(fmt.Errorf("error updating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}

	if d.HasChange("cloudwatch_log_group_arn") {
		_, n := d.GetChange("cloudwatch_log_group_arn")
		_, err := conn.UpdateLoggingConfigurationWithContext(ctx, &prometheusservice.UpdateLoggingConfigurationInput{WorkspaceId: aws.String(d.Id()), LogGroupArn: aws.String(n.(string))})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Logging Configuration (log group arn: %s) for Workspace (%s): %w", n.(string), d.Id(), err))
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Creating AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).AMPConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &prometheusservice.CreateWorkspaceInput{}
	if v, ok := d.GetOk("alias"); ok {
		req.Alias = aws.String(v.(string))
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateWorkspaceWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Prometheus WorkSpace (%s): %w", d.Id(), err))
	}
	d.SetId(aws.StringValue(result.WorkspaceId))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Workspace (%s) to be created: %w", d.Id(), err))
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		_, err := conn.CreateLoggingConfigurationWithContext(ctx, &prometheusservice.CreateLoggingConfigurationInput{WorkspaceId: aws.String(d.Id()), LogGroupArn: aws.String(v.(string))})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating Logging Configuration (log group arn: %s) for Workspace (%s): %w", v.(string), d.Id(), err))
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting AMP workspace %s", d.Id())
	conn := meta.(*conns.AWSClient).AMPConn

	_, err := conn.DeleteWorkspaceWithContext(ctx, &prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Prometheus Workspace (%s): %w", d.Id(), err))
	}

	if _, err := waitWorkspaceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Prometheus Workspace (%s) to be deleted: %w", d.Id(), err))
	}

	return nil
}
