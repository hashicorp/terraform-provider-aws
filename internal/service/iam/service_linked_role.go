// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_service_linked_role", name="Service Linked Role")
// @Tags(identifierAttribute="id", resourceType="ServiceLinkedRole")
func resourceServiceLinkedRole() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceLinkedRoleCreate,
		ReadWithoutTimeout:   resourceServiceLinkedRoleRead,
		UpdateWithoutTimeout: resourceServiceLinkedRoleUpdate,
		DeleteWithoutTimeout: resourceServiceLinkedRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`\.`), "must be a full service hostname e.g. elasticbeanstalk.amazonaws.com"),
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_suffix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.Contains(d.Get("aws_service_name").(string), ".application-autoscaling.") && new == "" {
						return true
					}
					return false
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceLinkedRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	serviceName := d.Get("aws_service_name").(string)
	input := &iam.CreateServiceLinkedRoleInput{
		AWSServiceName: aws.String(serviceName),
	}

	if v, ok := d.GetOk("custom_suffix"); ok {
		input.CustomSuffix = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateServiceLinkedRole(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Service Linked Role (%s): %s", serviceName, err)
	}

	d.SetId(aws.ToString(output.Role.Arn))

	if tags := getTagsIn(ctx); len(tags) > 0 {
		_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		err = roleUpdateTags(ctx, conn, roleName, nil, KeyValueTags(ctx, tags))

		// If default tags only, continue. Otherwise, error.
		partition := meta.(*conns.AWSClient).Partition
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceServiceLinkedRoleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Service Linked Role (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceLinkedRoleRead(ctx, d, meta)...)
}

func resourceServiceLinkedRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	serviceName, roleName, customSuffix, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findRoleByName(ctx, conn, roleName)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Service Linked Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Linked Role (%s): %s", d.Id(), err)
	}

	role := outputRaw.(*awstypes.Role)

	d.Set(names.AttrARN, role.Arn)
	d.Set("aws_service_name", serviceName)
	d.Set("create_date", aws.ToTime(role.CreateDate).Format(time.RFC3339))
	d.Set("custom_suffix", customSuffix)
	d.Set(names.AttrDescription, role.Description)
	d.Set(names.AttrName, role.RoleName)
	d.Set(names.AttrPath, role.Path)
	d.Set("unique_id", role.RoleId)

	setTagsOut(ctx, role.Tags)

	return diags
}

func resourceServiceLinkedRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChangesExcept(names.AttrTagsAll, names.AttrTags) {
		_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &iam.UpdateRoleInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			RoleName:    aws.String(roleName),
		}

		_, err = conn.UpdateRole(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Service Linked Role (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceLinkedRoleRead(ctx, d, meta)...)
}

func resourceServiceLinkedRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IAM Service Linked Role: %s", d.Id())
	if err := deleteServiceLinkedRole(ctx, conn, roleName); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteServiceLinkedRole(ctx context.Context, conn *iam.Client, roleName string) error {
	input := &iam.DeleteServiceLinkedRoleInput{
		RoleName: aws.String(roleName),
	}

	output, err := conn.DeleteServiceLinkedRole(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IAM Service Linked Role (%s): %w", roleName, err)
	}

	deletionTaskID := aws.ToString(output.DeletionTaskId)

	if deletionTaskID == "" {
		return nil
	}

	if err := waitServiceLinkedRoleDeleted(ctx, conn, deletionTaskID); err != nil {
		return fmt.Errorf("waiting for IAM Service Linked Role (%s) delete: %w", roleName, err)
	}

	return nil
}

func waitServiceLinkedRoleDeleted(ctx context.Context, conn *iam.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DeletionTaskStatusTypeInProgress, awstypes.DeletionTaskStatusTypeNotStarted),
		Target:  enum.Slice(awstypes.DeletionTaskStatusTypeSucceeded),
		Refresh: statusServiceLinkedRoleDeletion(ctx, conn, id),
		Timeout: 5 * time.Minute,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iam.GetServiceLinkedRoleDeletionStatusOutput); ok {
		if reason := output.Reason; reason != nil {
			var errs []error

			for _, v := range reason.RoleUsageList {
				errs = append(errs, fmt.Errorf("%s: %s", aws.ToString(v.Region), strings.Join(v.Resources, ", ")))
			}

			tfresource.SetLastError(err, fmt.Errorf("%s: %w", aws.ToString(reason.Reason), errors.Join(errs...)))
		}

		return err
	}

	return err
}

func statusServiceLinkedRoleDeletion(ctx context.Context, conn *iam.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServiceLinkedRoleDeletionStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findServiceLinkedRoleDeletionStatusByID(ctx context.Context, conn *iam.Client, id string) (*iam.GetServiceLinkedRoleDeletionStatusOutput, error) {
	input := &iam.GetServiceLinkedRoleDeletionStatusInput{
		DeletionTaskId: aws.String(id),
	}

	output, err := conn.GetServiceLinkedRoleDeletionStatus(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
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

func DecodeServiceLinkedRoleID(id string) (serviceName, roleName, customSuffix string, err error) {
	idArn, err := arn.Parse(id)

	if err != nil {
		return "", "", "", err
	}

	resourceParts := strings.Split(idArn.Resource, "/")

	if len(resourceParts) != 4 {
		return "", "", "", fmt.Errorf("expected IAM Service Role ARN (arn:PARTITION:iam::ACCOUNTID:role/aws-service-role/SERVICENAME/ROLENAME), received: %s", id)
	}

	serviceName = resourceParts[2]
	roleName = resourceParts[3]

	roleNameParts := strings.Split(roleName, "_")
	if len(roleNameParts) == 2 {
		customSuffix = roleNameParts[1]
	}

	return
}
