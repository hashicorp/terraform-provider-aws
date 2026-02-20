// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_user", name="User")
// @Tags(identifierAttribute="id", resourceType="User")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/iam/types;types.User", importIgnore="force_destroy")
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrForceDestroy: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete user even if it has non-Terraform-managed IAM access keys, login profile or MFA devices",
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					"must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs",
				),
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"permissions_boundary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			/*
				The UniqueID could be used as the Id(), but none of the API
				calls allow specifying a user by the UniqueID: they require the
				name. The only way to locate a user by UniqueID is to list them
				all and that would make this provider unnecessarily complex
				and inefficient. Still, there are other reasons one might want
				the UniqueID, so we can make it available.
			*/
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	path := d.Get(names.AttrPath).(string)
	input := iam.CreateUserInput{
		Path:     aws.String(path),
		Tags:     getTagsIn(ctx),
		UserName: aws.String(name),
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		input.PermissionsBoundary = aws.String(v.(string))
	}

	output, err := retryCreateUser(ctx, conn, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition(ctx)
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryCreateUser(ctx, conn, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM User (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.User.UserName))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := userCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceUserRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM User (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	user, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.User, error) {
		return findUserByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set(names.AttrName, user.UserName)
	d.Set(names.AttrPath, user.Path)
	if user.PermissionsBoundary != nil {
		d.Set("permissions_boundary", user.PermissionsBoundary.PermissionsBoundaryArn)
	} else {
		d.Set("permissions_boundary", nil)
	}
	d.Set("unique_id", user.UserId)

	setTagsOut(ctx, user.Tags)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChanges(names.AttrName, names.AttrPath) {
		o, n := d.GetChange(names.AttrName)
		input := iam.UpdateUserInput{
			NewUserName: aws.String(n.(string)),
			NewPath:     aws.String(d.Get(names.AttrPath).(string)),
			UserName:    aws.String(o.(string)),
		}

		_, err := conn.UpdateUser(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User (%s): %s", d.Id(), err)
		}

		d.SetId(n.(string))
	}

	if d.HasChange("permissions_boundary") {
		if v, ok := d.GetOk("permissions_boundary"); ok {
			input := iam.PutUserPermissionsBoundaryInput{
				PermissionsBoundary: aws.String(v.(string)),
				UserName:            aws.String(d.Id()),
			}

			_, err := conn.PutUserPermissionsBoundary(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting IAM User (%s) permissions boundary: %s", d.Id(), err)
			}
		} else {
			input := iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String(d.Id()),
			}
			_, err := conn.DeleteUserPermissionsBoundary(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IAM User (%s) permissions boundary: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	// IAM Users must be removed from all groups before they can be deleted.
	if err := deleteUserGroupMemberships(ctx, conn, d.Id()); err != nil {
		if !errs.IsA[*awstypes.NoSuchEntityException](err) {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) group memberships: %s", d.Id(), err)
		}
	}

	// All access keys, MFA devices and login profile for the user must be removed.
	if d.Get(names.AttrForceDestroy).(bool) {
		for _, v := range []struct {
			f      func(context.Context, *iam.Client, string) error
			format string
		}{
			{deleteUserPolicies, "removing IAM User (%s) policies: %s"},
			{detachUserPolicies, "detaching IAM User (%s) policies: %s"},
			{deleteUserAccessKeys, "removing IAM User (%s) access keys: %s"},
			{deleteUserSSHKeys, "removing IAM User (%s) access keys: %s"},
			{deleteUserVirtualMFADevices, "removing IAM User (%s) Virtual MFA devices: %s"},
			{deactivateUserMFADevices, "removing IAM User (%s) MFA devices: %s"},
			{deleteUserLoginProfile, "removing IAM User (%s) login profile: %s"},
			{deleteUserSigningCertificates, "removing IAM User (%s) signing certificate: %s"},
			{deleteServiceSpecificCredentials, "removing IAM User (%s) Service Specific Credentials: %s"},
		} {
			if err := v.f(ctx, conn, d.Id()); err != nil {
				if !errs.IsA[*awstypes.NoSuchEntityException](err) {
					return sdkdiag.AppendErrorf(diags, v.format, d.Id(), err)
				}
			}
		}
	}

	log.Println("[DEBUG] Deleting IAM User:", d.Id())
	input := iam.DeleteUserInput{
		UserName: aws.String(d.Id()),
	}
	_, err := conn.DeleteUser(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserByName(ctx context.Context, conn *iam.Client, name string) (*awstypes.User, error) {
	input := iam.GetUserInput{
		UserName: aws.String(name),
	}

	return findUser(ctx, conn, &input)
}

func findUser(ctx context.Context, conn *iam.Client, input *iam.GetUserInput) (*awstypes.User, error) {
	output, err := conn.GetUser(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.User, nil
}

func deleteUserGroupMemberships(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListGroupsForUserInput{
		UserName: aws.String(user),
	}
	var groupNames []string

	pages := iam.NewListGroupsForUserPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) groups: %w", user, err)
		}

		for _, v := range page.Groups {
			groupNames = append(groupNames, aws.ToString(v.GroupName))
		}
	}

	for _, groupName := range groupNames {
		// use iam group membership func to remove user from all groups
		log.Printf("[DEBUG] Removing IAM User %s from IAM Group %s", user, groupName)
		if err := removeUsersFromGroup(ctx, conn, []string{user}, groupName); err != nil {
			return err
		}
	}

	return nil
}

func deleteUserSSHKeys(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListSSHPublicKeysInput{
		UserName: aws.String(user),
	}
	var ids []string

	pages := iam.NewListSSHPublicKeysPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) SSH public keys: %w", user, err)
		}

		for _, v := range page.SSHPublicKeys {
			ids = append(ids, aws.ToString(v.SSHPublicKeyId))
		}
	}

	for _, v := range ids {
		input := iam.DeleteSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(v),
			UserName:       aws.String(user),
		}
		_, err := conn.DeleteSSHPublicKey(ctx, &input)

		if err != nil {
			return fmt.Errorf("deleting IAM User (%s) SSH public key (%s): %w", user, v, err)
		}
	}

	return nil
}

