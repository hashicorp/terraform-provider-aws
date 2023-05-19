package conns

import tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"

type ProviderMeta struct {
	AccountID         string
	AllowedRegions    []string
	AWSClients        map[string]*AWSClient
	DefaultTagsConfig *tftags.DefaultConfig
	IgnoreTagsConfig  *tftags.IgnoreConfig
	Partition         string
	Region            string
	ServicePackages   map[string]ServicePackage
}
