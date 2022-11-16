package auditmanager

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceControlCreate,
		ReadWithoutTimeout:   resourceControlRead,
		UpdateWithoutTimeout: resourceControlUpdate,
		DeleteWithoutTimeout: resourceControlDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"action_plan_instructions": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"action_plan_title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control_mapping_sources": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Note: ForceNew is applied to all Elem attributes because the set type will
						// not preserve the computed source_id required for Update API requests.
						"source_description": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"source_frequency": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.SourceFrequency](),
						},
						"source_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_keyword": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"keyword_input_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.KeywordInputType](),
									},
									"keyword_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"source_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_set_up_option": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.SourceSetUpOption](),
						},
						"source_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.SourceType](),
						},
						"troubleshooting_text": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"testing_information": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameControl = "Control"
)

func resourceControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AuditManagerClient

	in := &auditmanager.CreateControlInput{
		Name:                  aws.String(d.Get("name").(string)),
		ControlMappingSources: expandControlMappingSourcesCreate(d.Get("control_mapping_sources").(*schema.Set)),
	}

	if v, ok := d.GetOk("action_plan_instructions"); ok {
		in.ActionPlanInstructions = aws.String((v.(string)))
	}
	if v, ok := d.GetOk("action_plan_title"); ok {
		in.ActionPlanTitle = aws.String((v.(string)))
	}
	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String((v.(string)))
	}
	if v, ok := d.GetOk("testing_information"); ok {
		in.TestingInformation = aws.String((v.(string)))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateControl(ctx, in)
	if err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionCreating, ResNameControl, d.Get("name").(string), err)
	}
	if out == nil || out.Control == nil {
		return create.DiagError(names.AuditManager, create.ErrActionCreating, ResNameControl, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Control.Id))
	return resourceControlRead(ctx, d, meta)
}

func resourceControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AuditManagerClient

	out, err := FindControlByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AuditManager Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionReading, ResNameControl, d.Id(), err)
	}
	d.SetId(aws.ToString(out.Id))

	d.Set("action_plan_instructions", aws.ToString(out.ActionPlanInstructions))
	d.Set("action_plan_title", aws.ToString(out.ActionPlanTitle))
	d.Set("arn", aws.ToString(out.Arn))
	if err := d.Set("control_mapping_sources", flattenControlMappingSources(out.ControlMappingSources)); err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionSetting, ResNameControl, d.Id(), err)
	}
	d.Set("description", aws.ToString(out.Description))
	d.Set("name", aws.ToString(out.Name))
	d.Set("testing_information", aws.ToString(out.TestingInformation))
	d.Set("type", string(out.Type))

	tags, err := ListTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionReading, ResNameControl, d.Id(), err)
	}
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionSetting, ResNameControl, d.Id(), err)
	}
	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.AuditManager, create.ErrActionSetting, ResNameControl, d.Id(), err)
	}

	return nil
}

func resourceControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AuditManagerClient

	if d.HasChanges(
		"action_plan_instructions",
		"action_plan_title",
		"control_mapping_sources",
		"description",
		"testing_information",
	) {
		in := &auditmanager.UpdateControlInput{
			ControlId:             aws.String(d.Id()),
			Name:                  aws.String(d.Get("name").(string)),
			ControlMappingSources: expandControlMappingSourcesUpdate(d.Get("control_mapping_sources").(*schema.Set)),
		}
		if v, ok := d.GetOk("action_plan_instructions"); ok {
			in.ActionPlanInstructions = aws.String((v.(string)))
		}
		if v, ok := d.GetOk("action_plan_title"); ok {
			in.ActionPlanTitle = aws.String((v.(string)))
		}
		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String((v.(string)))
		}
		if v, ok := d.GetOk("testing_information"); ok {
			in.TestingInformation = aws.String((v.(string)))
		}

		log.Printf("[DEBUG] Updating AuditManager Control (%s): %#v", d.Id(), in)
		_, err := conn.UpdateControl(ctx, in)
		if err != nil {
			return create.DiagError(names.AuditManager, create.ErrActionUpdating, ResNameControl, d.Id(), err)
		}
	}

	if d.HasChanges("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.AuditManager, create.ErrActionUpdating, ResNameControl, d.Id(), err)
		}
	}

	return resourceControlRead(ctx, d, meta)
}

func resourceControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AuditManagerClient

	log.Printf("[INFO] Deleting AuditManager Control %s", d.Id())

	_, err := conn.DeleteControl(ctx, &auditmanager.DeleteControlInput{
		ControlId: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.AuditManager, create.ErrActionDeleting, ResNameControl, d.Id(), err)
	}

	return nil
}

func FindControlByID(ctx context.Context, conn *auditmanager.Client, id string) (*types.Control, error) {
	in := &auditmanager.GetControlInput{
		ControlId: aws.String(id),
	}
	out, err := conn.GetControl(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Control == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Control, nil
}

func flattenControlMappingSources(apiObjects []types.ControlMappingSource) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, obj := range apiObjects {
		m := map[string]interface{}{
			"source_description":   aws.ToString(obj.SourceDescription),
			"source_frequency":     string(obj.SourceFrequency),
			"source_id":            aws.ToString(obj.SourceId),
			"source_keyword":       flattenControlMappingSourceSourceKeyword(obj.SourceKeyword),
			"source_name":          aws.ToString(obj.SourceName),
			"source_set_up_option": string(obj.SourceSetUpOption),
			"source_type":          string(obj.SourceType),
			"troubleshooting_text": aws.ToString(obj.TroubleshootingText),
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenControlMappingSourceSourceKeyword(apiObject *types.SourceKeyword) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"keyword_input_type": string(apiObject.KeywordInputType),
		"keyword_value":      aws.ToString(apiObject.KeywordValue),
	}
	return []interface{}{m}
}

func expandControlMappingSourcesCreate(tfSet *schema.Set) []types.CreateControlMappingSource {
	tfList := tfSet.List()
	var ccms []types.CreateControlMappingSource

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		item := types.CreateControlMappingSource{
			SourceName:        aws.String(m["source_name"].(string)),
			SourceSetUpOption: types.SourceSetUpOption(m["source_set_up_option"].(string)),
		}
		if v, ok := m["source_description"].(string); ok {
			item.SourceDescription = aws.String(v)
		}
		if v, ok := m["source_frequency"].(string); ok {
			item.SourceFrequency = types.SourceFrequency(v)
		}
		if v, ok := m["source_keyword"].([]interface{}); ok {
			item.SourceKeyword = expandControlMappingSourceSourceKeyword(v)
		}
		if v, ok := m["source_type"].(string); ok {
			item.SourceType = types.SourceType(v)
		}
		if v, ok := m["troubleshooting_text"].(string); ok {
			item.TroubleshootingText = aws.String(v)
		}

		ccms = append(ccms, item)
	}
	return ccms
}

func expandControlMappingSourcesUpdate(tfSet *schema.Set) []types.ControlMappingSource {
	tfList := tfSet.List()
	var cms []types.ControlMappingSource

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		item := types.ControlMappingSource{
			// SourceId is required on update. This is the only field that differs between
			// the ControlMappingSource and CreateControlMappingSource structs
			SourceId:          aws.String(m["source_id"].(string)),
			SourceName:        aws.String(m["source_name"].(string)),
			SourceSetUpOption: types.SourceSetUpOption(m["source_set_up_option"].(string)),
		}
		if v, ok := m["source_description"].(string); ok {
			item.SourceDescription = aws.String(v)
		}
		if v, ok := m["source_frequency"].(string); ok {
			item.SourceFrequency = types.SourceFrequency(v)
		}
		if v, ok := m["source_keyword"].([]interface{}); ok {
			item.SourceKeyword = expandControlMappingSourceSourceKeyword(v)
		}
		if v, ok := m["source_type"].(string); ok {
			item.SourceType = types.SourceType(v)
		}
		if v, ok := m["troubleshooting_text"].(string); ok {
			item.TroubleshootingText = aws.String(v)
		}

		cms = append(cms, item)
	}
	return cms
}

func expandControlMappingSourceSourceKeyword(tfList []interface{}) *types.SourceKeyword {
	if len(tfList) == 0 {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sk := types.SourceKeyword{}
	if v, ok := tfMap["keyword_input_type"].(string); ok {
		sk.KeywordInputType = types.KeywordInputType(v)
	}
	if v, ok := tfMap["keyword_value"].(string); ok {
		sk.KeywordValue = aws.String(v)
	}

	return &sk
}
