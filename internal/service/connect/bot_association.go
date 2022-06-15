package connect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBotAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBotAssociationCreate,
		ReadContext:   resourceBotAssociationRead,
		DeleteContext: resourceBotAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"lex_bot": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lex_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(2, 50),
						},
					},
				},
			},
			/* We would need a schema like this to support a v1/v2 hybrid
			"lex_v2_bot": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
							ForceNew: true,
						},
					},
				},
			},
			*/
		},
	}
}

func resourceBotAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceId := d.Get("instance_id").(string)

	input := &connect.AssociateBotInput{
		InstanceId: aws.String(instanceId),
	}

	if v, ok := d.GetOk("lex_bot"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		lexBot := expandLexBot(v.([]interface{}))
		if lexBot.LexRegion == nil {
			lexBot.LexRegion = aws.String(meta.(*conns.AWSClient).Region)
		}
		input.LexBot = lexBot
	}

	/* We would need something like this and additionally the opposite on the above
	if v, ok := d.GetOk("lex_v2_bot"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LexV2Bot = expandLexV2Bot(v.([]interface{}))
	}
	*/

	_, err := tfresource.RetryWhen(
		botAssociationCreateTimeout,
		func() (interface{}, error) {
			return conn.AssociateBotWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, connect.ErrCodeInvalidRequestException) {
				return true, err
			}

			return false, err
		},
	)

	lbaId := BotV1AssociationCreateResourceID(instanceId, aws.StringValue(input.LexBot.Name), aws.StringValue(input.LexBot.LexRegion))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Bot Association (%s): %w", lbaId, err))
	}

	d.SetId(lbaId)

	return resourceBotAssociationRead(ctx, d, meta)
}

func resourceBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceId, name, region, err := BotV1AssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	lexBot, err := FindBotAssociationV1ByNameAndRegionWithContext(ctx, conn, instanceId, name, region)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Bot Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Bot Association (%s): %w", d.Id(), err))
	}

	if lexBot == nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Bot Association (%s): empty output", d.Id()))
	}

	d.Set("instance_id", instanceId)
	if err := d.Set("lex_bot", flattenLexBot(lexBot)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting lex_bot: %w", err))
	}

	return nil
}

func resourceBotAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, name, region, err := BotV1AssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	lexBot := &connect.LexBot{
		Name:      aws.String(name),
		LexRegion: aws.String(region),
	}

	input := &connect.DisassociateBotInput{
		InstanceId: aws.String(instanceID),
		LexBot:     lexBot,
	}

	_, err = conn.DisassociateBotWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Bot Association (%s): %w", d.Id(), err))
	}

	return nil
}

func expandLexBot(l []interface{}) *connect.LexBot {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.LexBot{
		Name: aws.String(tfMap["name"].(string)),
	}

	if v, ok := tfMap["lex_region"].(string); ok && v != "" {
		result.LexRegion = aws.String(v)
	}

	return result
}

func flattenLexBot(bot *connect.LexBot) []interface{} {
	if bot == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"lex_region": bot.LexRegion,
		"name":       bot.Name,
	}

	return []interface{}{m}
}
