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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceACLCreate,
		ReadWithoutTimeout:   resourceACLRead,
		UpdateWithoutTimeout: resourceACLUpdate,
		DeleteWithoutTimeout: resourceACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateResourceName(aclNameMaxLength),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateResourceNamePrefix(aclNameMaxLength - resource.UniqueIDSuffixLength),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, userNameMaxLength),
				},
			},
		},
	}
}

func resourceACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateACLInput{
		ACLName: aws.String(name),
		Tags:    Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("user_names"); ok && v.(*schema.Set).Len() > 0 {
		input.UserNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating MemoryDB ACL: %s", input)
	_, err := conn.CreateACLWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MemoryDB ACL (%s): %s", name, err)
	}

	if err := waitACLActive(ctx, conn, name); err != nil {
		return diag.Errorf("error waiting for MemoryDB ACL (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resourceACLRead(ctx, d, meta)
}

func resourceACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &memorydb.UpdateACLInput{
			ACLName: aws.String(d.Id()),
		}

		o, n := d.GetChange("user_names")
		oldSet, newSet := o.(*schema.Set), n.(*schema.Set)

		if toAdd := newSet.Difference(oldSet); toAdd.Len() > 0 {
			input.UserNamesToAdd = flex.ExpandStringSet(toAdd)
		}

		// When a user is deleted, MemoryDB will implicitly remove it from any
		// ACL-s it was associated with.
		//
		// Attempting to remove a user that isn't in the ACL will fail with
		// InvalidParameterValueException. To work around this, filter out any
		// users that have been reported as no longer being in the group.

		initialState, err := FindACLByName(ctx, conn, d.Id())
		if err != nil {
			return diag.Errorf("error getting MemoryDB ACL (%s) current state: %s", d.Id(), err)
		}

		initialUserNames := map[string]struct{}{}
		for _, userName := range initialState.UserNames {
			initialUserNames[aws.StringValue(userName)] = struct{}{}
		}

		for _, v := range oldSet.Difference(newSet).List() {
			userNameToRemove := v.(string)
			_, userNameStillPresent := initialUserNames[userNameToRemove]

			if userNameStillPresent {
				input.UserNamesToRemove = append(input.UserNamesToRemove, aws.String(userNameToRemove))
			}
		}

		if len(input.UserNamesToAdd) > 0 || len(input.UserNamesToRemove) > 0 {
			log.Printf("[DEBUG] Updating MemoryDB ACL (%s)", d.Id())

			_, err := conn.UpdateACLWithContext(ctx, input)
			if err != nil {
				return diag.Errorf("error updating MemoryDB ACL (%s): %s", d.Id(), err)
			}

			if err := waitACLActive(ctx, conn, d.Id()); err != nil {
				return diag.Errorf("error waiting for MemoryDB ACL (%s) to be modified: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating MemoryDB ACL (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceACLRead(ctx, d, meta)
}

func resourceACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	acl, err := FindACLByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MemoryDB ACL (%s): %s", d.Id(), err)
	}

	d.Set("arn", acl.ARN)
	d.Set("minimum_engine_version", acl.MinimumEngineVersion)
	d.Set("name", acl.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(acl.Name)))
	d.Set("user_names", flex.FlattenStringSet(acl.UserNames))

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for MemoryDB ACL (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags for MemoryDB ACL (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all for MemoryDB ACL (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn()

	log.Printf("[DEBUG] Deleting MemoryDB ACL: (%s)", d.Id())
	_, err := conn.DeleteACLWithContext(ctx, &memorydb.DeleteACLInput{
		ACLName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeACLNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting MemoryDB ACL (%s): %s", d.Id(), err)
	}

	if err := waitACLDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for MemoryDB ACL (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
