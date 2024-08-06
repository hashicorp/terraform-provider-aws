// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

// @SDKResource("aws_quicksight_dashboard", name="Dashboard")
// @Tags(identifierAttribute="arn")
func ResourceDashboard() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDashboardCreate,
		ReadWithoutTimeout:   resourceDashboardRead,
		UpdateWithoutTimeout: resourceDashboardUpdate,
		DeleteWithoutTimeout: resourceDashboardDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				"dashboard_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"dashboard_publish_options": quicksightschema.DashboardPublishOptionsSchema(),
				"definition":                quicksightschema.DashboardDefinitionSchema(),
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
				"source_entity": quicksightschema.DashboardSourceEntitySchema(),
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
				"theme_arn": {
					Type:     schema.TypeString,
					Optional: true,
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
	ResNameDashboard             = "Dashboard"
	DashboardLatestVersion int64 = -1
)

func resourceDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}
	dashboardId := d.Get("dashboard_id").(string)

	d.SetId(createDashboardId(awsAccountId, dashboardId))

	input := &quicksight.CreateDashboardInput{
		AwsAccountId: aws.String(awsAccountId),
		DashboardId:  aws.String(dashboardId),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SourceEntity = quicksightschema.ExpandDashboardSourceEntity(v.([]interface{}))
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Definition = quicksightschema.ExpandDashboardDefinition(d.Get("definition").([]interface{}))
	}

	if v, ok := d.GetOk("dashboard_publish_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DashboardPublishOptions = quicksightschema.ExpandDashboardPublishOptions(d.Get("dashboard_publish_options").([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]interface{}))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = expandResourcePermissions(v.List())
	}

	_, err := conn.CreateDashboardWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameDashboard, d.Get(names.AttrName).(string), err)
	}

	if _, err := waitDashboardCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForCreation, ResNameDashboard, d.Id(), err)
	}

	return append(diags, resourceDashboardRead(ctx, d, meta)...)
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dashboardId, err := ParseDashboardId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := FindDashboardByID(ctx, conn, d.Id(), DashboardLatestVersion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Dashboard (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionReading, ResNameDashboard, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set(names.AttrCreatedTime, out.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, out.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Version.Status)
	d.Set("source_entity_arn", out.Version.SourceEntityArn)
	d.Set("dashboard_id", out.DashboardId)
	d.Set("version_description", out.Version.Description)
	d.Set("version_number", out.Version.VersionNumber)

	descResp, err := conn.DescribeDashboardDefinitionWithContext(ctx, &quicksight.DescribeDashboardDefinitionInput{
		AwsAccountId:  aws.String(awsAccountId),
		DashboardId:   aws.String(dashboardId),
		VersionNumber: out.Version.VersionNumber,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Dashboard (%s) Definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenDashboardDefinition(descResp.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	if err := d.Set("dashboard_publish_options", quicksightschema.FlattenDashboardPublishOptions(descResp.DashboardPublishOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dashboard_publish_options: %s", err)
	}

	permsResp, err := conn.DescribeDashboardPermissionsWithContext(ctx, &quicksight.DescribeDashboardPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DashboardId:  aws.String(dashboardId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Dashboard (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dashboardId, err := ParseDashboardId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		in := &quicksight.UpdateDashboardInput{
			AwsAccountId:       aws.String(awsAccountId),
			DashboardId:        aws.String(dashboardId),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			VersionDescription: aws.String(d.Get("version_description").(string)),
		}

		_, createdFromEntity := d.GetOk("source_entity")
		if createdFromEntity {
			in.SourceEntity = quicksightschema.ExpandDashboardSourceEntity(d.Get("source_entity").([]interface{}))
		} else {
			in.Definition = quicksightschema.ExpandDashboardDefinition(d.Get("definition").([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]interface{}))
		}

		if v, ok := d.GetOk("dashboard_publish_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.DashboardPublishOptions = quicksightschema.ExpandDashboardPublishOptions(d.Get("dashboard_publish_options").([]interface{}))
		}

		log.Printf("[DEBUG] Updating QuickSight Dashboard (%s): %#v", d.Id(), in)
		out, err := conn.UpdateDashboardWithContext(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionUpdating, ResNameDashboard, d.Id(), err)
		}

		updatedVersionNumber := extractVersionFromARN(aws.StringValue(out.VersionArn))
		if _, err := waitDashboardUpdated(ctx, conn, d.Id(), updatedVersionNumber, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForUpdate, ResNameDashboard, d.Id(), err)
		}

		publishVersion := &quicksight.UpdateDashboardPublishedVersionInput{
			AwsAccountId:  aws.String(awsAccountId),
			DashboardId:   aws.String(dashboardId),
			VersionNumber: aws.Int64(updatedVersionNumber),
		}
		_, err = conn.UpdateDashboardPublishedVersionWithContext(ctx, publishVersion)
		if err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionUpdating, ResNameDashboard, d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		oraw, nraw := d.GetChange(names.AttrPermissions)
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		toGrant, toRevoke := DiffPermissions(o.List(), n.List())

		params := &quicksight.UpdateDashboardPermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			DashboardId:  aws.String(dashboardId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDashboardPermissionsWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Dashboard (%s) permissions: %s", dashboardId, err)
		}
	}

	return append(diags, resourceDashboardRead(ctx, d, meta)...)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dashboardId, err := ParseDashboardId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Dashboard %s", d.Id())
	_, err = conn.DeleteDashboardWithContext(ctx, &quicksight.DeleteDashboardInput{
		AwsAccountId: aws.String(awsAccountId),
		DashboardId:  aws.String(dashboardId),
	})

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionDeleting, ResNameDashboard, d.Id(), err)
	}

	return diags
}

// Pass version as DashboardLatestVersion for latest published version, or specify a specific version if required
func FindDashboardByID(ctx context.Context, conn *quicksight.QuickSight, id string, version int64) (*quicksight.Dashboard, error) {
	awsAccountId, dashboardId, err := ParseDashboardId(id)
	if err != nil {
		return nil, err
	}

	descOpts := &quicksight.DescribeDashboardInput{
		AwsAccountId: aws.String(awsAccountId),
		DashboardId:  aws.String(dashboardId),
	}
	if version != DashboardLatestVersion {
		descOpts.VersionNumber = aws.Int64(version)
	}

	out, err := conn.DescribeDashboardWithContext(ctx, descOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: descOpts,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Dashboard == nil {
		return nil, tfresource.NewEmptyResultError(descOpts)
	}

	return out.Dashboard, nil
}

func ParseDashboardId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,DASHBOARD_ID", id)
	}
	return parts[0], parts[1], nil
}

func createDashboardId(awsAccountID, dashboardId string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, dashboardId)
}

func extractVersionFromARN(arn string) int64 {
	version, _ := strconv.Atoi(arn[strings.LastIndex(arn, "/")+1:])
	return int64(version)
}
