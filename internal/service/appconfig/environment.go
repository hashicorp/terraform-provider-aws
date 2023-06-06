package appconfig

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func ResourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentCreate,
		ReadWithoutTimeout:   resourceEnvironmentRead,
		UpdateWithoutTimeout: resourceEnvironmentUpdate,
		DeleteWithoutTimeout: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"monitor": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_arn": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 2048),
								verify.ValidARN,
							),
						},
						"alarm_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

	appId := d.Get("application_id").(string)
	input := &appconfig.CreateEnvironmentInput{
		Name:          aws.String(d.Get("name").(string)),
		ApplicationId: aws.String(appId),
		Tags:          GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("monitor"); ok && v.(*schema.Set).Len() > 0 {
		input.Monitors = expandEnvironmentMonitors(v.(*schema.Set).List())
	}

	environment, err := conn.CreateEnvironmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Environment for Application (%s): %s", appId, err)
	}

	if environment == nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Environment for Application (%s): empty response", appId)
	}

	d.Set("environment_id", environment.Id)
	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(environment.Id), aws.StringValue(environment.ApplicationId)))

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

	envID, appID, err := EnvironmentParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Environment (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appID),
		EnvironmentId: aws.String(envID),
	}

	output, err := conn.GetEnvironmentWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Environment (%s) for Application (%s) not found, removing from state", envID, appID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Environment (%s) for Application (%s): %s", envID, appID, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Environment (%s) for Application (%s): empty response", envID, appID)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("environment_id", output.Id)
	d.Set("description", output.Description)
	d.Set("name", output.Name)
	d.Set("state", output.State)

	if err := d.Set("monitor", flattenEnvironmentMonitors(output.Monitors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitor: %s", err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/environment/%s", appID, envID),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

	if d.HasChangesExcept("tags", "tags_all") {
		envID, appID, err := EnvironmentParseID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Environment (%s): %s", d.Id(), err)
		}

		updateInput := &appconfig.UpdateEnvironmentInput{
			EnvironmentId: aws.String(envID),
			ApplicationId: aws.String(appID),
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			updateInput.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("monitor") {
			updateInput.Monitors = expandEnvironmentMonitors(d.Get("monitor").(*schema.Set).List())
		}

		_, err = conn.UpdateEnvironmentWithContext(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Environment (%s) for Application (%s): %s", envID, appID, err)
		}
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

	envID, appID, err := EnvironmentParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Environment (%s): %s", d.Id(), err)
	}

	input := &appconfig.DeleteEnvironmentInput{
		EnvironmentId: aws.String(envID),
		ApplicationId: aws.String(appID),
	}

	_, err = conn.DeleteEnvironmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Environment (%s) for Application (%s): %s", envID, appID, err)
	}

	return diags
}

func EnvironmentParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected EnvironmentID:ApplicationID", id)
	}

	return parts[0], parts[1], nil
}

func expandEnvironmentMonitor(tfMap map[string]interface{}) *appconfig.Monitor {
	if tfMap == nil {
		return nil
	}

	monitor := &appconfig.Monitor{}

	if v, ok := tfMap["alarm_arn"].(string); ok && v != "" {
		monitor.AlarmArn = aws.String(v)
	}

	if v, ok := tfMap["alarm_role_arn"].(string); ok && v != "" {
		monitor.AlarmRoleArn = aws.String(v)
	}

	return monitor
}

func expandEnvironmentMonitors(tfList []interface{}) []*appconfig.Monitor {
	// AppConfig API requires a 0 length slice instead of a nil value
	// when updating from N monitors to 0/nil monitors
	monitors := make([]*appconfig.Monitor, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		monitor := expandEnvironmentMonitor(tfMap)

		if monitor == nil {
			continue
		}

		monitors = append(monitors, monitor)
	}

	return monitors
}

func flattenEnvironmentMonitor(monitor *appconfig.Monitor) map[string]interface{} {
	if monitor == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := monitor.AlarmArn; v != nil {
		tfMap["alarm_arn"] = aws.StringValue(v)
	}

	if v := monitor.AlarmRoleArn; v != nil {
		tfMap["alarm_role_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEnvironmentMonitors(monitors []*appconfig.Monitor) []interface{} {
	if len(monitors) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, monitor := range monitors {
		if monitor == nil {
			continue
		}

		tfList = append(tfList, flattenEnvironmentMonitor(monitor))
	}

	return tfList
}
