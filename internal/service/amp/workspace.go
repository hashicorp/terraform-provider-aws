package amp

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			// Once set, alias cannot be unset.
			customdiff.ForceNewIfChange("alias", func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) != "" && new.(string) == ""
			}),
			verify.SetTagsDiff,
		),

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
	conn := meta.(*conns.AWSClient).AMPConn()
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
		return diag.Errorf("waiting for Prometheus Workspace (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &prometheusservice.CreateLoggingConfigurationInput{
			LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.CreateLoggingConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("creating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
		}

		if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for Prometheus Workspace (%s) logging configuration create: %s", d.Id(), err)
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ws, err := FindWorkspaceByID(ctx, conn, d.Id())

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

	loggingConfiguration, err := FindLoggingConfigurationByWorkspaceID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		d.Set("logging_configuration", nil)
	} else if err != nil {
		return diag.Errorf("reading Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
	} else {
		if err := d.Set("logging_configuration", []interface{}{flattenLoggingConfigurationMetadata(loggingConfiguration)}); err != nil {
			return diag.Errorf("setting logging_configuration: %s", err)
		}
	}

	tags, err := ListTags(ctx, conn, arn)

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
	conn := meta.(*conns.AWSClient).AMPConn()

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
			return diag.Errorf("waiting for Prometheus Workspace (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("logging_configuration") {
		if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if o, _ := d.GetChange("logging_configuration"); o == nil || len(o.([]interface{})) == 0 || o.([]interface{})[0] == nil {
				input := &prometheusservice.CreateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.CreateLoggingConfigurationWithContext(ctx, input); err != nil {
					return diag.Errorf("creating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id()); err != nil {
					return diag.Errorf("waiting for Prometheus Workspace (%s) logging configuration create: %s", d.Id(), err)
				}
			} else {
				input := &prometheusservice.UpdateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.UpdateLoggingConfigurationWithContext(ctx, input); err != nil {
					return diag.Errorf("updating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationUpdated(ctx, conn, d.Id()); err != nil {
					return diag.Errorf("waiting for Prometheus Workspace (%s) logging configuration update: %s", d.Id(), err)
				}
			}
		} else {
			_, err := conn.DeleteLoggingConfigurationWithContext(ctx, &prometheusservice.DeleteLoggingConfigurationInput{
				WorkspaceId: aws.String(d.Id()),
			})

			if err != nil {
				return diag.Errorf("deleting Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
			}

			if _, err := waitLoggingConfigurationDeleted(ctx, conn, d.Id()); err != nil {
				return diag.Errorf("waiting for Prometheus Workspace (%s) logging configuration delete: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return diag.Errorf("updating Prometheus Workspace (%s) tags: %s", arn, err)
		}
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn()

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

func flattenLoggingConfigurationMetadata(apiObject *prometheusservice.LoggingConfigurationMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupArn; v != nil {
		tfMap["log_group_arn"] = aws.StringValue(v)
	}

	return tfMap
}
