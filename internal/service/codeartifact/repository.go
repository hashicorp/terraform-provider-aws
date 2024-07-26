// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_repository", name="Repository")
// @Tags(identifierAttribute="arn")
func resourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreate,
		ReadWithoutTimeout:   resourceRepositoryRead,
		UpdateWithoutTimeout: resourceRepositoryUpdate,
		DeleteWithoutTimeout: resourceRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"administrator_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"external_connections": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"package_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"upstream": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRepositoryName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	input := &codeartifact.CreateRepositoryInput{
		Domain:     aws.String(d.Get(names.AttrDomain).(string)),
		Repository: aws.String(d.Get("repository").(string)),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		input.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("upstream"); ok {
		input.Upstreams = expandUpstreams(v.([]interface{}))
	}

	output, err := conn.CreateRepository(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Repository: %s", err)
	}

	repository := output.Repository
	d.SetId(aws.ToString(repository.Arn))

	if v, ok := d.GetOk("external_connections"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		externalConnection := tfMap["external_connection_name"].(string)
		input := &codeartifact.AssociateExternalConnectionInput{
			Domain:             repository.DomainName,
			DomainOwner:        repository.DomainOwner,
			ExternalConnection: aws.String(externalConnection),
			Repository:         repository.Name,
		}

		_, err := conn.AssociateExternalConnection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "associating CodeArtifact Repository (%s) external connection (%s): %s", d.Id(), externalConnection, err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, repositoryName, err := parseRepositoryARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	repository, err := findRepositoryByThreePartKey(ctx, conn, owner, domainName, repositoryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeArtifact Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Repository (%s): %s", d.Id(), err)
	}

	d.Set("administrator_account", repository.AdministratorAccount)
	d.Set(names.AttrARN, repository.Arn)
	d.Set(names.AttrDescription, repository.Description)
	d.Set(names.AttrDomain, repository.DomainName)
	d.Set("domain_owner", repository.DomainOwner)
	if err := d.Set("external_connections", flattenRepositoryExternalConnectionInfos(repository.ExternalConnections)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting external_connections: %s", err)
	}
	d.Set("repository", repository.Name)
	if err := d.Set("upstream", flattenUpstreamRepositoryInfos(repository.Upstreams)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting upstream: %s", err)
	}

	return diags
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, repositoryName, err := parseRepositoryARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges(names.AttrDescription, "upstream") {
		input := &codeartifact.UpdateRepositoryInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(owner),
			Repository:  aws.String(repositoryName),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("upstream") {
			if v, ok := d.GetOk("upstream"); ok && len(v.([]interface{})) > 0 {
				input.Upstreams = expandUpstreams(v.([]interface{}))
			}
		}

		_, err := conn.UpdateRepository(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeArtifact Repository (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("external_connections") {
		if v, ok := d.GetOk("external_connections"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})
			externalConnection := tfMap["external_connection_name"].(string)
			input := &codeartifact.AssociateExternalConnectionInput{
				Domain:             aws.String(domainName),
				DomainOwner:        aws.String(owner),
				ExternalConnection: aws.String(externalConnection),
				Repository:         aws.String(repositoryName),
			}

			_, err := conn.AssociateExternalConnection(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating CodeArtifact Repository (%s) external connection (%s): %s", d.Id(), externalConnection, err)
			}
		} else {
			o, _ := d.GetChange("external_connections")
			tfMap := o.([]interface{})[0].(map[string]interface{})
			externalConnection := tfMap["external_connection_name"].(string)
			input := &codeartifact.DisassociateExternalConnectionInput{
				Domain:             aws.String(domainName),
				DomainOwner:        aws.String(owner),
				ExternalConnection: aws.String(externalConnection),
				Repository:         aws.String(repositoryName),
			}

			_, err := conn.DisassociateExternalConnection(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating CodeArtifact Repository (%s) external connection (%s): %s", d.Id(), externalConnection, err)
			}
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, repositoryName, err := parseRepositoryARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting CodeArtifact Repository: %s", d.Id())
	_, err = conn.DeleteRepository(ctx, &codeartifact.DeleteRepositoryInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
		Repository:  aws.String(repositoryName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository (%s): %s", d.Id(), err)
	}

	return diags
}

func parseRepositoryARN(v string) (string, string, string, error) {
	// arn:${Partition}:codeartifact:${Region}:${Account}:repository/${DomainName}/${RepositoryName}
	arn, err := arn.Parse(v)
	if err != nil {
		return "", "", "", err
	}

	parts := strings.Split(strings.TrimPrefix(arn.Resource, "repository/"), "/")
	if len(parts) != 2 {
		return "", "", "", errors.New("invalid repository ARN")
	}

	return arn.AccountID, parts[0], parts[1], nil
}

func findRepositoryByThreePartKey(ctx context.Context, conn *codeartifact.Client, owner, domainName, repositoryName string) (*types.RepositoryDescription, error) {
	input := &codeartifact.DescribeRepositoryInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
		Repository:  aws.String(repositoryName),
	}

	output, err := conn.DescribeRepository(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Repository == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Repository, nil
}

func expandUpstreams(tfList []interface{}) []types.UpstreamRepository {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []types.UpstreamRepository{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.UpstreamRepository{}

		if v, ok := tfMap[names.AttrRepositoryName].(string); ok && v != "" {
			apiObject.RepositoryName = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenUpstreamRepositoryInfos(apiObjects []types.UpstreamRepositoryInfo) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if v := apiObject.RepositoryName; v != nil {
			tfMap[names.AttrRepositoryName] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRepositoryExternalConnectionInfos(apiObjects []types.RepositoryExternalConnectionInfo) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"package_format": apiObject.PackageFormat,
			names.AttrStatus: apiObject.Status,
		}

		if v := apiObject.ExternalConnectionName; v != nil {
			tfMap["external_connection_name"] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
