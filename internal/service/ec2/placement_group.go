package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlacementGroupCreate,
		Read:   resourcePlacementGroupRead,
		Update: resourcePlacementGroupUpdate,
		Delete: resourcePlacementGroupDelete,
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
				Required: true,
				ForceNew: true,
			},
			"strategy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.PlacementStrategyCluster,
					ec2.PlacementStrategyPartition,
					ec2.PlacementStrategySpread,
				}, false),
			},
			"placement_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(name),
		Strategy:          aws.String(d.Get("strategy").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypePlacementGroup),
	}
	log.Printf("[DEBUG] Creating EC2 Placement group: %s", input)
	_, err := conn.CreatePlacementGroup(&input)
	if err != nil {
		return err
	}

	wait := resource.StateChangeConf{
		Pending:    []string{ec2.PlacementGroupStatePending},
		Target:     []string{ec2.PlacementGroupStateAvailable},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
				GroupNames: []*string{aws.String(name)},
			})

			if err != nil {
				// Fix timing issue where describe is called prior to
				// create being effectively processed by AWS
				if tfawserr.ErrMessageContains(err, "InvalidPlacementGroup.Unknown", "") {
					return out, "pending", nil
				}
				return out, "", err
			}

			if len(out.PlacementGroups) == 0 {
				return out, "", fmt.Errorf("Placement group not found (%q)", name)
			}
			pg := out.PlacementGroups[0]

			return out, aws.StringValue(pg.State), nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] EC2 Placement group created: %q", name)

	d.SetId(name)

	return resourcePlacementGroupRead(d, meta)
}

func resourcePlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := ec2.DescribePlacementGroupsInput{
		GroupNames: []*string{aws.String(d.Id())},
	}
	out, err := conn.DescribePlacementGroups(&input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidPlacementGroup.Unknown", "") {
			log.Printf("[WARN] Placement Group %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}
	pg := out.PlacementGroups[0]

	log.Printf("[DEBUG] Received EC2 Placement Group: %s", pg)

	d.Set("name", pg.GroupName)
	d.Set("strategy", pg.Strategy)
	d.Set("placement_group_id", pg.GroupId)
	tags := tftags.Ec2KeyValueTags(pg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("placement-group/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourcePlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		pgId := d.Get("placement_group_id").(string)
		if err := tftags.Ec2UpdateTags(conn, pgId, o, n); err != nil {
			return fmt.Errorf("error updating Placement Group (%s) tags: %s", pgId, err)
		}
	}

	return resourcePlacementGroupRead(d, meta)
}

func resourcePlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Placement Group %q", d.Id())
	_, err := conn.DeletePlacementGroup(&ec2.DeletePlacementGroupInput{
		GroupName: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	wait := resource.StateChangeConf{
		Pending:    []string{ec2.PlacementGroupStateAvailable, ec2.PlacementGroupStateDeleting},
		Target:     []string{ec2.PlacementGroupStateDeleted},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
				GroupNames: []*string{aws.String(d.Id())},
			})

			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidPlacementGroup.Unknown", "") {
					return out, ec2.PlacementGroupStateDeleted, nil
				}
				return out, "", err
			}

			if len(out.PlacementGroups) == 0 {
				return out, ec2.PlacementGroupStateDeleted, nil
			}

			pg := out.PlacementGroups[0]
			if aws.StringValue(pg.State) == ec2.PlacementGroupStateAvailable {
				log.Printf("[DEBUG] Accepted status when deleting EC2 Placement group: %q %v", d.Id(),
					aws.StringValue(pg.State))
			}

			return out, aws.StringValue(pg.State), nil
		},
	}

	_, err = wait.WaitForState()
	return err
}
