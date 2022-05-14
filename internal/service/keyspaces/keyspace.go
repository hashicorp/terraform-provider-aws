package keyspaces

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceKeyspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyspaceCreate,
		ReadWithoutTimeout:   resourceKeyspaceRead,
		UpdateWithoutTimeout: resourceKeyspaceUpdate,
		DeleteWithoutTimeout: resourceKeyspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 48),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,47}$`),
						"The name must consist of alphanumerics and underscores.",
					),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &keyspaces.CreateKeyspaceInput{
		KeyspaceName: aws.String(name),
	}

	if tags := Tags(tags.IgnoreAWS()); len(tags) > 0 {
		// The Keyspaces API requires that when Tags is set, it's non-empty.
		input.Tags = tags
	}

	log.Printf("[DEBUG] Creating Keyspaces Keyspace: %s", input)
	_, err := conn.CreateKeyspaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Keyspaces Keyspace (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFoundContext(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindKeyspaceByName(ctx, conn, d.Id())
	})

	if err != nil {
		return diag.Errorf("waiting for Keyspaces Keyspace (%s) create: %s", d.Id(), err)
	}

	return resourceKeyspaceRead(ctx, d, meta)
}

func resourceKeyspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	keyspace, err := FindKeyspaceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Keyspaces Keyspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Keyspaces Keyspace (%s): %s", d.Id(), err)
	}

	d.Set("arn", keyspace.ResourceArn)
	d.Set("name", keyspace.KeyspaceName)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for Keyspaces Keyspace (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceKeyspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Keyspaces Keyspace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceKeyspaceRead(ctx, d, meta)
}

func resourceKeyspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn

	log.Printf("[DEBUG] Deleting Keyspaces Keyspace: (%s)", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteKeyspaceWithContext(ctx, &keyspaces.DeleteKeyspaceInput{
				KeyspaceName: aws.String(d.Id()),
			})
		},
		keyspaces.ErrCodeConflictException, "a table under it is currently being created or deleted")

	if tfawserr.ErrCodeEquals(err, keyspaces.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Keyspaces Keyspace (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFoundContext(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return FindKeyspaceByName(ctx, conn, d.Id())
	})

	if err != nil {
		return diag.Errorf("waiting for Keyspaces Keyspace (%s) delete: %s", d.Id(), err)
	}

	return nil
}
