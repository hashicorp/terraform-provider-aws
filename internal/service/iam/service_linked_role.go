// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_service_linked_role", name="Service Linked Role")
// @Tags
func ResourceServiceLinkedRole() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceLinkedRoleCreate,
		ReadWithoutTimeout:   resourceServiceLinkedRoleRead,
		UpdateWithoutTimeout: resourceServiceLinkedRoleUpdate,
		DeleteWithoutTimeout: resourceServiceLinkedRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\.`), "must be a full service hostname e.g. elasticbeanstalk.amazonaws.com"),
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
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
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	serviceName := d.Get("aws_service_name").(string)
	input := &iam.CreateServiceLinkedRoleInput{
		AWSServiceName: aws.String(serviceName),
	}

	if v, ok := d.GetOk("custom_suffix"); ok {
		input.CustomSuffix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateServiceLinkedRoleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Service Linked Role (%s): %s", serviceName, err)
	}

	d.SetId(aws.StringValue(output.Role.Arn))

	if tags := getTagsIn(ctx); len(tags) > 0 {
		_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		err = roleUpdateTags(ctx, conn, roleName, nil, KeyValueTags(ctx, tags))

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	serviceName, roleName, customSuffix, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindRoleByName(ctx, conn, roleName)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Service Linked Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Linked Role (%s): %s", d.Id(), err)
	}

	role := outputRaw.(*iam.Role)

	d.Set("arn", role.Arn)
	d.Set("aws_service_name", serviceName)
	d.Set("create_date", aws.TimeValue(role.CreateDate).Format(time.RFC3339))
	d.Set("custom_suffix", customSuffix)
	d.Set("description", role.Description)
	d.Set("name", role.RoleName)
	d.Set("path", role.Path)
	d.Set("unique_id", role.RoleId)

	setTagsOut(ctx, role.Tags)

	return diags
}

func resourceServiceLinkedRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &iam.UpdateRoleInput{
			Description: aws.String(d.Get("description").(string)),
			RoleName:    aws.String(roleName),
		}

		_, err = conn.UpdateRoleWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Service Linked Role (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := roleUpdateTags(ctx, conn, roleName, o, n)

		// Some partitions (e.g. ISO) may not support tagging.
		if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceServiceLinkedRoleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Service Linked Role (%s): updating tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceLinkedRoleRead(ctx, d, meta)...)
}

func resourceServiceLinkedRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IAM Service Linked Role: %s", d.Id())
	output, err := conn.DeleteServiceLinkedRoleWithContext(ctx, &iam.DeleteServiceLinkedRoleInput{
		RoleName: aws.String(roleName),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Service Linked Role (%s): %s", d.Id(), err)
	}

	deletionTaskID := aws.StringValue(output.DeletionTaskId)

	if deletionTaskID == "" {
		return diags
	}

	if err := waitServiceLinkedRoleDeleted(ctx, conn, deletionTaskID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM Service Linked Role (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitServiceLinkedRoleDeleted(ctx context.Context, conn *iam.IAM, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{iam.DeletionTaskStatusTypeInProgress, iam.DeletionTaskStatusTypeNotStarted},
		Target:  []string{iam.DeletionTaskStatusTypeSucceeded},
		Refresh: statusServiceLinkedRoleDeletion(ctx, conn, id),
		Timeout: 5 * time.Minute,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iam.GetServiceLinkedRoleDeletionStatusOutput); ok {
		if reason := output.Reason; reason != nil {
			var errs *multierror.Error

			for _, v := range reason.RoleUsageList {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(v.Region), strings.Join(aws.StringValueSlice(v.Resources), ", ")))
			}

			tfresource.SetLastError(err, fmt.Errorf("%s: %w", aws.StringValue(reason.Reason), errs.ErrorOrNil()))
		}

		return err
	}

	return err
}

func statusServiceLinkedRoleDeletion(ctx context.Context, conn *iam.IAM, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServiceLinkedRoleDeletionStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func findServiceLinkedRoleDeletionStatusByID(ctx context.Context, conn *iam.IAM, id string) (*iam.GetServiceLinkedRoleDeletionStatusOutput, error) {
	input := &iam.GetServiceLinkedRoleDeletionStatusInput{
		DeletionTaskId: aws.String(id),
	}

	output, err := conn.GetServiceLinkedRoleDeletionStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
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
