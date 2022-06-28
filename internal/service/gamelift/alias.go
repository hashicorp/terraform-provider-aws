package gamelift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliasCreate,
		Read:   resourceAliasRead,
		Update: resourceAliasUpdate,
		Delete: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"routing_strategy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"message": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								gamelift.RoutingStrategyTypeSimple,
								gamelift.RoutingStrategyTypeTerminal,
							}, false),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	rs := expandRoutingStrategy(d.Get("routing_strategy").([]interface{}))
	input := gamelift.CreateAliasInput{
		Name:            aws.String(d.Get("name").(string)),
		RoutingStrategy: rs,
		Tags:            Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating GameLift Alias: %s", input)
	out, err := conn.CreateAlias(&input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(out.Alias.AliasId))

	return resourceAliasRead(d, meta)
}

func resourceAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Describing GameLift Alias: %s", d.Id())
	out, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
		AliasId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			d.SetId("")
			log.Printf("[WARN] GameLift Alias (%s) not found, removing from state", d.Id())
			return nil
		}
		return err
	}
	a := out.Alias

	arn := aws.StringValue(a.AliasArn)
	d.Set("arn", arn)
	d.Set("description", a.Description)
	d.Set("name", a.Name)
	d.Set("routing_strategy", flattenRoutingStrategy(a.RoutingStrategy))
	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Alias (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Updating GameLift Alias: %s", d.Id())
	_, err := conn.UpdateAlias(&gamelift.UpdateAliasInput{
		AliasId:         aws.String(d.Id()),
		Name:            aws.String(d.Get("name").(string)),
		Description:     aws.String(d.Get("description").(string)),
		RoutingStrategy: expandRoutingStrategy(d.Get("routing_strategy").([]interface{})),
	})
	if err != nil {
		return err
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Alias (%s) tags: %s", arn, err)
		}
	}

	return resourceAliasRead(d, meta)
}

func resourceAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Deleting GameLift Alias: %s", d.Id())
	_, err := conn.DeleteAlias(&gamelift.DeleteAliasInput{
		AliasId: aws.String(d.Id()),
	})
	return err
}

func expandRoutingStrategy(cfg []interface{}) *gamelift.RoutingStrategy {
	if len(cfg) < 1 {
		return nil
	}

	strategy := cfg[0].(map[string]interface{})

	out := gamelift.RoutingStrategy{
		Type: aws.String(strategy["type"].(string)),
	}

	if v, ok := strategy["fleet_id"].(string); ok && len(v) > 0 {
		out.FleetId = aws.String(v)
	}
	if v, ok := strategy["message"].(string); ok && len(v) > 0 {
		out.Message = aws.String(v)
	}

	return &out
}

func flattenRoutingStrategy(rs *gamelift.RoutingStrategy) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.FleetId != nil {
		m["fleet_id"] = aws.StringValue(rs.FleetId)
	}
	if rs.Message != nil {
		m["message"] = aws.StringValue(rs.Message)
	}
	m["type"] = aws.StringValue(rs.Type)

	return []interface{}{m}
}
