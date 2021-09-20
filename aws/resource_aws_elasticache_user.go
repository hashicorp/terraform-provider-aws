package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"REDIS"}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"no_password_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"passwords": {
				Type:      schema.TypeSet,
				Optional:  true,
				MaxItems:  2,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Sensitive: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get("engine").(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		UserId:             aws.String(d.Get("user_id").(string)),
		UserName:           aws.String(d.Get("user_name").(string)),
	}

	if v, ok := d.GetOk("passwords"); ok {
		input.Passwords = flex.ExpandStringSet(v.(*schema.Set))
	}

	// Tags are currently only supported in AWS Commercial.
	if len(tags) > 0 && meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		input.Tags = tags.IgnoreAws().ElasticacheTags()
	}

	out, err := conn.CreateUser(input)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache User: %w", err)
	}

	d.SetId(aws.StringValue(out.UserId))

	return resourceUserRead(d, meta)

}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := finder.ElastiCacheUserById(conn, d.Id())
	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, elasticache.ErrCodeUserNotFoundFault, "")) {
		log.Printf("[WARN] ElastiCache User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing ElastiCache User (%s): %w", d.Id(), err)
	}

	d.Set("access_string", resp.AccessString)
	d.Set("engine", resp.Engine)
	d.Set("user_id", resp.UserId)
	d.Set("user_name", resp.UserName)
	d.Set("arn", resp.ARN)

	// Tags are currently only supported in AWS Commercial.
	if meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		tags, err := tftags.ElasticacheListTags(conn, aws.StringValue(resp.ARN))

		if err != nil {
			return fmt.Errorf("error listing tags for ElastiCache User (%s): %w", aws.StringValue(resp.ARN), err)
		}

		tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	} else {
		d.Set("tags", nil)
		d.Set("tags_all", nil)
	}

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	hasChange := false

	if d.HasChangesExcept("tags", "tags_all") {
		req := &elasticache.ModifyUserInput{
			UserId: aws.String(d.Id()),
		}

		if d.HasChange("access_string") {
			req.AccessString = aws.String(d.Get("access_string").(string))
			hasChange = true
		}

		if d.HasChange("no_password_required") {
			req.NoPasswordRequired = aws.Bool(d.Get("no_password_required").(bool))
			hasChange = true
		}

		if d.HasChange("passwords") {
			req.Passwords = flex.ExpandStringSet(d.Get("passwords").(*schema.Set))
			hasChange = true
		}

		if hasChange {
			_, err := conn.ModifyUser(req)
			if err != nil {
				return fmt.Errorf("error updating ElastiCache User (%s): %w", d.Id(), err)
			}

			if err := waiter.UserActive(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for ElastiCache User (%s) to be modified: %w", d.Id(), err)
			}
		}

	}
	// Tags are currently only supported in AWS Commercial.
	if d.HasChange("tags_all") && meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		o, n := d.GetChange("tags_all")

		if err := tftags.ElasticacheUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating ElastiCache User (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.DeleteUserInput{
		UserId: aws.String(d.Id()),
	}

	_, err := conn.DeleteUser(input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ElastiCache User (%s): %w", d.Id(), err)
	}

	if err := waiter.UserDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error waiting for ElastiCache User (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
