package memorydb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetGroupCreate,
		ReadWithoutTimeout:   resourceSubnetGroupRead,
		UpdateWithoutTimeout: resourceSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceSubnetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateResourceName(subnetGroupNameMaxLength),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateResourceNamePrefix(subnetGroupNameMaxLength - resource.UniqueIDSuffixLength),
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateSubnetGroupInput{
		Description:     aws.String(d.Get("description").(string)),
		SubnetGroupName: aws.String(name),
		SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating MemoryDB Subnet Group: %s", input)
	_, err := conn.CreateSubnetGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MemoryDB Subnet Group (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceSubnetGroupRead(ctx, d, meta)
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &memorydb.UpdateSubnetGroupInput{
			Description:     aws.String(d.Get("description").(string)),
			SubnetGroupName: aws.String(d.Id()),
			SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		log.Printf("[DEBUG] Updating MemoryDB Subnet Group: %s", input)
		_, err := conn.UpdateSubnetGroupWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating MemoryDB Subnet Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating MemoryDB Subnet Group (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceSubnetGroupRead(ctx, d, meta)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	group, err := FindSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	var subnetIds []*string
	for _, subnet := range group.Subnets {
		subnetIds = append(subnetIds, subnet.Identifier)
	}

	d.Set("arn", group.ARN)
	d.Set("description", group.Description)
	d.Set("subnet_ids", flex.FlattenStringSet(subnetIds))
	d.Set("name", group.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(group.Name)))
	d.Set("vpc_id", group.VpcId)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags for MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all for MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	log.Printf("[DEBUG] Deleting MemoryDB Subnet Group: (%s)", d.Id())
	_, err := conn.DeleteSubnetGroupWithContext(ctx, &memorydb.DeleteSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSubnetGroupNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	return nil
}
