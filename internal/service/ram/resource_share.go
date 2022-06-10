package ram

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourceShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceShareCreate,
		Read:   resourceResourceShareRead,
		Update: resourceResourceShareUpdate,
		Delete: resourceResourceShareDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allow_external_principals": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"permission_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceResourceShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ram.CreateResourceShareInput{
		AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
		Name:                    aws.String(name),
	}

	if v, ok := d.GetOk("permission_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.PermissionArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating RAM Resource Share: %s", input)
	output, err := conn.CreateResourceShare(input)

	if err != nil {
		return fmt.Errorf("creating RAM Resource Share (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ResourceShare.ResourceShareArn))

	if _, err = WaitResourceShareOwnedBySelfActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) to become ready: %w", d.Id(), err)
	}

	return resourceResourceShareRead(d, meta)
}

func resourceResourceShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resourceShare, err := FindResourceShareOwnerSelfByARN(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		log.Printf("[WARN] RAM Resource Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading RAM Resource Share (%s): %w", d.Id(), err)
	}

	if !d.IsNewResource() && aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusActive {
		log.Printf("[WARN] RAM Resource Share (%s) not active, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("allow_external_principals", resourceShare.AllowExternalPrincipals)
	d.Set("arn", resourceShare.ResourceShareArn)
	d.Set("name", resourceShare.Name)

	tags := KeyValueTags(resourceShare.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	perms, err := conn.ListResourceSharePermissions(&ram.ListResourceSharePermissionsInput{
		ResourceShareArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("listing RAM Resource Share (%s) permissions: %w", d.Id(), err)
	}

	permissionARNs := make([]*string, 0, len(perms.Permissions))

	for _, v := range perms.Permissions {
		permissionARNs = append(permissionARNs, v.Arn)
	}

	d.Set("permission_arns", aws.StringValueSlice(permissionARNs))

	return nil
}

func resourceResourceShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	if d.HasChanges("name", "allow_external_principals") {
		input := &ram.UpdateResourceShareInput{
			AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
			Name:                    aws.String(d.Get("name").(string)),
			ResourceShareArn:        aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating RAM Resource Share: %s", input)
		_, err := conn.UpdateResourceShare(input)

		if err != nil {
			return fmt.Errorf("updating RAM Resource Share (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating RAM Resource Share (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceResourceShareRead(d, meta)
}

func resourceResourceShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	log.Printf("[DEBUG] Delete RAM Resource Share: %s", d.Id())
	_, err := conn.DeleteResourceShare(&ram.DeleteResourceShareInput{
		ResourceShareArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting RAM Resource Share (%s): %w", d.Id(), err)
	}

	if _, err = WaitResourceShareOwnedBySelfDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) delete: %w", d.Id(), err)
	}

	return nil
}
