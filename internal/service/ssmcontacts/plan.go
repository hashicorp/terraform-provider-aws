// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssmcontacts_plan", name="Plan")
func ResourcePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlanCreate,
		ReadWithoutTimeout:   resourcePlanRead,
		UpdateWithoutTimeout: resourcePlanUpdate,
		DeleteWithoutTimeout: resourcePlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"contact_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStage: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_in_minutes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrTarget: {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"channel_target_info": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"contact_channel_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"retry_interval_in_minutes": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"contact_target_info": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"is_essential": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"contact_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	ResNamePlan = "Plan"
)

func resourcePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	contactId := d.Get("contact_id").(string)
	stages := expandStages(d.Get(names.AttrStage).([]interface{}))
	plan := &types.Plan{
		Stages: stages,
	}

	in := &ssmcontacts.UpdateContactInput{
		ContactId: aws.String(contactId),
		Plan:      plan,
	}

	_, err := conn.UpdateContact(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts,
			create.ErrActionCreating,
			ResNamePlan,
			contactId,
			err)
	}

	d.SetId(contactId)

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	out, err := findContactByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMContacts Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionReading, ResNamePlan, d.Id(), err)
	}

	if err := setPlanResourceData(d, out); err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionReading, ResNamePlan, d.Id(), err)
	}

	return diags
}

func resourcePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	update := false

	in := &ssmcontacts.UpdateContactInput{
		ContactId: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrStage) {
		stages := expandStages(d.Get(names.AttrStage).([]interface{}))
		in.Plan = &types.Plan{
			Stages: stages,
		}
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating SSMContacts Plan (%s): %#v", d.Id(), in)
	_, err := conn.UpdateContact(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionUpdating, ResNamePlan, d.Id(), err)
	}

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	log.Printf("[INFO] Deleting SSMContacts Plan %s", d.Id())

	_, err := conn.UpdateContact(ctx, &ssmcontacts.UpdateContactInput{
		ContactId: aws.String(d.Id()),
		Plan: &types.Plan{
			Stages: []types.Stage{},
		},
	})

	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionDeleting, ResNamePlan, d.Id(), err)
	}

	return diags
}
