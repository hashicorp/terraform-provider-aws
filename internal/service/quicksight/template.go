// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_template", name="Template")
// @Tags(identifierAttribute="arn")
func ResourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTemplateCreate,
		ReadWithoutTimeout:   resourceTemplateRead,
		UpdateWithoutTimeout: resourceTemplateUpdate,
		DeleteWithoutTimeout: resourceTemplateDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"definition": quicksightschema.TemplateDefinitionSchema(),
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
				names.AttrPermissions: {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 1,
					MaxItems: 64,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 16,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrPrincipal: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
				"source_entity": quicksightschema.TemplateSourceEntitySchema(),
				"source_entity_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"template_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"version_description": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 512),
				},
				"version_number": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameTemplate = "Template"
)

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}
	templateId := d.Get("template_id").(string)

	d.SetId(createTemplateId(awsAccountId, templateId))

	input := &quicksight.CreateTemplateInput{
		AwsAccountId: aws.String(awsAccountId),
		TemplateId:   aws.String(templateId),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SourceEntity = quicksightschema.ExpandTemplateSourceEntity(v.([]interface{}))
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Definition = quicksightschema.ExpandTemplateDefinition(d.Get("definition").([]interface{}))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = expandResourcePermissions(v.List())
	}

	_, err := conn.CreateTemplateWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameTemplate, d.Get(names.AttrName).(string), err)
	}

	if _, err := waitTemplateCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForCreation, ResNameTemplate, d.Id(), err)
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, templateId, err := ParseTemplateId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := FindTemplateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionReading, ResNameTemplate, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set(names.AttrCreatedTime, out.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, out.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Version.Status)
	d.Set("source_entity_arn", out.Version.SourceEntityArn)
	d.Set("template_id", out.TemplateId)
	d.Set("version_description", out.Version.Description)
	d.Set("version_number", out.Version.VersionNumber)

	descResp, err := conn.DescribeTemplateDefinitionWithContext(ctx, &quicksight.DescribeTemplateDefinitionInput{
		AwsAccountId:  aws.String(awsAccountId),
		TemplateId:    aws.String(templateId),
		VersionNumber: out.Version.VersionNumber,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Template (%s) Definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenTemplateDefinition(descResp.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	permsResp, err := conn.DescribeTemplatePermissionsWithContext(ctx, &quicksight.DescribeTemplatePermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		TemplateId:   aws.String(templateId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Template (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, templateId, err := ParseTemplateId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		in := &quicksight.UpdateTemplateInput{
			AwsAccountId:       aws.String(awsAccountId),
			TemplateId:         aws.String(templateId),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			VersionDescription: aws.String(d.Get("version_description").(string)),
		}

		// One of source_entity or definition is required for update
		if _, ok := d.GetOk("source_entity"); ok {
			in.SourceEntity = quicksightschema.ExpandTemplateSourceEntity(d.Get("source_entity").([]interface{}))
		} else {
			in.Definition = quicksightschema.ExpandTemplateDefinition(d.Get("definition").([]interface{}))
		}

		log.Printf("[DEBUG] Updating QuickSight Template (%s): %#v", d.Id(), in)
		_, err := conn.UpdateTemplateWithContext(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionUpdating, ResNameTemplate, d.Id(), err)
		}

		if _, err := waitTemplateUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForUpdate, ResNameTemplate, d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		oraw, nraw := d.GetChange(names.AttrPermissions)
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		toGrant, toRevoke := DiffPermissions(o.List(), n.List())

		params := &quicksight.UpdateTemplatePermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			TemplateId:   aws.String(templateId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateTemplatePermissionsWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Template (%s) permissions: %s", templateId, err)
		}
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, templateId, err := ParseTemplateId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Template %s", d.Id())
	_, err = conn.DeleteTemplateWithContext(ctx, &quicksight.DeleteTemplateInput{
		AwsAccountId: aws.String(awsAccountId),
		TemplateId:   aws.String(templateId),
	})

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionDeleting, ResNameTemplate, d.Id(), err)
	}

	return diags
}

func FindTemplateByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.Template, error) {
	awsAccountId, templateId, err := ParseTemplateId(id)
	if err != nil {
		return nil, err
	}

	descOpts := &quicksight.DescribeTemplateInput{
		AwsAccountId: aws.String(awsAccountId),
		TemplateId:   aws.String(templateId),
	}

	out, err := conn.DescribeTemplateWithContext(ctx, descOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: descOpts,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Template == nil {
		return nil, tfresource.NewEmptyResultError(descOpts)
	}

	return out.Template, nil
}

func ParseTemplateId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,TEMPLATE_ID", id)
	}
	return parts[0], parts[1], nil
}

func createTemplateId(awsAccountID, templateId string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, templateId)
}
