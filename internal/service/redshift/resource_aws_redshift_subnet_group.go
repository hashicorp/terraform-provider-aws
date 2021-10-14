package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetGroupCreate,
		Read:   resourceSubnetGroupRead,
		Update: resourceSubnetGroupUpdate,
		Delete: resourceSubnetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringNotInSlice([]string{"default"}, false),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	subnetIdsSet := d.Get("subnet_ids").(*schema.Set)
	subnetIds := make([]*string, subnetIdsSet.Len())
	for i, subnetId := range subnetIdsSet.List() {
		subnetIds[i] = aws.String(subnetId.(string))
	}

	createOpts := redshift.CreateClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(d.Get("name").(string)),
		Description:            aws.String(d.Get("description").(string)),
		SubnetIds:              subnetIds,
		Tags:                   tags.IgnoreAws().RedshiftTags(),
	}

	log.Printf("[DEBUG] Create Redshift Subnet Group: %#v", createOpts)
	_, err := conn.CreateClusterSubnetGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Redshift Subnet Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.ClusterSubnetGroupName))
	log.Printf("[INFO] Redshift Subnet Group ID: %s", d.Id())
	return resourceSubnetGroupRead(d, meta)
}

func resourceSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := redshift.DescribeClusterSubnetGroupsInput{
		ClusterSubnetGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeClusterSubnetGroups(&describeOpts)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ClusterSubnetGroupNotFoundFault", "") {
			log.Printf("[INFO] Redshift Subnet Group: %s was not found", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(describeResp.ClusterSubnetGroups) == 0 {
		return fmt.Errorf("Unable to find Redshift Subnet Group: %#v", describeResp.ClusterSubnetGroups)
	}

	d.Set("name", d.Id())
	d.Set("description", describeResp.ClusterSubnetGroups[0].Description)
	d.Set("subnet_ids", subnetIdsToSlice(describeResp.ClusterSubnetGroups[0].Subnets))
	tags := tftags.RedshiftKeyValueTags(describeResp.ClusterSubnetGroups[0].Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subnetgroup:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.RedshiftUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Subnet Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

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

		_, err := conn.ModifyClusterSubnetGroup(&redshift.ModifyClusterSubnetGroupInput{
			ClusterSubnetGroupName: aws.String(d.Id()),
			Description:            aws.String(d.Get("description").(string)),
			SubnetIds:              sIds,
		})

		if err != nil {
			return err
		}
	}

	return resourceSubnetGroupRead(d, meta)
}

func resourceSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	_, err := conn.DeleteClusterSubnetGroup(&redshift.DeleteClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(d.Id()),
	})
	if err != nil && tfawserr.ErrMessageContains(err, "ClusterSubnetGroupNotFoundFault", "") {
		return nil
	}

	return err
}

func subnetIdsToSlice(subnetIds []*redshift.Subnet) []string {
	subnetsSlice := make([]string, 0, len(subnetIds))
	for _, s := range subnetIds {
		subnetsSlice = append(subnetsSlice, *s.SubnetIdentifier)
	}
	return subnetsSlice
}
