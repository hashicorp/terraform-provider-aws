package configservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceAggregateAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAggregateAuthorizationPut,
		ReadWithoutTimeout:   resourceAggregateAuthorizationRead,
		UpdateWithoutTimeout: resourceAggregateAuthorizationUpdate,
		DeleteWithoutTimeout: resourceAggregateAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAggregateAuthorizationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	accountId := d.Get("account_id").(string)
	region := d.Get("region").(string)

	req := &configservice.PutAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
		Tags:                Tags(tags.IgnoreAWS()),
	}

	_, err := conn.PutAggregationAuthorizationWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating aggregate authorization: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, region))

	return append(diags, resourceAggregateAuthorizationRead(ctx, d, meta)...)
}

func resourceAggregateAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accountId, region, err := AggregateAuthorizationParseID(d.Id())
	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameAggregateAuthorization, d.Id(), err)
	}

	d.Set("account_id", accountId)
	d.Set("region", region)

	aggregateAuthorizations, err := DescribeAggregateAuthorizations(ctx, conn)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameAggregateAuthorization, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameAggregateAuthorization, d.Id(), err)
	}

	var aggregationAuthorization *configservice.AggregationAuthorization
	// Check for existing authorization
	for _, auth := range aggregateAuthorizations {
		if accountId == aws.StringValue(auth.AuthorizedAccountId) && region == aws.StringValue(auth.AuthorizedAwsRegion) {
			aggregationAuthorization = auth
		}
	}

	if !d.IsNewResource() && aggregationAuthorization == nil {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameAggregateAuthorization, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && aggregationAuthorization == nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameAggregateAuthorization, d.Id(), errors.New("not found after creation"))
	}

	d.Set("arn", aggregationAuthorization.AggregationAuthorizationArn)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Config Aggregate Authorization (%s): %s", d.Get("arn").(string), err)
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

func resourceAggregateAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Config Aggregate Authorization (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceAggregateAuthorizationRead(ctx, d, meta)...)
}

func resourceAggregateAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	accountId, region, err := AggregateAuthorizationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Config Aggregate Authorization (%s): %s", d.Id(), err)
	}

	req := &configservice.DeleteAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountId),
		AuthorizedAwsRegion: aws.String(region),
	}

	_, err = conn.DeleteAggregationAuthorizationWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Config Aggregate Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

func DescribeAggregateAuthorizations(ctx context.Context, conn *configservice.ConfigService) ([]*configservice.AggregationAuthorization, error) {
	aggregationAuthorizations := []*configservice.AggregationAuthorization{}
	input := &configservice.DescribeAggregationAuthorizationsInput{}

	for {
		output, err := conn.DescribeAggregationAuthorizationsWithContext(ctx, input)
		if err != nil {
			return aggregationAuthorizations, err
		}
		aggregationAuthorizations = append(aggregationAuthorizations, output.AggregationAuthorizations...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return aggregationAuthorizations, nil
}

func AggregateAuthorizationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("Please make sure the ID is in the form account_id:region (i.e. 123456789012:us-east-1") // lintignore:AWSAT003
	}
	accountId := idParts[0]
	region := idParts[1]
	return accountId, region, nil
}
