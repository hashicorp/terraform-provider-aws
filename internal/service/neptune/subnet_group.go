package neptune

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	subnetIdsSet := d.Get("subnet_ids").(*schema.Set)
	subnetIds := make([]*string, subnetIdsSet.Len())
	for i, subnetId := range subnetIdsSet.List() {
		subnetIds[i] = aws.String(subnetId.(string))
	}

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := neptune.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(groupName),
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		SubnetIds:                subnetIds,
		Tags:                     Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Create Neptune Subnet Group: %#v", createOpts)
	_, err := conn.CreateDBSubnetGroupWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Subnet Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.DBSubnetGroupName))
	log.Printf("[INFO] Neptune Subnet Group ID: %s", d.Id())
	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := neptune.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	var subnetGroups []*neptune.DBSubnetGroup
	if err := conn.DescribeDBSubnetGroupsPagesWithContext(ctx, &describeOpts, func(resp *neptune.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
		return !lastPage
	}); err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBSubnetGroupNotFoundFault) {
			log.Printf("[WARN] Neptune Subnet Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Neptune Subnet Group (%s): %s", d.Id(), err)
	}

	if len(subnetGroups) == 0 {
		log.Printf("[WARN] Unable to find Neptune Subnet Group: %#v, removing from state", subnetGroups)
		d.SetId("")
		return diags
	}

	subnetGroup := subnetGroups[0]

	if subnetGroup.DBSubnetGroupName == nil {
		return sdkdiag.AppendErrorf(diags, "Unable to find Neptune Subnet Group: %#v", subnetGroups)
	}

	d.Set("name", subnetGroup.DBSubnetGroupName)
	d.Set("description", subnetGroup.DBSubnetGroupDescription)

	subnets := make([]string, 0, len(subnetGroup.Subnets))
	for _, s := range subnetGroup.Subnets {
		subnets = append(subnets, aws.StringValue(s.SubnetIdentifier))
	}
	if err := d.Set("subnet_ids", subnets); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	//Amazon Neptune shares the format of Amazon RDS ARNs. Neptune ARNs contain rds and not neptune.
	//https://docs.aws.amazon.com/neptune/latest/userguide/tagging.ARN.html
	d.Set("arn", subnetGroup.DBSubnetGroupArn)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Neptune Subnet Group (%s): %s", d.Get("arn").(string), err)
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
	conn := meta.(*conns.AWSClient).NeptuneConn()
	if d.HasChanges("subnet_ids", "description") {
		_, n := d.GetChange("subnet_ids")
		if n == nil {
			n = new(schema.Set)
		}
		ns := n.(*schema.Set)

		var sIds []*string
		for _, s := range ns.List() {
			sIds = append(sIds, aws.String(s.(string)))
		}

		_, err := conn.ModifyDBSubnetGroupWithContext(ctx, &neptune.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Subnet Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Subnet Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	input := neptune.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Neptune Subnet Group: %s", d.Id())
	_, err := conn.DeleteDBSubnetGroupWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBSubnetGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}
