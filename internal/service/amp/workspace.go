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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &prometheusservice.CreateWorkspaceInput{}

	if v, ok := d.GetOk("alias"); ok {
		input.Alias = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateWorkspaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Prometheus Workspace: %s", err)
	}

	d.SetId(aws.StringValue(result.WorkspaceId))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Prometheus Workspace (%s) create: %w", d.Id(), err)
	}

	// TODO
	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		_, err := conn.CreateLoggingConfigurationWithContext(ctx, &prometheusservice.CreateLoggingConfigurationInput{WorkspaceId: aws.String(d.Id()), LogGroupArn: aws.String(v.(string))})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating Logging Configuration (log group arn: %s) for Workspace (%s): %w", v.(string), d.Id(), err))
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ws, err := FindWorkspaceByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Prometheus Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Prometheus Workspace (%s): %s", d.Id(), err)
	}

	d.Set("alias", ws.Alias)
	arn := aws.StringValue(ws.Arn)
	d.Set("arn", arn)
	d.Set("prometheus_endpoint", ws.PrometheusEndpoint)

	// TODO
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

	tags, err := ListTagsWithContext(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Prometheus Workspace (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}
	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	if d.HasChange("alias") {
		input := &prometheusservice.UpdateWorkspaceAliasInput{
			Alias:       aws.String(d.Get("alias").(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkspaceAliasWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Prometheus Workspace alias (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for Prometheus Workspace (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := UpdateTagsWithContext(ctx, conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("updating Prometheus Workspace (%s) tags: %s", arn, err))
		}
	}

	// TODO
	if d.HasChange("cloudwatch_log_group_arn") {
		_, n := d.GetChange("cloudwatch_log_group_arn")
		_, err := conn.UpdateLoggingConfigurationWithContext(ctx, &prometheusservice.UpdateLoggingConfigurationInput{WorkspaceId: aws.String(d.Id()), LogGroupArn: aws.String(n.(string))})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Logging Configuration (log group arn: %s) for Workspace (%s): %w", n.(string), d.Id(), err))
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn

	log.Printf("[INFO] Deleting Prometheus Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspaceWithContext(ctx, &prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Prometheus Workspace (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkspaceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Prometheus Workspace (%s) delete: %s", d.Id(), err)
	}

	return nil
}
