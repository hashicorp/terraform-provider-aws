// +build !windows

package url

import (
	"net/url"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func parse(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}
