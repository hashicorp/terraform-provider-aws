package docdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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

		Schema: map[string]*schema.Schema{
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
				ValidateFunc:  validSubnetGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validSubnetGroupNamePrefix,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	subnetIds := flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set))

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := docdb.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(groupName),
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		SubnetIds:                subnetIds,
		Tags:                     Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Create DocDB Subnet Group: %#v", createOpts)
	_, err := conn.CreateDBSubnetGroupWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocDB Subnet Group: %s", err)
	}

	d.SetId(groupName)

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	var subnetGroups []*docdb.DBSubnetGroup
	if err := conn.DescribeDBSubnetGroupsPagesWithContext(ctx, &describeOpts, func(resp *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
		return !lastPage
	}); err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			log.Printf("[WARN] DocDB Subnet Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DocDB Subnet Group (%s) parameters: %s", d.Id(), err)
	}

	if !d.IsNewResource() && (len(subnetGroups) != 1 || aws.StringValue(subnetGroups[0].DBSubnetGroupName) != d.Id()) {
		log.Printf("[WARN] DocDB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	subnetGroup := subnetGroups[0]
	d.Set("name", subnetGroup.DBSubnetGroupName)
	d.Set("description", subnetGroup.DBSubnetGroupDescription)
	d.Set("arn", subnetGroup.DBSubnetGroupArn)

	subnets := make([]string, 0, len(subnetGroup.Subnets))
	for _, s := range subnetGroup.Subnets {
		subnets = append(subnets, aws.StringValue(s.SubnetIdentifier))
	}
	if err := d.Set("subnet_ids", subnets); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DocumentDB Subnet Group (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn()

	if d.HasChanges("subnet_ids", "description") {
		_, n := d.GetChange("subnet_ids")
		if n == nil {
			n = new(schema.Set)
		}
		sIds := flex.ExpandStringSet(n.(*schema.Set))

		_, err := conn.ModifyDBSubnetGroupWithContext(ctx, &docdb.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modify DocDB Subnet Group (%s) parameters: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DocumentDB Subnet Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn()

	delOpts := docdb.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DocDB Subnet Group: %s", d.Id())

	_, err := conn.DeleteDBSubnetGroupWithContext(ctx, &delOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DocDB Subnet Group (%s): %s", d.Id(), err)
	}

	if err := WaitForSubnetGroupDeletion(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocDB Subnet Group (%s): %s", d.Id(), err)
	}
	return diags
}

func WaitForSubnetGroupDeletion(ctx context.Context, conn *docdb.DocDB, name string) error {
	params := &docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(name),
	}

	err := resource.RetryContext(ctx, 10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeDBSubnetGroupsWithContext(ctx, params)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("DocDB Subnet Group (%s) still exists", name))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeDBSubnetGroupsWithContext(ctx, params)
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting DocDB subnet group: %s", err)
	}
	return nil
}
