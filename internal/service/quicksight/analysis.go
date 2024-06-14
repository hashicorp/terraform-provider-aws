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

const (
	recoveryWindowInDaysMin     = 7
	recoveryWindowInDaysMax     = 30
	recoveryWindowInDaysDefault = recoveryWindowInDaysMax
)

// @SDKResource("aws_quicksight_analysis", name="Analysis")
// @Tags(identifierAttribute="arn")
func ResourceAnalysis() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnalysisCreate,
		ReadWithoutTimeout:   resourceAnalysisRead,
		UpdateWithoutTimeout: resourceAnalysisUpdate,
		DeleteWithoutTimeout: resourceAnalysisDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("recovery_window_in_days", recoveryWindowInDaysDefault)
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
				"analysis_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"definition": quicksightschema.AnalysisDefinitionSchema(),
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"last_published_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
				names.AttrParameters: quicksightschema.ParametersSchema(),
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
				"recovery_window_in_days": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  30,
					ValidateFunc: validation.Any(
						validation.IntBetween(recoveryWindowInDaysMin, recoveryWindowInDaysMax),
						validation.IntInSlice([]int{0}),
					),
				},
				"source_entity": quicksightschema.AnalysisSourceEntitySchema(),
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"theme_arn": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameAnalysis = "Analysis"
)

func resourceAnalysisCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}
	analysisId := d.Get("analysis_id").(string)

	d.SetId(createAnalysisId(awsAccountId, analysisId))

	input := &quicksight.CreateAnalysisInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SourceEntity = quicksightschema.ExpandAnalysisSourceEntity(v.([]interface{}))
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Definition = quicksightschema.ExpandAnalysisDefinition(d.Get("definition").([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]interface{}))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = expandResourcePermissions(v.List())
	}

	_, err := conn.CreateAnalysisWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameAnalysis, d.Get(names.AttrName).(string), err)
	}

	if _, err := waitAnalysisCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForCreation, ResNameAnalysis, d.Id(), err)
	}

	return append(diags, resourceAnalysisRead(ctx, d, meta)...)
}

func resourceAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, analysisId, err := ParseAnalysisId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := FindAnalysisByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Analysis (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Ressource is logically deleted with DELETED status
	if !d.IsNewResource() && aws.StringValue(out.Status) == quicksight.ResourceStatusDeleted {
		log.Printf("[WARN] QuickSight Analysis (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionReading, ResNameAnalysis, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set(names.AttrCreatedTime, out.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, out.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Status)
	d.Set("analysis_id", out.AnalysisId)

	descResp, err := conn.DescribeAnalysisDefinitionWithContext(ctx, &quicksight.DescribeAnalysisDefinitionInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Analysis (%s) Definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenAnalysisDefinition(descResp.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	permsResp, err := conn.DescribeAnalysisPermissionsWithContext(ctx, &quicksight.DescribeAnalysisPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Analysis (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceAnalysisUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, analysisId, err := ParseAnalysisId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		in := &quicksight.UpdateAnalysisInput{
			AwsAccountId: aws.String(awsAccountId),
			AnalysisId:   aws.String(analysisId),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		_, createdFromEntity := d.GetOk("source_entity")
		if createdFromEntity {
			in.SourceEntity = quicksightschema.ExpandAnalysisSourceEntity(d.Get("source_entity").([]interface{}))
		} else {
			in.Definition = quicksightschema.ExpandAnalysisDefinition(d.Get("definition").([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]interface{}))
		}

		log.Printf("[DEBUG] Updating QuickSight Analysis (%s): %#v", d.Id(), in)
		_, err := conn.UpdateAnalysisWithContext(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionUpdating, ResNameAnalysis, d.Id(), err)
		}

		if _, err := waitAnalysisUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForUpdate, ResNameAnalysis, d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		oraw, nraw := d.GetChange(names.AttrPermissions)
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		toGrant, toRevoke := DiffPermissions(o.List(), n.List())

		params := &quicksight.UpdateAnalysisPermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			AnalysisId:   aws.String(analysisId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateAnalysisPermissionsWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Analysis (%s) permissions: %s", analysisId, err)
		}
	}

	return append(diags, resourceAnalysisRead(ctx, d, meta)...)
}

func resourceAnalysisDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, analysisId, err := ParseAnalysisId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &quicksight.DeleteAnalysisInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	}

	recoveryWindowInDays := d.Get("recovery_window_in_days").(int)
	if recoveryWindowInDays == 0 {
		input.ForceDeleteWithoutRecovery = aws.Bool(true)
	} else {
		input.RecoveryWindowInDays = aws.Int64(int64(recoveryWindowInDays))
	}

	log.Printf("[INFO] Deleting QuickSight Analysis %s", d.Id())
	_, err = conn.DeleteAnalysisWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionDeleting, ResNameAnalysis, d.Id(), err)
	}

	return diags
}

func FindAnalysisByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.Analysis, error) {
	awsAccountId, analysisId, err := ParseAnalysisId(id)
	if err != nil {
		return nil, err
	}

	descOpts := &quicksight.DescribeAnalysisInput{
		AwsAccountId: aws.String(awsAccountId),
		AnalysisId:   aws.String(analysisId),
	}

	out, err := conn.DescribeAnalysisWithContext(ctx, descOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: descOpts,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Analysis == nil {
		return nil, tfresource.NewEmptyResultError(descOpts)
	}

	return out.Analysis, nil
}

func ParseAnalysisId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,ANALYSIS_ID", id)
	}
	return parts[0], parts[1], nil
}

func createAnalysisId(awsAccountID, analysisId string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, analysisId)
}
