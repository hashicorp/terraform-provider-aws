// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type ListResourceWithSDKv2Identity struct {
	resource *schema.Resource
	identity *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Identity) WithTranslatedIdentity(attrs []inttypes.IdentityAttribute) {
	var identity schema.ResourceIdentity
	out := make(map[string]*schema.Schema)
	for _, v := range attrs {
		out[v.Name()] = &schema.Schema{
			Type: schema.TypeString,
		}
		if v.Required() {
			out[v.Name()].Required = true
		} else {
			out[v.Name()].Optional = true
		}
	}

	identity.SchemaFunc = func() map[string]*schema.Schema {
		return out
	}

	l.identity = &identity
	l.resource.Identity = &identity
}

func (l *ListResourceWithSDKv2Identity) GetResource() *schema.Resource {
	return l.resource
}

func (l *ListResourceWithSDKv2Identity) SetResource(resource *schema.Resource) {
	l.resource = resource
}