func deleteUserVirtualMFADevices(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListVirtualMFADevicesInput{
		AssignmentStatus: awstypes.AssignmentStatusTypeAssigned,
	}
	var virtualMFADevices []string

	pages := iam.NewListVirtualMFADevicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM Virtual MFA Devices: %w", err)
		}

		for _, v := range page.VirtualMFADevices {
			// UserName is `nil` for the root user
			if aws.ToString(v.User.UserName) == user {
				virtualMFADevices = append(virtualMFADevices, aws.ToString(v.SerialNumber))
			}
		}
	}

	for _, v := range virtualMFADevices {
		inputDeactivate := iam.DeactivateMFADeviceInput{
			SerialNumber: aws.String(v),
			UserName:     aws.String(user),
		}
		_, err := conn.DeactivateMFADevice(ctx, &inputDeactivate)

		if err != nil {
			return fmt.Errorf("deactivating IAM User (%s) virtual MFA device (%s): %w", user, v, err)
		}

		inputDelete := iam.DeleteVirtualMFADeviceInput{
			SerialNumber: aws.String(v),
		}
		_, err = conn.DeleteVirtualMFADevice(ctx, &inputDelete)

		if err != nil {
			return fmt.Errorf("deleting IAM Virtual MFA Device (%s): %w", v, err)
		}
	}

	return nil
}

func deactivateUserMFADevices(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListMFADevicesInput{
		UserName: aws.String(user),
	}
	var mfaDevices []string

	pages := iam.NewListMFADevicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) MFA devices: %w", user, err)
		}

		for _, v := range page.MFADevices {
			mfaDevices = append(mfaDevices, aws.ToString(v.SerialNumber))
		}
	}

	for _, v := range mfaDevices {
		input := iam.DeactivateMFADeviceInput{
			SerialNumber: aws.String(v),
			UserName:     aws.String(user),
		}
		_, err := conn.DeactivateMFADevice(ctx, &input)

		if err != nil {
			return fmt.Errorf("deactivating IAM User (%s) MFA device (%s): %w", user, v, err)
		}
	}

	return nil
}

func deleteUserLoginProfile(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.DeleteLoginProfileInput{
		UserName: aws.String(user),
	}

	err := tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.DeleteLoginProfile(ctx, &input)
		if err != nil {
			var errNoSuchEntityException *awstypes.NoSuchEntityException
			if tfawserr.ErrCodeEquals(err, errNoSuchEntityException.ErrorCode()) {
				return nil
			}
			// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
			var etu *awstypes.EntityTemporarilyUnmodifiableException
			if tfawserr.ErrCodeEquals(err, etu.ErrorCode()) {
				return tfresource.RetryableError(err)
			}
			return tfresource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("deleting IAM User (%s) login profile: %w", user, err)
	}

	return nil
}

