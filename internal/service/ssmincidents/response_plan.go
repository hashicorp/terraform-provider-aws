// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameResponsePlan = "Response Plan"
)

// @SDKResource("aws_ssmincidents_response_plan", name="Response Plan")
// @Tags(identifierAttribute="id")
func ResourceResponsePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResponsePlanCreate,
		ReadWithoutTimeout:   resourceResponsePlanRead,
		UpdateWithoutTimeout: resourceResponsePlanUpdate,
		DeleteWithoutTimeout: resourceResponsePlanDelete,

		Schema: map[string]*schema.Schema{
			"action": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssm_automation": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"document_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"document_version": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"target_account": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"values": {
													Type:     schema.TypeSet,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"dynamic_parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"chat_channel": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engagements": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"incident_template": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:     schema.TypeString,
							Required: true,
						},
						"impact": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"dedupe_string": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"incident_tags": tftags.TagsSchema(),
						"notification_target": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sns_topic_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"summary": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"integration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pagerduty": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"service_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"secret_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceResponsePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	input := &ssmincidents.CreateResponsePlanInput{
		Actions:          expandAction(d.Get("action").([]interface{})),
		ChatChannel:      expandChatChannel(d.Get("chat_channel").(*schema.Set)),
		DisplayName:      aws.String(d.Get("display_name").(string)),
		Engagements:      flex.ExpandStringValueSet(d.Get("engagements").(*schema.Set)),
		IncidentTemplate: expandIncidentTemplate(d.Get("incident_template").([]interface{})),
		Integrations:     expandIntegration(d.Get("integration").([]interface{})),
		Name:             aws.String(d.Get("name").(string)),
		Tags:             getTagsIn(ctx),
	}

	output, err := client.CreateResponsePlan(ctx, input)

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameResponsePlan, d.Get("name").(string), err)
	}

	if output == nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameResponsePlan, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(output.Arn))

	return resourceResponsePlanRead(ctx, d, meta)
}

func resourceResponsePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	responsePlan, err := FindResponsePlanByID(ctx, client, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMIncidents ResponsePlan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameResponsePlan, d.Id(), err)
	}

	if d, err := setResponsePlanResourceData(d, responsePlan); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameResponsePlan, d.Id(), err)
	}

	return nil
}

func resourceResponsePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ssmincidents.UpdateResponsePlanInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChanges("action") {
			input.Actions = expandAction(d.Get("action").([]interface{}))
		}

		if d.HasChanges("chat_channel") {
			input.ChatChannel = expandChatChannel(d.Get("chat_channel").(*schema.Set))
		}

		if d.HasChanges("display_name") {
			input.DisplayName = aws.String(d.Get("display_name").(string))
		}

		if d.HasChanges("engagements") {
			input.Engagements = flex.ExpandStringValueSet(d.Get("engagements").(*schema.Set))
		}

		if d.HasChanges("incident_template") {
			incidentTemplate := d.Get("incident_template")
			template := expandIncidentTemplate(incidentTemplate.([]interface{}))
			updateResponsePlanInputWithIncidentTemplate(input, template)
		}

		if d.HasChanges("integration") {
			input.Integrations = expandIntegration(d.Get("integration").([]interface{}))
		}

		_, err := client.UpdateResponsePlan(ctx, input)

		if err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameResponsePlan, d.Id(), err)
		}
	}

	return resourceResponsePlanRead(ctx, d, meta)
}

func resourceResponsePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	log.Printf("[INFO] Deleting SSMIncidents ResponsePlan %s", d.Id())

	input := &ssmincidents.DeleteResponsePlanInput{
		Arn: aws.String(d.Id()),
	}

	_, err := client.DeleteResponsePlan(ctx, input)

	if err != nil {
		var notFoundError *types.ResourceNotFoundException

		if errors.As(err, &notFoundError) {
			return nil
		}

		return create.DiagError(names.SSMIncidents, create.ErrActionDeleting, ResNameResponsePlan, d.Id(), err)
	}

	return nil
}

// input validation already done in flattenIncidentTemplate function
func updateResponsePlanInputWithIncidentTemplate(input *ssmincidents.UpdateResponsePlanInput, template *types.IncidentTemplate) {
	input.IncidentTemplateImpact = template.Impact
	input.IncidentTemplateTitle = template.Title
	input.IncidentTemplateTags = template.IncidentTags
	input.IncidentTemplateNotificationTargets = template.NotificationTargets
	input.IncidentTemplateDedupeString = template.DedupeString
	input.IncidentTemplateSummary = template.Summary
}
