package template

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
	tflang "github.com/hashicorp/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	ctyconvert "github.com/zclconf/go-cty/cty/convert"
)

func dataSourceFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileRead,

		Schema: map[string]*schema.Schema{
			"template": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Contents of the template",
				ConflictsWith: []string{"filename"},
			},
			"filename": {
				Type:     schema.TypeString,
				Optional: true,
				// Make a "best effort" attempt to relativize the file path.
				StateFunc: func(v interface{}) string {
					if v == nil || v.(string) == "" {
						return ""
					}
					pwd, err := os.Getwd()
					if err != nil {
						return v.(string)
					}
					rel, err := filepath.Rel(pwd, v.(string))
					if err != nil {
						return v.(string)
					}
					return rel
				},
				Removed:       "Use the 'template' attribute instead.",
				ConflictsWith: []string{"template"},
			},
			"vars": {
				Type:         schema.TypeMap,
				Optional:     true,
				Default:      make(map[string]interface{}),
				Description:  "variables to substitute",
				ValidateFunc: validateVarsAttribute,
			},
			"rendered": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "rendered template",
			},
		},
	}
}

func dataSourceFileRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderFile(d)
	if err != nil {
		return err
	}
	d.Set("rendered", rendered)
	d.SetId(hash(rendered))
	return nil
}

type templateRenderError error

func renderFile(d *schema.ResourceData) (string, error) {
	template := d.Get("template").(string)
	filename := d.Get("filename").(string)
	vars := d.Get("vars").(map[string]interface{})

	contents := template
	if template == "" && filename != "" {
		data, _, err := pathorcontents.Read(filename)
		if err != nil {
			return "", err
		}

		contents = data
	}

	rendered, err := execute(contents, vars)
	if err != nil {
		return "", templateRenderError(
			fmt.Errorf("failed to render %v: %v", filename, err),
		)
	}

	return rendered, nil
}

// execute parses and executes a template using vars.
func execute(s string, vars map[string]interface{}) (string, error) {
	expr, diags := hclsyntax.ParseTemplate([]byte(s), "<template_file>", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", diags
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}
	for k, v := range vars {
		// In practice today this is always a string due to limitations of
		// the schema system. In future we'd like to support other types here.
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("unexpected type for variable %q: %T", k, v)
		}
		ctx.Variables[k] = cty.StringVal(s)
	}

	// We borrow the functions from Terraform itself here. This is convenient
	// but note that this is coming from whatever version of Terraform we
	// have vendored in to this codebase, not from the version of Terraform
	// the user is running, and so the set of functions won't always match
	// between Terraform itself and this provider.
	// (Over time users will hopefully transition over to Terraform's built-in
	// templatefile function instead and we can phase this provider out.)
	scope := &tflang.Scope{
		BaseDir: ".",
	}
	ctx.Functions = scope.Functions()

	result, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return "", diags
	}

	// Our result must always be a string, so we'll try to convert it.
	var err error
	result, err = ctyconvert.Convert(result, cty.String)
	if err != nil {
		return "", fmt.Errorf("invalid template result: %s", err)
	}

	return result.AsString(), nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func validateVarsAttribute(v interface{}, key string) (ws []string, es []error) {
	// vars can only be primitives right now
	var badVars []string
	for k, v := range v.(map[string]interface{}) {
		switch v.(type) {
		case []interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (list)", k))
		case map[string]interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (map)", k))
		}
	}
	if len(badVars) > 0 {
		es = append(es, fmt.Errorf(
			"%s: cannot contain non-primitives; bad keys: %s",
			key, strings.Join(badVars, ", ")))
	}
	return
}
