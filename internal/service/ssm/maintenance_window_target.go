// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_maintenance_window_target", name="Maintenance Window Target")
func resourceMaintenanceWindowTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMaintenanceWindowTargetCreate,
		ReadWithoutTimeout:   resourceMaintenanceWindowTargetRead,
		UpdateWithoutTimeout: resourceMaintenanceWindowTargetUpdate,
		DeleteWithoutTimeout: resourceMaintenanceWindowTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%s), expected WINDOW_ID/WINDOW_TARGET_ID", d.Id())
				}
				d.Set("window_id", idParts[0])
				d.SetId(idParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,128}$`), "Only alphanumeric characters, hyphens, dots & underscores allowed"),
			},
			"owner_information": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"targets": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^[\p{L}\p{Z}\p{N}_.:/=\-@]*$|resource-groups:ResourceTypeFilters|resource-groups:Name$`), ""),
								validation.StringLenBetween(1, 163),
							),
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrResourceType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MaintenanceWindowResourceType](),
			},
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceMaintenanceWindowTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.RegisterTargetWithMaintenanceWindowInput{
		ResourceType: awstypes.MaintenanceWindowResourceType(d.Get(names.AttrResourceType).(string)),
		Targets:      expandTargets(d.Get("targets").([]interface{})),
		WindowId:     aws.String(d.Get("window_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner_information"); ok {
		input.OwnerInformation = aws.String(v.(string))
	}

	output, err := conn.RegisterTargetWithMaintenanceWindow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Maintenance Window Target: %s", err)
	}

	d.SetId(aws.ToString(output.WindowTargetId))

	return append(diags, resourceMaintenanceWindowTargetRead(ctx, d, meta)...)
}

func resourceMaintenanceWindowTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	windowID := d.Get("window_id").(string)
	target, err := findMaintenanceWindowTargetByTwoPartKey(ctx, conn, windowID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Maintenance Window Target %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Maintenance Window Target (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, target.Description)
	d.Set(names.AttrName, target.Name)
	d.Set("owner_information", target.OwnerInformation)
	d.Set(names.AttrResourceType, target.ResourceType)
	if err := d.Set("targets", flattenTargets(target.Targets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting targets: %s", err)
	}
	d.Set("window_id", target.WindowId)

	return diags
}

func resourceMaintenanceWindowTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.UpdateMaintenanceWindowTargetInput{
		Targets:        expandTargets(d.Get("targets").([]interface{})),
		WindowId:       aws.String(d.Get("window_id").(string)),
		WindowTargetId: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrName) {
		input.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("owner_information") {
		input.OwnerInformation = aws.String(d.Get("owner_information").(string))
	}

	_, err := conn.UpdateMaintenanceWindowTarget(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSM Maintenance Window Target (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMaintenanceWindowTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[INFO] Deleting SSM Maintenance Window Target: %s", d.Id())
	_, err := conn.DeregisterTargetFromMaintenanceWindow(ctx, &ssm.DeregisterTargetFromMaintenanceWindowInput{
		WindowId:       aws.String(d.Get("window_id").(string)),
		WindowTargetId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Maintenance Window Target (%s): %s", d.Id(), err)
	}

	return diags
}

func findMaintenanceWindowTargetByTwoPartKey(ctx context.Context, conn *ssm.Client, windowID, windowTargetID string) (*awstypes.MaintenanceWindowTarget, error) {
	input := &ssm.DescribeMaintenanceWindowTargetsInput{
		Filters: []awstypes.MaintenanceWindowFilter{
			{
				Key:    aws.String("WindowTargetId"),
				Values: []string{windowTargetID},
			},
		},
		WindowId: aws.String(windowID),
	}

	return findMaintenanceWindowTarget(ctx, conn, input)
}

func findMaintenanceWindowTarget(ctx context.Context, conn *ssm.Client, input *ssm.DescribeMaintenanceWindowTargetsInput) (*awstypes.MaintenanceWindowTarget, error) {
	output, err := findMaintenanceWindowTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMaintenanceWindowTargets(ctx context.Context, conn *ssm.Client, input *ssm.DescribeMaintenanceWindowTargetsInput) ([]awstypes.MaintenanceWindowTarget, error) {
	var output []awstypes.MaintenanceWindowTarget

	pages := ssm.NewDescribeMaintenanceWindowTargetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Targets...)
	}

	return output, nil
}
