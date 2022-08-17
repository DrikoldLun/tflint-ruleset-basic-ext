package rules

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-basic-ext/project"
	"sort"
	"strings"
)

// TerraformOutputOrderRule checks whether the outputs are sorted in expected order
type TerraformOutputOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformOutputOrderRule returns a new rule
func NewTerraformOutputOrderRule() *TerraformOutputOrderRule {
	return &TerraformOutputOrderRule{}
}

// Name returns the rule name
func (r *TerraformOutputOrderRule) Name() string {
	return "terraform_output_order"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformOutputOrderRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformOutputOrderRule) Severity() tflint.Severity {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformOutputOrderRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether the outputs are sorted in expected order
func (r *TerraformOutputOrderRule) Check(runner tflint.Runner) error {

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for filename, file := range files {
		if ignoreFile(filename, r) {
			continue
		}
		if subErr := r.checkOutputOrder(runner, file); subErr != nil {
			err = multierror.Append(err, subErr)
		}
	}
	return err
}

func (r *TerraformOutputOrderRule) checkOutputOrder(runner tflint.Runner, file *hcl.File) error {

	blocks := file.Body.(*hclsyntax.Body).Blocks

	var outputNames []string
	var firstOutputBlockRange hcl.Range
	outputHclTxts := make(map[string]string)
	for _, block := range blocks {
		switch block.Type {
		case "output":
			if IsRangeEmpty(firstOutputBlockRange) {
				firstOutputBlockRange = block.DefRange()
			}
			outputName := block.Labels[0]
			outputNames = append(outputNames, outputName)
			outputHclTxts[outputName] = string(block.Range().SliceBytes(file.Bytes))
		}
	}

	if sort.StringsAreSorted(outputNames) {
		return nil
	}
	sort.Strings(outputNames)
	var sortedOutputHclTxts []string
	for _, outputName := range outputNames {
		sortedOutputHclTxts = append(sortedOutputHclTxts, outputHclTxts[outputName])
	}
	sortedOutputHclBytes := hclwrite.Format([]byte(strings.Join(sortedOutputHclTxts, "\n\n")))
	return runner.EmitIssue(
		r,
		fmt.Sprintf("Recommended output order:\n%s", sortedOutputHclBytes),
		firstOutputBlockRange,
	)
}
