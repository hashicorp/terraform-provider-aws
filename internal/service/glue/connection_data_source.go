package glue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectionRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_properties": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlueConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("id").(string)
	catalogID, connectionName, err := DecodeConnectionID(id)
	if err != nil {
		return diag.Errorf("error decoding Glue Connection %s: %s", id, err)
	}

	connection, err := FindConnectionByName(conn, connectionName, catalogID)
	if err != nil {
		if tfresource.NotFound(err) {
			return diag.Errorf("error Glue Connection (%s) not found", id)
		}
		return diag.Errorf("error reading Glue Connection (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("catalog_id", catalogID)
	d.Set("connection_type", connection.ConnectionType)
	d.Set("name", connection.Name)
	d.Set("description", connection.Description)

	connectionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("connection/%s", connectionName),
	}.String()
	d.Set("arn", connectionArn)

	if err := d.Set("connection_properties", aws.StringValueMap(connection.ConnectionProperties)); err != nil {
		return diag.Errorf("error setting connection_properties: %s", err)
	}

	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return diag.Errorf("error setting physical_connection_requirements: %s", err)
	}

	if err := d.Set("match_criteria", flex.FlattenStringList(connection.MatchCriteria)); err != nil {
		return diag.Errorf("error setting match_criteria: %s", err)
	}

	tags, err := ListTags(conn, connectionArn)

	if err != nil {
		return diag.Errorf("error listing tags for Glue Connection (%s): %s", connectionArn, err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
