// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_subnet_group", name="Subnet Group")
// @Tags(identifierAttribute="arn")
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
				ValidateFunc:  validateResourceNamePrefix(subnetGroupNameMaxLength - id.UniqueIDSuffixLength),
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateSubnetGroupInput{
		Description:     aws.String(d.Get("description").(string)),
		SubnetGroupName: aws.String(name),
		SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:            getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating MemoryDB Subnet Group: %s", input)
	_, err := conn.CreateSubnetGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MemoryDB Subnet Group (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceSubnetGroupRead(ctx, d, meta)
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &memorydb.UpdateSubnetGroupInput{
			Description:     aws.String(d.Get("description").(string)),
			SubnetGroupName: aws.String(d.Id()),
			SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		log.Printf("[DEBUG] Updating MemoryDB Subnet Group: %s", input)
		_, err := conn.UpdateSubnetGroupWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MemoryDB Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return resourceSubnetGroupRead(ctx, d, meta)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	group, err := FindSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MemoryDB Subnet Group (%s): %s", d.Id(), err)
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

	return nil
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Subnet Group: (%s)", d.Id())
	_, err := conn.DeleteSubnetGroupWithContext(ctx, &memorydb.DeleteSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSubnetGroupNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting MemoryDB Subnet Group (%s): %s", d.Id(), err)
	}

	return nil
}
