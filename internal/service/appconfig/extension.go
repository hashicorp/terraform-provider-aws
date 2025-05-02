// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_extension", name="Extension")
// @Tags(identifierAttribute="arn")
func resourceExtension() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExtensionCreate,
		ReadWithoutTimeout:   resourceExtensionRead,
		UpdateWithoutTimeout: resourceExtensionUpdate,
		DeleteWithoutTimeout: resourceExtensionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"action_point": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
										Optional: true,
									},
									names.AttrURI: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"point": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ActionPoint](),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
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
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceExtensionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := appconfig.CreateExtensionInput{
		Actions: expandActionPoints(d.Get("action_point").(*schema.Set).List()),
		Name:    aws.String(name),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameter); ok && v.(*schema.Set).Len() > 0 {
		input.Parameters = expandParameters(v.(*schema.Set).List())
	}

	output, err := conn.CreateExtension(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Extension (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	output, err := findExtensionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Extension (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Extension (%s): %s", d.Id(), err)
	}

	if err := d.Set("action_point", flattenActionPoints(output.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting action_point: %s", err)
	}
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.Name)
	if err := d.Set(names.AttrParameter, flattenParameters(output.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}
	d.Set(names.AttrVersion, output.VersionNumber)

	return diags
}

func resourceExtensionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := appconfig.UpdateExtensionInput{
			ExtensionIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("action_point") {
			input.Actions = expandActionPoints(d.Get("action_point").(*schema.Set).List())
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrParameter) {
			input.Parameters = expandParameters(d.Get(names.AttrParameter).(*schema.Set).List())
		}

		_, err := conn.UpdateExtension(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Extension (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceExtensionRead(ctx, d, meta)...)
}

func resourceExtensionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Extension: %s", d.Id())
	input := appconfig.DeleteExtensionInput{
		ExtensionIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteExtension(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Extension (%s): %s", d.Id(), err)
	}

	return diags
}

func findExtensionByID(ctx context.Context, conn *appconfig.Client, id string) (*appconfig.GetExtensionOutput, error) {
	input := &appconfig.GetExtensionInput{
		ExtensionIdentifier: aws.String(id),
	}

	return findExtension(ctx, conn, input)
}

func findExtension(ctx context.Context, conn *appconfig.Client, input *appconfig.GetExtensionInput) (*appconfig.GetExtensionOutput, error) {
	output, err := conn.GetExtension(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandActions(tfList []any) []awstypes.Action {
	var apiObjects []awstypes.Action

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.Action{
			Description: aws.String(tfMap[names.AttrDescription].(string)),
			Name:        aws.String(tfMap[names.AttrName].(string)),
			RoleArn:     aws.String(tfMap[names.AttrRoleARN].(string)),
			Uri:         aws.String(tfMap[names.AttrURI].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandActionPoints(tfList []any) map[string][]awstypes.Action {
	if len(tfList) == 0 {
		return map[string][]awstypes.Action{}
	}

	apiObjects := make(map[string][]awstypes.Action)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects[tfMap["point"].(string)] = expandActions(tfMap[names.AttrAction].(*schema.Set).List())
	}

	return apiObjects
}

func expandParameters(tfList []any) map[string]awstypes.Parameter {
	if tfList == nil {
		return nil
	}

	apiObjects := make(map[string]awstypes.Parameter)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects[tfMap[names.AttrName].(string)] = awstypes.Parameter{
			Description: aws.String(tfMap[names.AttrDescription].(string)),
			Required:    tfMap["required"].(bool),
		}
	}

	return apiObjects
}

func flattenActions(apiObjects []awstypes.Action) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrDescription: aws.ToString(apiObject.Description),
			names.AttrName:        aws.ToString(apiObject.Name),
			names.AttrRoleARN:     aws.ToString(apiObject.RoleArn),
			names.AttrURI:         aws.ToString(apiObject.Uri),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenActionPoints(apiObjects map[string][]awstypes.Action) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for k, v := range apiObjects {
		rawActionPoint := map[string]any{
			names.AttrAction: flattenActions(v),
			"point":          k,
		}
		tfList = append(tfList, rawActionPoint)
	}

	return tfList
}

func flattenParameters(apiObjects map[string]awstypes.Parameter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for k, v := range apiObjects {
		tfMap := map[string]any{
			names.AttrDescription: aws.ToString(v.Description),
			names.AttrName:        k,
			"required":            v.Required,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
