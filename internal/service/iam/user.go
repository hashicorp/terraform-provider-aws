package iam

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					"must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs",
				),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"permissions_boundary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"force_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete user even if it has non-Terraform-managed IAM access keys, login profile or MFA devices",
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)
	path := d.Get("path").(string)

	request := &iam.CreateUserInput{
		Path:     aws.String(path),
		UserName: aws.String(name),
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		request.PermissionsBoundary = aws.String(v.(string))
	}

	if len(tags) > 0 {
		request.Tags = Tags(tags.IgnoreAWS())
	}

	log.Println("[DEBUG] Create IAM User request:", request)
	createResp, err := conn.CreateUserWithContext(ctx, request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM User (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		createResp, err = conn.CreateUserWithContext(ctx, request)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM User (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(createResp.User.UserName))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID {
		err := userUpdateTags(ctx, conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM User (%s): %s", d.Id(), err)
			return append(diags, resourceUserRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM User (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	request := &iam.GetUserInput{
		UserName: aws.String(d.Id()),
	}

	var output *iam.GetUserOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.GetUserWithContext(ctx, request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetUserWithContext(ctx, request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User (%s): %s", d.Id(), err)
	}

	if output == nil || output.User == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User (%s): empty response", d.Id())
	}

	d.Set("arn", output.User.Arn)
	d.Set("name", output.User.UserName)
	d.Set("path", output.User.Path)
	if output.User.PermissionsBoundary != nil {
		d.Set("permissions_boundary", output.User.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", output.User.UserId)

	tags := KeyValueTags(output.User.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	if d.HasChanges("name", "path") {
		on, nn := d.GetChange("name")
		_, np := d.GetChange("path")

		request := &iam.UpdateUserInput{
			UserName:    aws.String(on.(string)),
			NewUserName: aws.String(nn.(string)),
			NewPath:     aws.String(np.(string)),
		}

		log.Println("[DEBUG] Update IAM User request:", request)
		_, err := conn.UpdateUserWithContext(ctx, request)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User %s: %s", d.Id(), err)
		}

		d.SetId(nn.(string))
	}

	if d.HasChange("permissions_boundary") {
		permissionsBoundary := d.Get("permissions_boundary").(string)
		if permissionsBoundary != "" {
			input := &iam.PutUserPermissionsBoundaryInput{
				PermissionsBoundary: aws.String(permissionsBoundary),
				UserName:            aws.String(d.Id()),
			}
			_, err := conn.PutUserPermissionsBoundaryWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating IAM User permissions boundary: %s", err)
			}
		} else {
			input := &iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String(d.Id()),
			}
			_, err := conn.DeleteUserPermissionsBoundaryWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IAM User permissions boundary: %s", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := userUpdateTags(ctx, conn, d.Id(), o, n)

		// Some partitions may not support tagging, giving error
		if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM User (%s): %s", d.Id(), err)
			return append(diags, resourceUserRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for IAM User (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	// IAM Users must be removed from all groups before they can be deleted
	if err := DeleteUserGroupMemberships(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) group memberships: %s", d.Id(), err)
	}

	// All access keys, MFA devices and login profile for the user must be removed
	if d.Get("force_destroy").(bool) {
		if err := DeleteUserAccessKeys(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) access keys: %s", d.Id(), err)
		}

		if err := DeleteUserSSHKeys(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) SSH keys: %s", d.Id(), err)
		}

		if err := DeleteUserVirtualMFADevices(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) Virtual MFA devices: %s", d.Id(), err)
		}

		if err := DeactivateUserMFADevices(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) MFA devices: %s", d.Id(), err)
		}

		if err := DeleteUserLoginProfile(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) login profile: %s", d.Id(), err)
		}

		if err := deleteUserSigningCertificates(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) signing certificate: %s", d.Id(), err)
		}

		if err := DeleteServiceSpecificCredentials(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing IAM User (%s) Service Specific Credentials: %s", d.Id(), err)
		}
	}

	deleteUserInput := &iam.DeleteUserInput{
		UserName: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete IAM User request:", deleteUserInput)
	_, err := conn.DeleteUserWithContext(ctx, deleteUserInput)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User %s: %s", d.Id(), err)
	}

	return diags
}

func DeleteUserGroupMemberships(ctx context.Context, conn *iam.IAM, username string) error {
	var groups []string
	listGroups := &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	}
	pageOfGroups := func(page *iam.ListGroupsForUserOutput, lastPage bool) (shouldContinue bool) {
		for _, g := range page.Groups {
			groups = append(groups, *g.GroupName)
		}
		return !lastPage
	}
	err := conn.ListGroupsForUserPagesWithContext(ctx, listGroups, pageOfGroups)
	if err != nil {
		return fmt.Errorf("Error removing user %q from all groups: %s", username, err)
	}
	for _, g := range groups {
		// use iam group membership func to remove user from all groups
		log.Printf("[DEBUG] Removing IAM User %s from IAM Group %s", username, g)
		if err := removeUsersFromGroup(ctx, conn, []string{username}, g); err != nil {
			return err
		}
	}

	return nil
}

func DeleteUserSSHKeys(ctx context.Context, conn *iam.IAM, username string) error {
	var publicKeys []string
	var err error

	listSSHPublicKeys := &iam.ListSSHPublicKeysInput{
		UserName: aws.String(username),
	}
	pageOfListSSHPublicKeys := func(page *iam.ListSSHPublicKeysOutput, lastPage bool) (shouldContinue bool) {
		for _, k := range page.SSHPublicKeys {
			publicKeys = append(publicKeys, *k.SSHPublicKeyId)
		}
		return !lastPage
	}
	err = conn.ListSSHPublicKeysPagesWithContext(ctx, listSSHPublicKeys, pageOfListSSHPublicKeys)
	if err != nil {
		return fmt.Errorf("Error removing public SSH keys of user %s: %w", username, err)
	}
	for _, k := range publicKeys {
		_, err := conn.DeleteSSHPublicKeyWithContext(ctx, &iam.DeleteSSHPublicKeyInput{
			UserName:       aws.String(username),
			SSHPublicKeyId: aws.String(k),
		})
		if err != nil {
			return fmt.Errorf("Error deleting public SSH key %s: %w", k, err)
		}
	}

	return nil
}

func DeleteUserVirtualMFADevices(ctx context.Context, conn *iam.IAM, username string) error {
	var VirtualMFADevices []string
	var err error

	listVirtualMFADevices := &iam.ListVirtualMFADevicesInput{
		AssignmentStatus: aws.String("Assigned"),
	}
	pageOfVirtualMFADevices := func(page *iam.ListVirtualMFADevicesOutput, lastPage bool) (shouldContinue bool) {
		for _, m := range page.VirtualMFADevices {
			// UserName is `nil` for the root user
			if aws.StringValue(m.User.UserName) == username {
				VirtualMFADevices = append(VirtualMFADevices, *m.SerialNumber)
			}
		}
		return !lastPage
	}
	err = conn.ListVirtualMFADevicesPagesWithContext(ctx, listVirtualMFADevices, pageOfVirtualMFADevices)
	if err != nil {
		return fmt.Errorf("Error removing Virtual MFA devices of user %s: %w", username, err)
	}
	for _, m := range VirtualMFADevices {
		_, err := conn.DeactivateMFADeviceWithContext(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deactivating Virtual MFA device %s: %w", m, err)
		}
		_, err = conn.DeleteVirtualMFADeviceWithContext(ctx, &iam.DeleteVirtualMFADeviceInput{
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deleting Virtual MFA device %s: %w", m, err)
		}
	}

	return nil
}

func DeactivateUserMFADevices(ctx context.Context, conn *iam.IAM, username string) error {
	var MFADevices []string
	var err error

	listMFADevices := &iam.ListMFADevicesInput{
		UserName: aws.String(username),
	}
	pageOfMFADevices := func(page *iam.ListMFADevicesOutput, lastPage bool) (shouldContinue bool) {
		for _, m := range page.MFADevices {
			MFADevices = append(MFADevices, *m.SerialNumber)
		}
		return !lastPage
	}
	err = conn.ListMFADevicesPagesWithContext(ctx, listMFADevices, pageOfMFADevices)
	if err != nil {
		return fmt.Errorf("Error removing MFA devices of user %s: %w", username, err)
	}
	for _, m := range MFADevices {
		_, err := conn.DeactivateMFADeviceWithContext(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deactivating MFA device %s: %w", m, err)
		}
	}

	return nil
}

func DeleteUserLoginProfile(ctx context.Context, conn *iam.IAM, username string) error {
	var err error
	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	}
	err = resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err = conn.DeleteLoginProfileWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				return nil
			}
			// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeEntityTemporarilyUnmodifiableException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteLoginProfileWithContext(ctx, input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting Account Login Profile: %w", err)
	}

	return nil
}

