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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	instanceProfileNameMaxLen       = 128
	instanceProfileNamePrefixMaxLen = instanceProfileNameMaxLen - resource.UniqueIDSuffixLength
)

func ResourceInstanceProfile() *schema.Resource {
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
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	request := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(name),
		Path:                aws.String(d.Get("path").(string)),
	}

	if len(tags) > 0 {
		request.Tags = Tags(tags.IgnoreAWS())
	}

	var err error
	response, err := conn.CreateInstanceProfileWithContext(ctx, request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM Instance Profile (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		response, err = conn.CreateInstanceProfileWithContext(ctx, request)
	}

	if err == nil {
		err = instanceProfileReadResult(d, response.InstanceProfile, meta) // sets id
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Instance Profile (%s): %s", name, err)
	}

	waiterRequest := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}
	// don't return until the IAM service reports that the instance profile is ready.
	// this ensures that terraform resources which rely on the instance profile will 'see'
	// that the instance profile exists.
	err = conn.WaitUntilInstanceProfileExistsWithContext(ctx, waiterRequest)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "timed out while waiting for instance profile %s: %s", name, err)
	}

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 {
		err := instanceProfileUpdateTags(ctx, conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM Instance Profile (%s): %s", d.Id(), err)
			return append(diags, resourceInstanceProfileUpdate(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM Instance Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceProfileUpdate(ctx, d, meta)...)
}

func instanceProfileAddRole(ctx context.Context, conn *iam.IAM, profileName, roleName string) error {
	request := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := conn.AddRoleToInstanceProfileWithContext(ctx, request)
		// IAM unfortunately does not provide a better error code or message for eventual consistency
		// InvalidParameterValue: Value (XXX) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		// NoSuchEntity: The request was rejected because it referenced an entity that does not exist. The error message describes the entity. HTTP Status Code: 404
		if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Invalid IAM Instance Profile name") || tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "The role with name") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AddRoleToInstanceProfileWithContext(ctx, request)
	}
	if err != nil {
		return fmt.Errorf("adding IAM Role %s to Instance Profile %s: %w", roleName, profileName, err)
	}

	return err
}

func instanceProfileRemoveRole(ctx context.Context, conn *iam.IAM, profileName, roleName string) error {
	request := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := conn.RemoveRoleFromInstanceProfileWithContext(ctx, request)
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	return err
}

func instanceProfileRemoveAllRoles(ctx context.Context, d *schema.ResourceData, conn *iam.IAM) error {
	if role, ok := d.GetOk("role"); ok {
		err := instanceProfileRemoveRole(ctx, conn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("removing role (%s): %w", role, err)
		}
	}

	return nil
}

func resourceInstanceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	if d.HasChange("role") {
		oldRole, newRole := d.GetChange("role")

		if oldRole.(string) != "" {
			err := instanceProfileRemoveRole(ctx, conn, d.Id(), oldRole.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "removing role %s to IAM Instance Profile (%s): %s", oldRole.(string), d.Id(), err)
			}
		}

		if newRole.(string) != "" {
			err := instanceProfileAddRole(ctx, conn, d.Id(), newRole.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding role %s to IAM Instance Profile (%s): %s", newRole.(string), d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := instanceProfileUpdateTags(ctx, conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM Instance Profile (%s): %s", d.Id(), err)
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for IAM Instance Profile (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceInstanceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	request := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}

	result, err := conn.GetInstanceProfileWithContext(ctx, request)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Instance Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profile (%s): %s", d.Id(), err)
	}

	instanceProfile := result.InstanceProfile
	if instanceProfile.Roles != nil && len(instanceProfile.Roles) > 0 {
		roleName := aws.StringValue(instanceProfile.Roles[0].RoleName)
		input := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err := conn.GetRoleWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				err := instanceProfileRemoveRole(ctx, conn, d.Id(), roleName)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "removing role %s to IAM Instance Profile (%s): %s", roleName, d.Id(), err)
				}
			}
			return sdkdiag.AppendErrorf(diags, "reading IAM Role %s attached to IAM Instance Profile (%s): %s", roleName, d.Id(), err)
		}
	}

	if err := instanceProfileReadResult(d, instanceProfile, meta); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profile (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceInstanceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	if err := instanceProfileRemoveAllRoles(ctx, d, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Instance Profile (%s): %s", d.Id(), err)
	}

	request := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}
	_, err := conn.DeleteInstanceProfileWithContext(ctx, request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM Instance Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func instanceProfileReadResult(d *schema.ResourceData, result *iam.InstanceProfile, meta interface{}) error {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	d.SetId(aws.StringValue(result.InstanceProfileName))
	d.Set("name", result.InstanceProfileName)
	d.Set("arn", result.Arn)
	d.Set("create_date", result.CreateDate.Format(time.RFC3339))
	d.Set("path", result.Path)
	d.Set("unique_id", result.InstanceProfileId)

	if result.Roles != nil && len(result.Roles) > 0 {
		d.Set("role", result.Roles[0].RoleName) //there will only be 1 role returned
	}

	tags := KeyValueTags(result.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}
