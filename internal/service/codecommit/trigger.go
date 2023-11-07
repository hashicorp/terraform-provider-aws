// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_codecommit_trigger")
func ResourceTrigger() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTriggerCreate,
		ReadWithoutTimeout:   resourceTriggerRead,
		DeleteWithoutTimeout: resourceTriggerDelete,

		Schema: map[string]*schema.Schema{
			"repository_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"configuration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trigger": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Required: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"destination_arn": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"custom_data": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"branches": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"events": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	// Expand the "trigger" set to aws-sdk-go compat []*codecommit.RepositoryTrigger
	triggers := expandTriggers(d.Get("trigger").(*schema.Set).List())

	input := &codecommit.PutRepositoryTriggersInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
		Triggers:       triggers,
	}

	resp, err := conn.PutRepositoryTriggersWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeCommit Trigger: %s", err)
	}

	log.Printf("[INFO] Code Commit Trigger Created %s input %s", resp, input)

	d.SetId(d.Get("repository_name").(string))
	d.Set("configuration_id", resp.ConfigurationId)

	return append(diags, resourceTriggerRead(ctx, d, meta)...)
}

func resourceTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	input := &codecommit.GetRepositoryTriggersInput{
		RepositoryName: aws.String(d.Id()),
	}

	resp, err := conn.GetRepositoryTriggersWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Trigger: %s", err.Error())
	}

	log.Printf("[DEBUG] CodeCommit Trigger: %s", resp)

	return diags
}

func resourceTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	log.Printf("[DEBUG] Deleting Trigger: %q", d.Id())

	input := &codecommit.PutRepositoryTriggersInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
		Triggers:       []*codecommit.RepositoryTrigger{},
	}

	if _, err := conn.PutRepositoryTriggersWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Trigger (%s): %s", d.Id(), err)
	}

	return diags
}

func expandTriggers(configured []interface{}) []*codecommit.RepositoryTrigger {
	triggers := make([]*codecommit.RepositoryTrigger, 0, len(configured))
	// Loop over our configured triggers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})
		t := &codecommit.RepositoryTrigger{
			CustomData:     aws.String(data["custom_data"].(string)),
			DestinationArn: aws.String(data["destination_arn"].(string)),
			Name:           aws.String(data["name"].(string)),
		}

		branches := make([]*string, len(data["branches"].([]interface{})))
		for i, vv := range data["branches"].([]interface{}) {
			str := vv.(string)
			branches[i] = aws.String(str)
		}
		t.Branches = branches

		events := make([]*string, len(data["events"].([]interface{})))
		for i, vv := range data["events"].([]interface{}) {
			str := vv.(string)
			events[i] = aws.String(str)
		}
		t.Events = events

		triggers = append(triggers, t)
	}
	return triggers
}
