package ce

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceAnomalyMonitor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAnomalyMonitorCreate,
		ReadContext:   resourceAnomalyMonitorRead,
		UpdateContext: resourceAnomalyMonitorUpdate,
		DeleteContext: resourceAnomalyMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dimension": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice([]string{"SERVICE"}, false),
				ConflictsWith: []string{"specification"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Monitor Name matching expression: [\\S\\s]*")),
			},
			"specification": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				ConflictsWith:    []string{"dimension"},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"DIMENSIONAL", "CUSTOM"}, false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalyMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &costexplorer.CreateAnomalyMonitorInput{
		AnomalyMonitor: &costexplorer.AnomalyMonitor{
			MonitorName: aws.String(d.Get("name").(string)),
			MonitorType: aws.String(d.Get("type").(string)),
		},
	}
	switch d.Get("type").(string) {
	case costexplorer.MonitorTypeDimensional:
		if v, ok := d.GetOk("dimension"); ok {
			input.AnomalyMonitor.MonitorDimension = aws.String(v.(string))
		} else {
			return diag.Errorf("If Monitor Type is %s, dimension attrribute is required", costexplorer.MonitorTypeDimensional)
		}
	case costexplorer.MonitorTypeCustom:
		if v, ok := d.GetOk("specification"); ok {
			expression := costexplorer.Expression{}

			if err := json.Unmarshal([]byte(v.(string)), &expression); err != nil {
				return diag.Errorf("Error parsing specification: %s", err)
			}

			input.AnomalyMonitor.MonitorSpecification = &expression

		} else {
			return diag.Errorf("If Monitor Type is %s, dimension attrribute is required", costexplorer.MonitorTypeCustom)
		}
	}

	if len(tags) > 0 {
		input.ResourceTags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateAnomalyMonitorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("Error creating Anomaly Monitor: %s", err)
	}

	d.SetId(aws.StringValue(resp.MonitorArn))

	return resourceAnomalyMonitorRead(ctx, d, meta)
}

func resourceAnomalyMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig

	resp, err := conn.GetAnomalyMonitorsWithContext(ctx, &costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{d.Id()})})

	if !d.IsNewResource() && len(resp.AnomalyMonitors) < 1 {
		names.LogNotFoundRemoveState(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id(), err)
	}

	anomalyMonitor := resp.AnomalyMonitors[0]

	if anomalyMonitor.MonitorSpecification != nil {
		specificationToJson, err := json.Marshal(anomalyMonitor.MonitorSpecification)
		if err != nil {
			return diag.Errorf("Error parsing specification response: %s", err)
		}
		specificationToSet, err := structure.NormalizeJsonString(string(specificationToJson))

		if err != nil {
			return diag.Errorf("Specification (%s) is invalid JSON: %s", specificationToSet, err)
		}

		d.Set("specification", specificationToSet)
	}

	d.Set("arn", anomalyMonitor.MonitorArn)
	d.Set("dimension", anomalyMonitor.MonitorDimension)
	d.Set("name", anomalyMonitor.MonitorName)
	d.Set("type", anomalyMonitor.MonitorType)

	tags, err := ListTags(conn, aws.StringValue(anomalyMonitor.MonitorArn))

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalyMonitor, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResAnomalyMonitor, d.Id(), err)
	}

	return nil
}

func resourceAnomalyMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	requestUpdate := false

	input := &costexplorer.UpdateAnomalyMonitorInput{
		MonitorArn: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.MonitorName = aws.String(d.Get("Name").(string))
		requestUpdate = true
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return names.DiagError(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id(), err)
		}
	}

	if requestUpdate {
		_, err := conn.UpdateAnomalyMonitorWithContext(ctx, input)

		if err != nil {
			return names.DiagError(names.CE, names.ErrActionReading, ResAnomalyMonitor, d.Id(), err)
		}
	}

	return resourceAnomalyMonitorRead(ctx, d, meta)
}

func resourceAnomalyMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	_, err := conn.DeleteAnomalyMonitorWithContext(ctx, &costexplorer.DeleteAnomalyMonitorInput{MonitorArn: aws.String(d.Id())})

	if err != nil && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionDeleting, ResAnomalyMonitor, d.Id(), err)
	}

	return nil
}
