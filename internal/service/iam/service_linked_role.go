package iam

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceServiceLinkedRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceLinkedRoleCreate,
		Read:   resourceServiceLinkedRoleRead,
		Update: resourceServiceLinkedRoleUpdate,
		Delete: resourceServiceLinkedRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceLinkedRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

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

	log.Printf("[DEBUG] Creating IAM Service Linked Role: %s", input)
	output, err := conn.CreateServiceLinkedRole(input)

	if err != nil {
		return fmt.Errorf("error creating IAM Service Linked Role (%s): %w", serviceName, err)
	}

	d.SetId(aws.StringValue(output.Role.Arn))

	if len(tags) > 0 {
		_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

		if err != nil {
			return err
		}

		err = roleUpdateTags(conn, roleName, nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed adding tags after create for IAM Service Linked Role (%s): %s", d.Id(), err)
			return resourceServiceLinkedRoleRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for IAM Service Linked Role (%s): %w", d.Id(), err)
		}
	}

	return resourceServiceLinkedRoleRead(d, meta)
}

func resourceServiceLinkedRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	serviceName, roleName, customSuffix, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindRoleByName(conn, roleName)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Service Linked Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Service Linked Role (%s): %w", d.Id(), err)
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

	tags := KeyValueTags(role.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceServiceLinkedRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return err
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &iam.UpdateRoleInput{
			Description: aws.String(d.Get("description").(string)),
			RoleName:    aws.String(roleName),
		}

		log.Printf("[DEBUG] Updating IAM Service Linked Role: %s", input)
		_, err = conn.UpdateRole(input)

		if err != nil {
			return fmt.Errorf("error updating IAM Service Linked Role (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := roleUpdateTags(conn, roleName, o, n)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed updating tags for IAM Service Linked Role (%s): %s", d.Id(), err)
			return resourceServiceLinkedRoleRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed updating tags for IAM Service Linked Role (%s): %w", d.Id(), err)
		}
	}

	return resourceServiceLinkedRoleRead(d, meta)
}

func resourceServiceLinkedRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	_, roleName, _, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting IAM Service Linked Role: (%s)", d.Id())
	output, err := conn.DeleteServiceLinkedRole(&iam.DeleteServiceLinkedRoleInput{
		RoleName: aws.String(roleName),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IAM Service Linked Role (%s): %w", d.Id(), err)
	}

	deletionTaskID := aws.StringValue(output.DeletionTaskId)

	if deletionTaskID == "" {
		return nil
	}

	err = waitDeleteServiceLinkedRole(conn, deletionTaskID)

	if err != nil {
		return fmt.Errorf("error waiting for IAM Service Linked Role (%s) delete: %w", d.Id(), err)
	}

	return nil
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
