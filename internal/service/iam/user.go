// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	path := d.Get(names.AttrPath).(string)
	input := &iam.CreateUserInput{
		Path:     aws.String(path),
		Tags:     getTagsIn(ctx),
		UserName: aws.String(name),
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		input.PermissionsBoundary = aws.String(v.(string))
	}

	output, err := conn.CreateUser(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateUser(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM User (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.User.UserName))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := userCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceUserRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM User (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findUserByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User (%s): %s", d.Id(), err)
	}

	user := outputRaw.(*awstypes.User)

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

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChanges(names.AttrName, names.AttrPath) {
		o, n := d.GetChange(names.AttrName)
		input := &iam.UpdateUserInput{
			UserName:    aws.String(o.(string)),
			NewUserName: aws.String(n.(string)),
			NewPath:     aws.String(d.Get(names.AttrPath).(string)),
		}

		_, err := conn.UpdateUser(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User (%s): %s", d.Id(), err)
		}

		d.SetId(n.(string))
	}

	if d.HasChange("permissions_boundary") {
		if v, ok := d.GetOk("permissions_boundary"); ok {
			input := &iam.PutUserPermissionsBoundaryInput{
				PermissionsBoundary: aws.String(v.(string)),
				UserName:            aws.String(d.Id()),
			}

			_, err := conn.PutUserPermissionsBoundary(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting IAM User (%s) permissions boundary: %s", d.Id(), err)
			}
		} else {
			input := &iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String(d.Id()),
			}
			_, err := conn.DeleteUserPermissionsBoundary(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IAM User (%s) permissions boundary: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_, err := conn.DeleteUser(ctx, &iam.DeleteUserInput{
		UserName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserByName(ctx context.Context, conn *iam.Client, name string) (*awstypes.User, error) {
	input := &iam.GetUserInput{
		UserName: aws.String(name),
	}

	output, err := conn.GetUser(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.User, nil
}

func deleteUserGroupMemberships(ctx context.Context, conn *iam.Client, user string) error {
	input := &iam.ListGroupsForUserInput{
		UserName: aws.String(user),
	}
	var groupNames []string

	pages := iam.NewListGroupsForUserPaginator(conn, input)
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

func deleteUserSSHKeys(ctx context.Context, conn *iam.Client, username string) error {
	var publicKeys []string

	listSSHPublicKeys := &iam.ListSSHPublicKeysInput{
		UserName: aws.String(username),
	}

	pages := iam.NewListSSHPublicKeysPaginator(conn, listSSHPublicKeys)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("removing public SSH keys of user %s: %w", username, err)
		}

		for _, k := range page.SSHPublicKeys {
			publicKeys = append(publicKeys, *k.SSHPublicKeyId)
		}
	}

	for _, k := range publicKeys {
		_, err := conn.DeleteSSHPublicKey(ctx, &iam.DeleteSSHPublicKeyInput{
			UserName:       aws.String(username),
			SSHPublicKeyId: aws.String(k),
		})
		if err != nil {
			return fmt.Errorf("deleting public SSH key %s: %w", k, err)
		}
	}

	return nil
}

func deleteUserVirtualMFADevices(ctx context.Context, conn *iam.Client, username string) error {
	var VirtualMFADevices []string

	listVirtualMFADevices := &iam.ListVirtualMFADevicesInput{
		AssignmentStatus: awstypes.AssignmentStatusType("Assigned"),
	}

	pages := iam.NewListVirtualMFADevicesPaginator(conn, listVirtualMFADevices)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("removing Virtual MFA devices of user %s: %w", username, err)
		}

		for _, m := range page.VirtualMFADevices {
			// UserName is `nil` for the root user
			if aws.ToString(m.User.UserName) == username {
				VirtualMFADevices = append(VirtualMFADevices, *m.SerialNumber)
			}
		}
	}

	for _, m := range VirtualMFADevices {
		_, err := conn.DeactivateMFADevice(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("deactivating Virtual MFA device %s: %w", m, err)
		}
		_, err = conn.DeleteVirtualMFADevice(ctx, &iam.DeleteVirtualMFADeviceInput{
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("deleting Virtual MFA device %s: %w", m, err)
		}
	}

	return nil
}

func deactivateUserMFADevices(ctx context.Context, conn *iam.Client, username string) error {
	var MFADevices []string

	listMFADevices := &iam.ListMFADevicesInput{
		UserName: aws.String(username),
	}

	pages := iam.NewListMFADevicesPaginator(conn, listMFADevices)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("removing MFA devices of user %s: %w", username, err)
		}

		for _, v := range page.MFADevices {
			MFADevices = append(MFADevices, *v.SerialNumber)
		}
	}

	for _, m := range MFADevices {
		_, err := conn.DeactivateMFADevice(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("deactivating MFA device %s: %w", m, err)
		}
	}

	return nil
}

func deleteUserLoginProfile(ctx context.Context, conn *iam.Client, username string) error {
	var err error
	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	}
	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err = conn.DeleteLoginProfile(ctx, input)
		if err != nil {
			var errNoSuchEntityException *awstypes.NoSuchEntityException
			if tfawserr.ErrCodeEquals(err, errNoSuchEntityException.ErrorCode()) {
				return nil
			}
			// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
			var etu *awstypes.EntityTemporarilyUnmodifiableException
			if tfawserr.ErrCodeEquals(err, etu.ErrorCode()) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteLoginProfile(ctx, input)
	}
	if err != nil {
		return fmt.Errorf("deleting Account Login Profile: %w", err)
	}

	return nil
}

func deleteUserAccessKeys(ctx context.Context, conn *iam.Client, username string) error {
	accessKeys, err := findAccessKeysByUser(ctx, conn, username)

	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("listing access keys for IAM User (%s): %w", username, err)
	}

	var errs []error

	for _, k := range accessKeys {
		_, err := conn.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(username),
			AccessKeyId: k.AccessKeyId,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("deleting Access Key (%s) from User (%s): %w", aws.ToString(k.AccessKeyId), username, err))
		}
	}

	return errors.Join(errs...)
}

