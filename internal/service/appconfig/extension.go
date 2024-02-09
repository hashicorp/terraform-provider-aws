// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResExtension = "Extension"
)

// @SDKResource("aws_appconfig_extension", name="Extension")
// @Tags(identifierAttribute="arn")
func ResourceExtension() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExtensionCreate,
		ReadWithoutTimeout:   resourceExtensionRead,
		UpdateWithoutTimeout: resourceExtensionUpdate,
		DeleteWithoutTimeout: resourceExtensionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_point": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"point": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appconfig.ActionPoint_Values(), false),
						},
						"action": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"uri": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
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
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceExtensionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	in := appconfig.CreateExtensionInput{
		Actions: expandExtensionActionPoints(d.Get("action_point").(*schema.Set).List()),
		Name:    aws.String(d.Get("name").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter"); ok && v.(*schema.Set).Len() > 0 {
		in.Parameters = expandExtensionParameters(v.(*schema.Set).List())
	}

	out, err := conn.CreateExtensionWithContext(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtension, d.Get("name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtension, d.Get("name").(string), errors.New("No Extension returned with create request."))
	}

	d.SetId(aws.StringValue(out.Id))

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	out, err := FindExtensionById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppConfig, create.ErrActionReading, ResExtension, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, ResExtension, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("action_point", flattenExtensionActionPoints(out.Actions))
	d.Set("description", out.Description)
	d.Set("parameter", flattenExtensionParameters(out.Parameters))
	d.Set("name", out.Name)
	d.Set("version", out.VersionNumber)

	return diags
}

func resourceExtensionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)
	requestUpdate := false

	in := &appconfig.UpdateExtensionInput{
		ExtensionIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		in.Description = aws.String(d.Get("description").(string))
		requestUpdate = true
	}

	if d.HasChange("action_point") {
		in.Actions = expandExtensionActionPoints(d.Get("action_point").(*schema.Set).List())
		requestUpdate = true
	}

	if d.HasChange("parameter") {
		in.Parameters = expandExtensionParameters(d.Get("parameter").(*schema.Set).List())
		requestUpdate = true
	}

	if requestUpdate {
		out, err := conn.UpdateExtensionWithContext(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtension, d.Get("name").(string), err)
		}

		if out == nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtension, d.Get("name").(string), errors.New("No Extension returned with update request."))
		}
	}

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	_, err := conn.DeleteExtensionWithContext(ctx, &appconfig.DeleteExtensionInput{
		ExtensionIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionDeleting, ResExtension, d.Id(), err)
	}

	return diags
}

func expandExtensionActions(actionsListRaw interface{}) []*appconfig.Action {
	var actions []*appconfig.Action
	for _, actionRaw := range actionsListRaw.(*schema.Set).List() {
		actionMap, ok := actionRaw.(map[string]interface{})

		if !ok {
			continue
		}

		action := &appconfig.Action{
			Description: aws.String(actionMap["description"].(string)),
			Name:        aws.String(actionMap["name"].(string)),
			RoleArn:     aws.String(actionMap["role_arn"].(string)),
			Uri:         aws.String(actionMap["uri"].(string)),
		}

		actions = append(actions, action)
	}

	return actions
}

func expandExtensionActionPoints(actionsPointListRaw []interface{}) map[string][]*appconfig.Action {
	if len(actionsPointListRaw) == 0 {
		return map[string][]*appconfig.Action{}
	}

	actionsMap := make(map[string][]*appconfig.Action)
	for _, actionPointRaw := range actionsPointListRaw {
		actionPointMap := actionPointRaw.(map[string]interface{})
		actionsMap[actionPointMap["point"].(string)] = expandExtensionActions(actionPointMap["action"])
	}

	return actionsMap
}

func expandExtensionParameters(rawParameters []interface{}) map[string]*appconfig.Parameter {
	if rawParameters == nil {
		return nil
	}

	parameters := make(map[string]*appconfig.Parameter)

	for _, rawParameterMap := range rawParameters {
		parameterMap, ok := rawParameterMap.(map[string]interface{})

		if !ok {
			continue
		}

		parameter := &appconfig.Parameter{
			Description: aws.String(parameterMap["description"].(string)),
			Required:    aws.Bool(parameterMap["required"].(bool)),
		}
		parameters[parameterMap["name"].(string)] = parameter
	}

	return parameters
}

func flattenExtensionActions(actions []*appconfig.Action) []interface{} {
	var rawActions []interface{}
	for _, action := range actions {
		rawAction := map[string]interface{}{
			"name":        aws.StringValue(action.Name),
			"description": aws.StringValue(action.Description),
			"role_arn":    aws.StringValue(action.RoleArn),
			"uri":         aws.StringValue(action.Uri),
		}
		rawActions = append(rawActions, rawAction)
	}
	return rawActions
}

func flattenExtensionActionPoints(actionPointsMap map[string][]*appconfig.Action) []interface{} {
	if len(actionPointsMap) == 0 {
		return nil
	}

	var rawActionPoints []interface{}
	for actionPoint, actions := range actionPointsMap {
		rawActionPoint := map[string]interface{}{
			"point":  actionPoint,
			"action": flattenExtensionActions(actions),
		}
		rawActionPoints = append(rawActionPoints, rawActionPoint)
	}

	return rawActionPoints
}

func flattenExtensionParameters(parameters map[string]*appconfig.Parameter) []interface{} {
	if len(parameters) == 0 {
		return nil
	}

	var rawParameters []interface{}
	for key, parameter := range parameters {
		rawParameter := map[string]interface{}{
			"name":        key,
			"description": aws.StringValue(parameter.Description),
			"required":    aws.BoolValue(parameter.Required),
		}

		rawParameters = append(rawParameters, rawParameter)
	}

	return rawParameters
}
