package iam

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
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
	createResp, err := conn.CreateUser(request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed creating IAM User (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		createResp, err = conn.CreateUser(request)
	}

	if err != nil {
		return fmt.Errorf("failed creating IAM User (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(createResp.User.UserName))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID {
		err := userUpdateTags(conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed adding tags after create for IAM User (%s): %s", d.Id(), err)
			return resourceUserRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for IAM User (%s): %w", d.Id(), err)
		}
	}

	return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	request := &iam.GetUserInput{
		UserName: aws.String(d.Id()),
	}

	var output *iam.GetUserOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.GetUser(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetUser(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM User (%s): %w", d.Id(), err)
	}

	if output == nil || output.User == nil {
		return fmt.Errorf("error reading IAM User (%s): empty response", d.Id())
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
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	if d.HasChanges("name", "path") {
		on, nn := d.GetChange("name")
		_, np := d.GetChange("path")

		request := &iam.UpdateUserInput{
			UserName:    aws.String(on.(string)),
			NewUserName: aws.String(nn.(string)),
			NewPath:     aws.String(np.(string)),
		}

		log.Println("[DEBUG] Update IAM User request:", request)
		_, err := conn.UpdateUser(request)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				log.Printf("[WARN] No IAM user by name (%s) found", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating IAM User %s: %w", d.Id(), err)
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
			_, err := conn.PutUserPermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error updating IAM User permissions boundary: %w", err)
			}
		} else {
			input := &iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String(d.Id()),
			}
			_, err := conn.DeleteUserPermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error deleting IAM User permissions boundary: %w", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := userUpdateTags(conn, d.Id(), o, n)

		// Some partitions may not support tagging, giving error
		if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed updating tags for IAM User (%s): %s", d.Id(), err)
			return resourceUserRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed updating tags for IAM User (%s): %w", d.Id(), err)
		}
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	// IAM Users must be removed from all groups before they can be deleted
	if err := DeleteUserGroupMemberships(conn, d.Id()); err != nil {
		return fmt.Errorf("error removing IAM User (%s) group memberships: %s", d.Id(), err)
	}

	// All access keys, MFA devices and login profile for the user must be removed
	if d.Get("force_destroy").(bool) {
		if err := DeleteUserAccessKeys(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) access keys: %w", d.Id(), err)
		}

		if err := DeleteUserSSHKeys(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) SSH keys: %w", d.Id(), err)
		}

		if err := DeleteUserVirtualMFADevices(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) Virtual MFA devices: %w", d.Id(), err)
		}

		if err := DeactivateUserMFADevices(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) MFA devices: %w", d.Id(), err)
		}

		if err := DeleteUserLoginProfile(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) login profile: %w", d.Id(), err)
		}

		if err := deleteUserSigningCertificates(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) signing certificate: %w", d.Id(), err)
		}

		if err := DeleteServiceSpecificCredentials(conn, d.Id()); err != nil {
			return fmt.Errorf("error removing IAM User (%s) Service Specific Credentials: %w", d.Id(), err)
		}
	}

	deleteUserInput := &iam.DeleteUserInput{
		UserName: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete IAM User request:", deleteUserInput)
	_, err := conn.DeleteUser(deleteUserInput)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting IAM User %s: %w", d.Id(), err)
	}

	return nil
}

func DeleteUserGroupMemberships(conn *iam.IAM, username string) error {
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
	err := conn.ListGroupsForUserPages(listGroups, pageOfGroups)
	if err != nil {
		return fmt.Errorf("Error removing user %q from all groups: %s", username, err)
	}
	for _, g := range groups {
		// use iam group membership func to remove user from all groups
		log.Printf("[DEBUG] Removing IAM User %s from IAM Group %s", username, g)
		if err := removeUsersFromGroup(conn, []*string{aws.String(username)}, g); err != nil {
			return err
		}
	}

	return nil
}

