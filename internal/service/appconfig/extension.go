// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrARN: {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ActionPoint](),
						},
						names.AttrAction: {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDescription: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrRoleARN: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrURI: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrDescription: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceExtensionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	in := appconfig.CreateExtensionInput{
		Actions: expandExtensionActionPoints(d.Get("action_point").(*schema.Set).List()),
		Name:    aws.String(d.Get(names.AttrName).(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameter); ok && v.(*schema.Set).Len() > 0 {
		in.Parameters = expandExtensionParameters(v.(*schema.Set).List())
	}

	out, err := conn.CreateExtension(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtension, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtension, d.Get(names.AttrName).(string), errors.New("No Extension returned with create request."))
	}

	d.SetId(aws.ToString(out.Id))

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	out, err := FindExtensionById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppConfig, create.ErrActionReading, ResExtension, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, ResExtension, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("action_point", flattenExtensionActionPoints(out.Actions))
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrParameter, flattenExtensionParameters(out.Parameters))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrVersion, out.VersionNumber)

	return diags
}

func resourceExtensionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)
	requestUpdate := false

	in := &appconfig.UpdateExtensionInput{
		ExtensionIdentifier: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		requestUpdate = true
	}

	if d.HasChange("action_point") {
		in.Actions = expandExtensionActionPoints(d.Get("action_point").(*schema.Set).List())
		requestUpdate = true
	}

	if d.HasChange(names.AttrParameter) {
		in.Parameters = expandExtensionParameters(d.Get(names.AttrParameter).(*schema.Set).List())
		requestUpdate = true
	}

	if requestUpdate {
		out, err := conn.UpdateExtension(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtension, d.Get(names.AttrName).(string), err)
		}

		if out == nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtension, d.Get(names.AttrName).(string), errors.New("No Extension returned with update request."))
		}
	}

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	_, err := conn.DeleteExtension(ctx, &appconfig.DeleteExtensionInput{
		ExtensionIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionDeleting, ResExtension, d.Id(), err)
	}

	return diags
}

func expandExtensionActions(actionsListRaw interface{}) []awstypes.Action {
	var actions []awstypes.Action
	for _, actionRaw := range actionsListRaw.(*schema.Set).List() {
		actionMap, ok := actionRaw.(map[string]interface{})

		if !ok {
			continue
		}

		action := awstypes.Action{
			Description: aws.String(actionMap[names.AttrDescription].(string)),
			Name:        aws.String(actionMap[names.AttrName].(string)),
			RoleArn:     aws.String(actionMap[names.AttrRoleARN].(string)),
			Uri:         aws.String(actionMap[names.AttrURI].(string)),
		}

		actions = append(actions, action)
	}

	return actions
}

func expandExtensionActionPoints(actionsPointListRaw []interface{}) map[string][]awstypes.Action {
	if len(actionsPointListRaw) == 0 {
		return map[string][]awstypes.Action{}
	}

	actionsMap := make(map[string][]awstypes.Action)
	for _, actionPointRaw := range actionsPointListRaw {
		actionPointMap := actionPointRaw.(map[string]interface{})
		actionsMap[actionPointMap["point"].(string)] = expandExtensionActions(actionPointMap[names.AttrAction])
	}

	return actionsMap
}

func expandExtensionParameters(rawParameters []interface{}) map[string]awstypes.Parameter {
	if rawParameters == nil {
		return nil
	}

	parameters := make(map[string]awstypes.Parameter)

	for _, rawParameterMap := range rawParameters {
		parameterMap, ok := rawParameterMap.(map[string]interface{})

		if !ok {
			continue
		}

		parameter := awstypes.Parameter{
			Description: aws.String(parameterMap[names.AttrDescription].(string)),
			Required:    parameterMap["required"].(bool),
		}
		parameters[parameterMap[names.AttrName].(string)] = parameter
	}

	return parameters
}

func flattenExtensionActions(actions []awstypes.Action) []interface{} {
	var rawActions []interface{}
	for _, action := range actions {
		rawAction := map[string]interface{}{
			names.AttrName:        aws.ToString(action.Name),
			names.AttrDescription: aws.ToString(action.Description),
			names.AttrRoleARN:     aws.ToString(action.RoleArn),
			names.AttrURI:         aws.ToString(action.Uri),
		}
		rawActions = append(rawActions, rawAction)
	}
	return rawActions
}

func flattenExtensionActionPoints(actionPointsMap map[string][]awstypes.Action) []interface{} {
	if len(actionPointsMap) == 0 {
		return nil
	}

	var rawActionPoints []interface{}
	for actionPoint, actions := range actionPointsMap {
		rawActionPoint := map[string]interface{}{
			"point":          actionPoint,
			names.AttrAction: flattenExtensionActions(actions),
		}
		rawActionPoints = append(rawActionPoints, rawActionPoint)
	}

	return rawActionPoints
}

func flattenExtensionParameters(parameters map[string]awstypes.Parameter) []interface{} {
	if len(parameters) == 0 {
		return nil
	}

	var rawParameters []interface{}
	for key, parameter := range parameters {
		rawParameter := map[string]interface{}{
			names.AttrName:        key,
			names.AttrDescription: aws.ToString(parameter.Description),
			"required":            parameter.Required,
		}

		rawParameters = append(rawParameters, rawParameter)
	}

	return rawParameters
}
