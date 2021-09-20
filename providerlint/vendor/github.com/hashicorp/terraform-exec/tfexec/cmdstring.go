// +build go1.13

package tfexec

import (
	"os/exec"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// cmdString handles go 1.12 as stringer was only added to exec.Cmd in 1.13
func cmdString(c *exec.Cmd) string {
	return c.String()
}
