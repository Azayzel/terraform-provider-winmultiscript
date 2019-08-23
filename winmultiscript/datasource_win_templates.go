package winmultiscript

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/terraform/helper/schema"
	tflang "github.com/hashicorp/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	ctyconvert "github.com/zclconf/go-cty/cty/convert"
)

type templateRenderError error

// dataSourceFiles Provide the Schema and Resource to the provider
func dataSourceFiles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFilesRead,

		Schema: map[string]*schema.Schema{
			"content_list": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Contents of the template files",
			},
			"base_vars": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     make(map[string]interface{}, 0),
				Description: "variables to substitute",
			},
			"secondary_vars": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     make(map[string]interface{}, 0),
				Description: "variables to substitute in secondary script(s)",
			},
			"rendered": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "rendered template",
			},
		},
	}
}

// dataSourceFilesRead Read and Render provided files
func dataSourceFilesRead(d *schema.ResourceData, meta interface{}) error {
	rendered, err := renderFiles(d)
	if err != nil {
		return err
	}
	d.Set("rendered", rendered)
	fmt.Fprintf(os.Stderr, rendered)
	d.SetId(hash(rendered))
	return nil
}

func renderFiles(d *schema.ResourceData) (string, error) {

	_templates := d.Get("content_list").([]interface{})
	vars := d.Get("base_vars").(map[string]interface{})
	secondaryVars := d.Get("secondary_vars").(map[string]interface{})
	rendered := ""

	// Start final Render string with powershell tag
	allContents := "<powershell> \n"
	for i := 0; i < len(_templates); i++ {
		_template := _templates[i].(string)
		fmt.Fprintf(os.Stderr, "current template %v", _template)

		allContents += _template + "\n"
	}

	// Close up Final Render string
	allContents += "</powershell>"
	rendered, err := execute(allContents, vars, secondaryVars)
	if err != nil {
		return "", templateRenderError(
			fmt.Errorf("failed to render templates: %v", err),
		)
	}

	// return the rendered result or nil
	return rendered, nil
}

// execute parses and executes a template using vars.
func execute(s string, vars map[string]interface{}, secondaryVars map[string]interface{}) (string, error) {
	expr, diags := hclsyntax.ParseTemplate([]byte(s), "<template_file>", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", diags
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Merge vars
	allVars := reduceItem(vars, secondaryVars)

	// Loop merged vars map and assign to ctx, the Eval Context
	for k, v := range allVars {
		// ver 11 TF requires String, when we move to 12, we can add other types
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("unexpected type for variable %q: %T", k, v)
		}
		ctx.Variables[k] = cty.StringVal(s)
	}

	// Use TF's Scope to set Base Dir
	scope := &tflang.Scope{
		BaseDir: ".",
	}

	// Assign TF Funcs to the Context for interpolation
	ctx.Functions = scope.Functions()

	result, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return "", diags
	}

	// convert result to string
	var err error
	result, err = ctyconvert.Convert(result, cty.String)
	if err != nil {
		return "", fmt.Errorf("invalid template result: %s", err)
	}

	// return the resultant string or nil
	return result.AsString(), nil
}

// reduceItem Join Maps
func reduceItem(vars map[string]interface{}, secondaryVars map[string]interface{}) map[string]interface{} {
	for ia, va := range secondaryVars {
		vars[ia] = va

	}
	return vars
}

// hash Make hash to check for validity
func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