func deleteUserSigningCertificates(ctx context.Context, conn *iam.Client, userName string) error {
	var certificateIDList []string

	listInput := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	pages := iam.NewListSigningCertificatesPaginator(conn, listInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("removing signing certificates of user %s: %w", userName, err)
		}

		for _, c := range page.Certificates {
			certificateIDList = append(certificateIDList, aws.ToString(c.CertificateId))
		}
	}

	for _, c := range certificateIDList {
		_, err := conn.DeleteSigningCertificate(ctx, &iam.DeleteSigningCertificateInput{
			CertificateId: aws.String(c),
			UserName:      aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("deleting signing certificate %s: %w", c, err)
		}
	}

	return nil
}

func deleteServiceSpecificCredentials(ctx context.Context, conn *iam.Client, username string) error {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: aws.String(username),
	}

	output, err := conn.ListServiceSpecificCredentials(ctx, input)
	if err != nil {
		return fmt.Errorf("listing Service Specific Credentials of user %s: %w", username, err)
	}
	for _, m := range output.ServiceSpecificCredentials {
		_, err := conn.DeleteServiceSpecificCredential(ctx, &iam.DeleteServiceSpecificCredentialInput{
			UserName:                    aws.String(username),
			ServiceSpecificCredentialId: m.ServiceSpecificCredentialId,
		})
		if err != nil {
			return fmt.Errorf("deleting Service Specific Credentials %v: %w", m, err)
		}
	}

	return nil
}

func deleteUserPolicies(ctx context.Context, conn *iam.Client, username string) error {
	input := &iam.ListUserPoliciesInput{
		UserName: aws.String(username),
	}

	output, err := conn.ListUserPolicies(ctx, input)
	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		// user not found
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing/deleting IAM User (%s) inline policies: %s", username, err)
	}

	for _, name := range output.PolicyNames {
		log.Printf("[DEBUG] Deleting IAM User (%s) inline policy %q", username, name)

		input := &iam.DeleteUserPolicyInput{
			PolicyName: aws.String(name),
			UserName:   aws.String(username),
		}

		if _, err := conn.DeleteUserPolicy(ctx, input); err != nil {
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				continue
			}
			return fmt.Errorf("deleting IAM User (%s) inline policies: %s", username, err)
		}
	}

	return nil
}

func detachUserPolicies(ctx context.Context, conn *iam.Client, username string) error {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(username),
	}

	output, err := conn.ListAttachedUserPolicies(ctx, input)
	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		// user was an entity 2 nanoseconds ago, but now it's not
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing/detaching IAM User (%s) attached policy: %s", username, err)
	}

	for _, policy := range output.AttachedPolicies {
		policyARN := aws.ToString(policy.PolicyArn)

		log.Printf("[DEBUG] Detaching IAM User (%s) attached policy: %s", username, policyARN)

		if err := detachPolicyFromUser(ctx, conn, username, policyARN); err != nil {
			return fmt.Errorf("detaching IAM User (%s) attached policy: %s", username, err)
		}
	}

	return nil
}

func userTags(ctx context.Context, conn *iam.Client, identifier string) ([]awstypes.Tag, error) {
	output, err := conn.ListUserTags(ctx, &iam.ListUserTagsInput{
		UserName: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
