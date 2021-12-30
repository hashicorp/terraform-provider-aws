package memorydb

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"acl_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 40),
					validation.StringDoesNotMatch(
						regexp.MustCompile(`[-][-]`),
						"The name may not contain two consecutive hyphens."),
					validation.StringMatch(
						// Similar to ElastiCache, MemoryDB normalises names to lowercase.
						regexp.MustCompile(`^[a-z0-9-]*[a-z0-9]$`),
						"Only lowercase alphanumeric characters and hyphens allowed. The name may not end with a hyphen."),
				),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 40-resource.UniqueIDSuffixLength),
					validation.StringDoesNotMatch(
						regexp.MustCompile(`[-][-]`),
						"The name may not contain two consecutive hyphens."),
					validation.StringMatch(
						// Similar to ElastiCache, MemoryDB normalises names to lowercase.
						regexp.MustCompile(`^[a-z0-9-]+$`),
						"Only lowercase alphanumeric characters and hyphens allowed."),
				),
			},
			"node_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateClusterInput{
		ACLName:     aws.String(d.Get("acl_name").(string)),
		ClusterName: aws.String(name),
		NodeType:    aws.String(d.Get("node_type").(string)),
		Tags:        Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating MemoryDB Cluster: %s", input)
	_, err := conn.CreateClusterWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MemoryDB Cluster (%s): %s", name, err)
	}

	if err := waitClusterAvailable(ctx, conn, name); err != nil {
		return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &memorydb.UpdateClusterInput{
			ClusterName: aws.String(d.Id()),
		}

		if d.HasChange("acl_name") {
			input.ACLName = aws.String(d.Get("acl_name").(string))
		}

		if d.HasChange("node_type") {
			input.NodeType = aws.String(d.Get("node_type").(string))
		}

		log.Printf("[DEBUG] Updating MemoryDB Cluster (%s)", d.Id())

		_, err := conn.UpdateClusterWithContext(ctx, input)
		if err != nil {
			return diag.Errorf("error updating MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		if err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be modified: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating MemoryDB Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("acl_name", cluster.ACLName)
	d.Set("arn", cluster.ARN)
	d.Set("name", cluster.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(cluster.Name)))
	d.Set("node_type", cluster.NodeType)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn

	log.Printf("[DEBUG] Deleting MemoryDB Cluster: (%s)", d.Id())
	_, err := conn.DeleteClusterWithContext(ctx, &memorydb.DeleteClusterInput{
		ClusterName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := waitClusterDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for MemoryDB Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
