// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	instanceProfileNameMaxLen       = 128
	instanceProfileNamePrefixMaxLen = instanceProfileNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_instance_profile", name="Instance Profile")
// @Tags(identifierAttribute="id", resourceType="InstanceProfile")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/iam.InstanceProfile")
func resourceInstanceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceProfileCreate,
		ReadWithoutTimeout:   resourceInstanceProfileRead,
		UpdateWithoutTimeout: resourceInstanceProfileUpdate,
		DeleteWithoutTimeout: resourceInstanceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validResourceName(instanceProfileNameMaxLen),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(instanceProfileNamePrefixMaxLen),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceInstanceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(name),
		Path:                aws.String(d.Get("path").(string)),
		Tags:                getTagsIn(ctx),
	}

	output, err := conn.CreateInstanceProfileWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateInstanceProfileWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Instance Profile (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.InstanceProfile.InstanceProfileName))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findInstanceProfileByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM Instance Profile (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("role"); ok {
		err := instanceProfileAddRole(ctx, conn, d.Id(), v.(string))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := instanceProfileCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Instance Profile (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	instanceProfile, err := findInstanceProfileByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Instance Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profile (%s): %s", d.Id(), err)
	}

	if len(instanceProfile.Roles) > 0 {
		roleName := aws.StringValue(instanceProfile.Roles[0].RoleName)
		_, err := findRoleByName(ctx, conn, roleName)

		if err != nil {
			if tfresource.NotFound(err) {
				err := instanceProfileRemoveRole(ctx, conn, d.Id(), roleName)

				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}

			return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s) attached to IAM Instance Profile (%s): %s", roleName, d.Id(), err)
		}
	}

	d.Set("arn", instanceProfile.Arn)
	d.Set("create_date", instanceProfile.CreateDate.Format(time.RFC3339))
	d.Set("name", instanceProfile.InstanceProfileName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(instanceProfile.InstanceProfileName)))
	d.Set("path", instanceProfile.Path)
	if len(instanceProfile.Roles) > 0 {
		d.Set("role", instanceProfile.Roles[0].RoleName) //there will only be 1 role returned
	}
	d.Set("unique_id", instanceProfile.InstanceProfileId)

	setTagsOut(ctx, instanceProfile.Tags)

	return diags
}

func resourceInstanceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if d.HasChange("role") {
		o, n := d.GetChange("role")

		if o := o.(string); o != "" {
			err := instanceProfileRemoveRole(ctx, conn, d.Id(), o)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if n := n.(string); n != "" {
			err := instanceProfileAddRole(ctx, conn, d.Id(), n)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if v, ok := d.GetOk("role"); ok {
		err := instanceProfileRemoveRole(ctx, conn, d.Id(), v.(string))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting IAM Instance Profile: %s", d.Id())
	_, err := conn.DeleteInstanceProfileWithContext(ctx, &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Instance Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func instanceProfileAddRole(ctx context.Context, conn *iam.IAM, profileName, roleName string) error {
	input := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.AddRoleToInstanceProfileWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			// IAM unfortunately does not provide a better error code or message for eventual consistency
			// InvalidParameterValue: Value (XXX) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
			// NoSuchEntity: The request was rejected because it referenced an entity that does not exist. The error message describes the entity. HTTP Status Code: 404
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Invalid IAM Instance Profile name") || tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "The role with name") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("adding IAM Role (%s) to IAM Instance Profile (%s): %w", roleName, profileName, err)
	}

	return nil
}

func instanceProfileRemoveRole(ctx context.Context, conn *iam.IAM, profileName, roleName string) error {
	input := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := conn.RemoveRoleFromInstanceProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing IAM Role (%s) from IAM Instance Profile (%s): %w", roleName, profileName, err)
	}

	return nil
}

func findInstanceProfileByName(ctx context.Context, conn *iam.IAM, name string) (*iam.InstanceProfile, error) {
	input := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}

	output, err := conn.GetInstanceProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InstanceProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InstanceProfile, nil
}

func instanceProfileTags(ctx context.Context, conn *iam.IAM, identifier string) ([]*iam.Tag, error) {
	output, err := conn.ListInstanceProfileTagsWithContext(ctx, &iam.ListInstanceProfileTagsInput{
		InstanceProfileName: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