func deleteUserAccessKeys(ctx context.Context, conn *iam.Client, user string) error {
	accessKeys, err := findAccessKeysByUser(ctx, conn, user)

	if err != nil && !retry.NotFound(err) {
		return fmt.Errorf("listing IAM User (%s) access keys: %w", user, err)
	}

	var errs []error
	for _, v := range accessKeys {
		accessKeyID := aws.ToString(v.AccessKeyId)
		input := iam.DeleteAccessKeyInput{
			AccessKeyId: aws.String(accessKeyID),
			UserName:    aws.String(user),
		}
		_, err := conn.DeleteAccessKey(ctx, &input)

		if err != nil {
			return fmt.Errorf("deleting IAM User (%s) access key (%s): %w", user, accessKeyID, err)
		}
	}

	return errors.Join(errs...)
}

func deleteUserSigningCertificates(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListSigningCertificatesInput{
		UserName: aws.String(user),
	}
	var ids []string

	pages := iam.NewListSigningCertificatesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) signing certificates: %w", user, err)
		}

		for _, v := range page.Certificates {
			ids = append(ids, aws.ToString(v.CertificateId))
		}
	}

	for _, v := range ids {
		input := iam.DeleteSigningCertificateInput{
			CertificateId: aws.String(v),
			UserName:      aws.String(user),
		}
		_, err := conn.DeleteSigningCertificate(ctx, &input)

		if err != nil {
			return fmt.Errorf("deleting IAM User (%s) signing certificate (%s): %w", user, v, err)
		}
	}

	return nil
}

func deleteServiceSpecificCredentials(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListServiceSpecificCredentialsInput{
		UserName: aws.String(user),
	}
	var ids []string

	err := listServiceSpecificCredentialsPages(ctx, conn, &input, func(page *iam.ListServiceSpecificCredentialsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ServiceSpecificCredentials {
			ids = append(ids, aws.ToString(v.ServiceSpecificCredentialId))
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("listing IAM User (%s) service-specific credentials: %w", user, err)
	}

	for _, v := range ids {
		input := iam.DeleteServiceSpecificCredentialInput{
			ServiceSpecificCredentialId: aws.String(v),
			UserName:                    aws.String(user),
		}
		_, err := conn.DeleteServiceSpecificCredential(ctx, &input)

		if err != nil {
			return fmt.Errorf("deleting IAM User (%s) service-specific credential (%s): %w", user, v, err)
		}
	}

	return nil
}

func deleteUserPolicies(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListUserPoliciesInput{
		UserName: aws.String(user),
	}
	var policies []string

	pages := iam.NewListUserPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) policies: %w", user, err)
		}

		policies = append(policies, page.PolicyNames...)
	}

	for _, v := range policies {
		input := iam.DeleteUserPolicyInput{
			PolicyName: aws.String(v),
			UserName:   aws.String(user),
		}

		_, err := conn.DeleteUserPolicy(ctx, &input)

		if err != nil {
			return fmt.Errorf("deleting IAM User (%s) policy (%s): %w", user, v, err)
		}
	}

	return nil
}

func detachUserPolicies(ctx context.Context, conn *iam.Client, user string) error {
	input := iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(user),
	}
	var policies []string

	pages := iam.NewListAttachedUserPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing IAM User (%s) attached policies: %w", user, err)
		}

		for _, v := range page.AttachedPolicies {
			policies = append(policies, aws.ToString(v.PolicyArn))
		}
	}

	for _, v := range policies {
		if err := detachPolicyFromUser(ctx, conn, user, v); err != nil {
			return err
		}
	}

	return nil
}

func userTags(ctx context.Context, conn *iam.Client, identifier string, optFns ...func(*iam.Options)) ([]awstypes.Tag, error) {
	input := iam.ListUserTagsInput{
		UserName: aws.String(identifier),
	}
	var output []awstypes.Tag

	pages := iam.NewListUserTagsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Tags...)
	}

	return output, nil
}

func retryCreateUser(ctx context.Context, conn *iam.Client, input *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateUser(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*iam.CreateUserOutput)
	if !ok || output == nil || aws.ToString(output.User.UserName) == "" {
		return nil, fmt.Errorf("create IAM user (%s) returned an empty result", aws.ToString(input.UserName))
	}

	return output, err
}
