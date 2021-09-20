package getter

import (
	"io/ioutil"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func tmpFile(dir, pattern string) (string, error) {
	f, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
