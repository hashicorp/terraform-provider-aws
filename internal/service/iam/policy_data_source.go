package iam

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePolicyRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"name", "path_prefix"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"arn"},
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"arn"},
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	name := d.Get("name").(string)
	pathPrefix := d.Get("path_prefix").(string)

	if arn == "" {
		raw, err := tfresource.RetryWhenNotFound(propagationTimeout,
			func() (interface{}, error) {
				return FindPolicyByName(conn, name, pathPrefix)
			},
		)

		if errors.Is(err, tfresource.ErrEmptyResult) {
			return fmt.Errorf("no IAM policy found matching criteria (%s); try different search", PolicySearchDetails(name, pathPrefix))
		}
		if errors.Is(err, tfresource.ErrTooManyResults) {
			return fmt.Errorf("multiple IAM policies found matching criteria (%s); try different search. %w", PolicySearchDetails(name, pathPrefix), err)
		}
		if err != nil {
			return fmt.Errorf("error reading IAM policy (%s): %w", PolicySearchDetails(name, pathPrefix), err)
		}

		arn = aws.StringValue((raw.(*iam.Policy)).Arn)
	}

	// We need to make a call to `iam.GetPolicy` because `iam.ListPolicies` doesn't return all values
	policy, err := FindPolicyByARN(conn, arn)
	if err != nil {
		return fmt.Errorf("error reading IAM policy (%s): %w", arn, err)
	}

	policyArn := aws.StringValue(policy.Arn)

	d.SetId(policyArn)
	d.Set("arn", policyArn)
	d.Set("description", policy.Description)
	d.Set("name", policy.PolicyName)
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	if err := d.Set("tags", KeyValueTags(policy.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	// Retrieve policy
	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: policy.Arn,
		VersionId: policy.DefaultVersionId,
	}

	// Handle IAM eventual consistency
	var policyVersionOutput *iam.GetPolicyVersionOutput
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		policyVersionOutput, err = conn.GetPolicyVersion(policyVersionInput)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		policyVersionOutput, err = conn.GetPolicyVersion(policyVersionInput)
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Policy (%s) version: %w", policyArn, err)
	}

	if policyVersionOutput == nil || policyVersionOutput.PolicyVersion == nil {
		return fmt.Errorf("error reading IAM Policy (%s) version: empty output", policyArn)
	}

	policyVersion := policyVersionOutput.PolicyVersion

	var policyDocument string
	if policyVersion != nil {
		policyDocument, err = url.QueryUnescape(aws.StringValue(policyVersion.Document))
		if err != nil {
			return fmt.Errorf("error parsing IAM Policy (%s) document: %w", policyArn, err)
		}
	}

	d.Set("policy", policyDocument)

	return nil
}

// PolicySearchDetails returns the configured search criteria as a printable string
func PolicySearchDetails(name, pathPrefix string) string {
	var policyDetails []string
	if name != "" {
		policyDetails = append(policyDetails, fmt.Sprintf("Name: %s", name))
	}
	if pathPrefix != "" {
		policyDetails = append(policyDetails, fmt.Sprintf("PathPrefix: %s", pathPrefix))
	}

	return strings.Join(policyDetails, ", ")
}
