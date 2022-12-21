package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMScope() *schema.Resource {
	return &schema.Resource{
		Create: ResourceIPAMScopeCreate,
		Read:   ResourceIPAMScopeRead,
		Update: ResourceIPAMScopeUpdate,
		Delete: ResourceIPAMScopeDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pool_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func ResourceIPAMScopeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateIpamScopeInput{
		ClientToken:       aws.String(resource.UniqueId()),
		IpamId:            aws.String(d.Get("ipam_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpamScope),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamScope(input)

	if err != nil {
		return fmt.Errorf("creating IPAM Scope: %w", err)
	}

	d.SetId(aws.StringValue(output.IpamScope.IpamScopeId))

	if _, err := WaitIPAMScopeCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for IPAM Scope (%s) create: %w", d.Id(), err)
	}

	return ResourceIPAMScopeRead(d, meta)
}

func ResourceIPAMScopeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	scope, err := FindIPAMScopeByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IPAM Scope (%s): %w", d.Id(), err)
	}

	ipamID := strings.Split(aws.StringValue(scope.IpamArn), "/")[1]
	d.Set("arn", scope.IpamScopeArn)
	d.Set("description", scope.Description)
	d.Set("ipam_arn", scope.IpamArn)
	d.Set("ipam_id", ipamID)
	d.Set("ipam_scope_type", scope.IpamScopeType)
	d.Set("is_default", scope.IsDefault)
	d.Set("pool_count", scope.PoolCount)

	tags := KeyValueTags(scope.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func ResourceIPAMScopeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("description") {
		input := &ec2.ModifyIpamScopeInput{
			IpamScopeId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.ModifyIpamScope(input)

		if err != nil {
			return fmt.Errorf("updating IPAM Scope (%s): %w", d.Id(), err)
		}

		if _, err := WaitIPAMScopeUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for IPAM Scope (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating IPAM Scope (%s) tags: %w", d.Id(), err)
		}
	}

	return ResourceIPAMScopeRead(d, meta)
}

func ResourceIPAMScopeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting IPAM Scope: %s", d.Id())
	_, err := conn.DeleteIpamScope(&ec2.DeleteIpamScopeInput{
		IpamScopeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IPAM Scope: (%s): %w", d.Id(), err)
	}

	if _, err := WaitIPAMScopeDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for IPAM Scope (%s) delete: %w", d.Id(), err)
	}

	return nil
}