func DeleteUserSSHKeys(svc *iam.IAM, username string) error {
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
	err = svc.ListSSHPublicKeysPages(listSSHPublicKeys, pageOfListSSHPublicKeys)
	if err != nil {
		return fmt.Errorf("Error removing public SSH keys of user %s: %w", username, err)
	}
	for _, k := range publicKeys {
		_, err := svc.DeleteSSHPublicKey(&iam.DeleteSSHPublicKeyInput{
			UserName:       aws.String(username),
			SSHPublicKeyId: aws.String(k),
		})
		if err != nil {
			return fmt.Errorf("Error deleting public SSH key %s: %w", k, err)
		}
	}

	return nil
}

func DeleteUserVirtualMFADevices(svc *iam.IAM, username string) error {
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
	err = svc.ListVirtualMFADevicesPages(listVirtualMFADevices, pageOfVirtualMFADevices)
	if err != nil {
		return fmt.Errorf("Error removing Virtual MFA devices of user %s: %w", username, err)
	}
	for _, m := range VirtualMFADevices {
		_, err := svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deactivating Virtual MFA device %s: %w", m, err)
		}
		_, err = svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deleting Virtual MFA device %s: %w", m, err)
		}
	}

	return nil
}

func DeactivateUserMFADevices(svc *iam.IAM, username string) error {
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
	err = svc.ListMFADevicesPages(listMFADevices, pageOfMFADevices)
	if err != nil {
		return fmt.Errorf("Error removing MFA devices of user %s: %w", username, err)
	}
	for _, m := range MFADevices {
		_, err := svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
			UserName:     aws.String(username),
			SerialNumber: aws.String(m),
		})
		if err != nil {
			return fmt.Errorf("Error deactivating MFA device %s: %w", m, err)
		}
	}

	return nil
}

func DeleteUserLoginProfile(svc *iam.IAM, username string) error {
	var err error
	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	}
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err = svc.DeleteLoginProfile(input)
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
		_, err = svc.DeleteLoginProfile(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting Account Login Profile: %w", err)
	}

	return nil
}

func DeleteUserAccessKeys(svc *iam.IAM, username string) error {
	var accessKeys []string
	var err error
	listAccessKeys := &iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}
	pageOfAccessKeys := func(page *iam.ListAccessKeysOutput, lastPage bool) (shouldContinue bool) {
		for _, k := range page.AccessKeyMetadata {
			accessKeys = append(accessKeys, *k.AccessKeyId)
		}
		return !lastPage
	}
	err = svc.ListAccessKeysPages(listAccessKeys, pageOfAccessKeys)
	if err != nil {
		return fmt.Errorf("Error removing access keys of user %s: %w", username, err)
	}
	for _, k := range accessKeys {
		_, err := svc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			UserName:    aws.String(username),
			AccessKeyId: aws.String(k),
		})
		if err != nil {
			return fmt.Errorf("Error deleting access key %s: %w", k, err)
		}
	}

	return nil
}

func deleteUserSigningCertificates(svc *iam.IAM, userName string) error {
	var certificateIDList []string

	listInput := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}
	err := svc.ListSigningCertificatesPages(listInput,
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
		_, err := svc.DeleteSigningCertificate(&iam.DeleteSigningCertificateInput{
			CertificateId: aws.String(c),
			UserName:      aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("Error deleting signing certificate %s: %w", c, err)
		}
	}

	return nil
}

func DeleteServiceSpecificCredentials(svc *iam.IAM, username string) error {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: aws.String(username),
	}

	output, err := svc.ListServiceSpecificCredentials(input)
	if err != nil {
		return fmt.Errorf("Error listing Service Specific Credentials of user %s: %w", username, err)
	}
	for _, m := range output.ServiceSpecificCredentials {
		_, err := svc.DeleteServiceSpecificCredential(&iam.DeleteServiceSpecificCredentialInput{
			UserName:                    aws.String(username),
			ServiceSpecificCredentialId: m.ServiceSpecificCredentialId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Service Specific Credentials %s: %w", m, err)
		}
	}

	return nil
}