func DeleteUserAccessKeys(ctx context.Context, conn *iam.IAM, username string) error {
	accessKeys, err := FindAccessKeys(ctx, conn, username)
	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("error listing access keys for IAM User (%s): %w", username, err)
	}
	var errs *multierror.Error
	for _, k := range accessKeys {
		_, err := conn.DeleteAccessKeyWithContext(ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(username),
			AccessKeyId: k.AccessKeyId,
		})
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error deleting Access Key (%s) from User (%s): %w", aws.StringValue(k.AccessKeyId), username, err))
		}
	}

	return errs.ErrorOrNil()
}

func deleteUserSigningCertificates(ctx context.Context, conn *iam.IAM, userName string) error {
	var certificateIDList []string

	listInput := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}
	err := conn.ListSigningCertificatesPagesWithContext(ctx, listInput,
		func(page *iam.ListSigningCertificatesOutput, lastPage bool) bool {
			for _, c := range page.Certificates {
				certificateIDList = append(certificateIDList, aws.StringValue(c.CertificateId))
			}
			return !lastPage
		})
	if err != nil {
		return fmt.Errorf("Error removing signing certificates of user %s: %w", userName, err)
	}

	for _, c := range certificateIDList {
		_, err := conn.DeleteSigningCertificateWithContext(ctx, &iam.DeleteSigningCertificateInput{
			CertificateId: aws.String(c),
			UserName:      aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("Error deleting signing certificate %s: %w", c, err)
		}
	}

	return nil
}

func DeleteServiceSpecificCredentials(ctx context.Context, conn *iam.IAM, username string) error {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: aws.String(username),
	}

	output, err := conn.ListServiceSpecificCredentialsWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("Error listing Service Specific Credentials of user %s: %w", username, err)
	}
	for _, m := range output.ServiceSpecificCredentials {
		_, err := conn.DeleteServiceSpecificCredentialWithContext(ctx, &iam.DeleteServiceSpecificCredentialInput{
			UserName:                    aws.String(username),
			ServiceSpecificCredentialId: m.ServiceSpecificCredentialId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Service Specific Credentials %s: %w", m, err)
		}
	}

	return nil
}
