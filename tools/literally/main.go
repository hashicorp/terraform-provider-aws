// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type ScoringStrategy int

const (
	STANDARD ScoringStrategy = iota
	MULT
	GMEAN
	TEST
	TEST_MULT
	RT_MEAN_SQ
)

type Options struct {
	AllowConsts     bool
	AllowNewlines   bool
	AllowSpaces     bool
	IncludePkgs     bool
	MaxStringLen    int
	MinCount        int
	MinStringLen    int
	MinPkgCount     int
	SchemaOnly      bool
	OutputFile      string
	ScoringStrategy ScoringStrategy
	// Replace mode options
	Replace       bool
	Literal       string
	ConstantName  string
	ConstantsFile string
	PackagePath   string
	DryRun        bool
}

var opts Options

func parseFlags() {
	flag.BoolVar(&opts.AllowNewlines, "allownewline", false, "whether to allow newlines in the counted string literals (messes up CSV output)")
	flag.BoolVar(&opts.AllowSpaces, "allowspace", false, "whether to allow spaces in the counted string literals")
	flag.BoolVar(&opts.AllowConsts, "allowconst", false, "whether to allow string literals assigned to constants to be counted")
	flag.BoolVar(&opts.IncludePkgs, "includepkgs", false, "whether to include the packages the string literal appears in in the output")
	flag.BoolVar(&opts.SchemaOnly, "schemaonly", false, "whether to only include string literals that are keys in a map[string]*schema.Schema")
	flag.IntVar(&opts.MaxStringLen, "maxlen", 50, "the maximum uninterpreted, quoted length of string literals to count")
	flag.IntVar(&opts.MinStringLen, "minlen", 1, "the minimum interpreted, unquoted length of string literals to count")
	flag.IntVar(&opts.MinCount, "mincount", 5, "the minimum count of a string literal to be output")
	flag.IntVar(&opts.MinPkgCount, "minpkgcount", 4, "the number of packages the string literal must appear in to be output")
	flag.StringVar(&opts.OutputFile, "output", "", "the file to write the output to (default is stdout)")

	// Replace mode flags
	flag.BoolVar(&opts.Replace, "replace", false, "replace mode: replace string literals with constants")
	flag.StringVar(&opts.Literal, "literal", "", "the string literal to replace (replace mode)")
	flag.StringVar(&opts.ConstantName, "constant", "", "the constant name to use (replace mode)")
	flag.StringVar(&opts.ConstantsFile, "constants-file", "", "the constants file to update (replace mode)")
	flag.StringVar(&opts.PackagePath, "package", ".", "the package path to process (replace mode)")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "show what would be changed without modifying files (replace mode)")

	var scoringStrategy string
	flag.StringVar(&scoringStrategy, "scoringstrategy", "STANDARD", "the scoring strategy to use (STANDARD, MULT, GMEAN, TEST, TEST_MULT, RT_MEAN_SQ)")

	flag.Parse() // must be after flag declarations and before flag uses

	if opts.Replace {
		if opts.Literal == "" || opts.ConstantName == "" || opts.ConstantsFile == "" {
			log.Fatal("Replace mode requires --literal, --constant, and --constants-file flags")
		}
		return
	}

	fmt.Printf("Scoring strategy: %s\n", scoringStrategy)
	switch scoringStrategy {
	case "STANDARD":
		opts.ScoringStrategy = STANDARD
	case "MULT":
		opts.ScoringStrategy = MULT
	case "GMEAN":
		opts.ScoringStrategy = GMEAN
	case "TEST":
		opts.ScoringStrategy = TEST
	case "TEST_MULT":
		opts.ScoringStrategy = TEST_MULT
	case "RT_MEAN_SQ":
		opts.ScoringStrategy = RT_MEAN_SQ
	default:
		log.Fatalf("Invalid scoring strategy: %s", scoringStrategy)
	}
}

type literal struct {
	count     int
	testCount int
	packages  map[string]int
}

type visitor struct {
	parents  []ast.Node
	literals map[string]literal
	path     string
	isTest   bool
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		// We're going up the tree, so remove the top parent
		v.parents = v.parents[:len(v.parents)-1]

		return v
	}

	// We're going down the tree, so add the current node as a parent
	v.parents = append(v.parents, n)

	x, ok := n.(*ast.BasicLit)

	// Check if the current node is a string literal
	if !ok || x.Kind != token.STRING {
		return v
	}

	// Check if the parent node is an import spec
	if _, ok := v.parents[len(v.parents)-2].(*ast.ImportSpec); ok {
		return v
	}

	// Check if the parent node is a field with a tag
	if field, ok := v.parents[len(v.parents)-2].(*ast.Field); ok && field.Tag != nil && field.Tag.Value == x.Value {
		return v
	}

	if !opts.AllowConsts && v.detectConstant(n) {
		return v
	}

	if opts.SchemaOnly && !v.detectSchemaKey(n) {
		return v
	}

	// Removal criteria: if the string is too long
	if len(x.Value) > opts.MaxStringLen {
		return v
	}

	str, err := strconv.Unquote(x.Value)
	if err != nil {
		log.Fatalf("error unquoting string literal (%s): %v\n", x.Value, err)
	}

	// Removal criteria: if the string is too short
	if len(str) < opts.MinStringLen {
		return v
	}

	// Check if the string contains newlines and if newlines are allowed
	if strings.Contains(str, "\n") && !opts.AllowNewlines {
		return v
	}

	// Check if the string contains spaces and if spaces are allowed
	if strings.Contains(str, " ") && !opts.AllowSpaces {
		return v
	}

	// Increment the count for this string
	if _, ok := v.literals[str]; !ok {
		tc := 0
		if v.isTest {
			tc = 1
		}

		v.literals[str] = literal{
			count:     1, // Initialize count to 1
			testCount: tc,
			packages:  make(map[string]int),
		}

		v.literals[str].packages[v.path] = 1

		return v
	}

	sc := v.literals[str]
	sc.count++
	sc.packages[v.path]++

	if v.isTest {
		sc.testCount++
	}

	v.literals[str] = sc

	return v
}

func (v *visitor) detectSchemaKey(n ast.Node) bool {
	// Check if the parent node is a key-value expression
	keyValueExpr, ok := v.parents[len(v.parents)-2].(*ast.KeyValueExpr)
	if !ok || keyValueExpr.Key != n {
		return false
	}

	// Check if the grandparent node is a composite literal
	compositeLit, ok := v.parents[len(v.parents)-3].(*ast.CompositeLit)
	if !ok {
		return false
	}

	// Check if the type of the composite literal is a map type
	mapType, ok := compositeLit.Type.(*ast.MapType)
	if !ok {
		return false
	}

	// Check if the key type of the map type is a string type
	ident, ok := mapType.Key.(*ast.Ident)
	if !ok || ident.Name != "string" {
		return false
	}

	// Check if the value type of the map type is a pointer to a named type
	starExpr, ok := mapType.Value.(*ast.StarExpr)
	if !ok {
		return false
	}

	// Check if the named type is schema.Schema
	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok || selExpr.Sel.Name != "Schema" {
		return false
	}

	// The string literal is a key for a map of type map[string]*schema.Schema
	return true
}

func (v *visitor) detectConstant(n ast.Node) bool {
	// Check if the parent node is a value spec (variable or constant declaration)
	valueSpec, ok := v.parents[len(v.parents)-2].(*ast.ValueSpec)
	if !ok || opts.AllowConsts {
		return false
	}

	// Check if the grandparent node is a general declaration with the CONST token
	genDecl, ok := v.parents[len(v.parents)-3].(*ast.GenDecl)
	if !ok || genDecl.Tok != token.CONST {
		return false
	}

	// Check if the string literal is assigned to the constant
	for _, value := range valueSpec.Values {
		if value == n {
			// The string literal is assigned to a constant, so return true
			return true
		}
	}

	return false
}

// score filters out literals that don't fit options and then calculates the
// score for each literal based on the scoring strategy and returns a map of
// the scores.
func (v *visitor) scoreLiterals() map[string][]float64 {
	scores := make(map[string][]float64)

	fmt.Printf("Scoring strategy: %d\n", opts.ScoringStrategy)

	for k, v := range v.literals {
		if v.count < opts.MinCount || len(v.packages) < opts.MinPkgCount {
			continue
		}

		var score []float64 // for primary, secondary, tertiary, etc. scoring

		// Components of scores
		pkgCount := float64(len(v.packages))
		length := float64(len(k))
		count := float64(v.count)
		testCount := float64(v.testCount)

		switch opts.ScoringStrategy {
		case STANDARD: // weighted sum
			score = []float64{
				(pkgCount * 0.5) + (length * 0.3) + (count * 0.2),
				pkgCount,
				length,
			}
		case MULT: // multiplicative
			score = []float64{
				pkgCount * length * count,
				pkgCount,
				length,
			}
		case GMEAN: // geometric mean
			score = []float64{
				math.Pow(pkgCount*length*count, 1.0/3.0),
				pkgCount * length,
				count,
			}
		case TEST: // test focus
			score = []float64{
				((pkgCount * 0.4) + (length * 0.3) + (testCount * 0.3)) * (testCount / count),
				testCount,
				pkgCount,
			}
		case TEST_MULT: // test focus multiplicative
			score = []float64{
				pkgCount * length * testCount,
				pkgCount,
				length,
			}
		case RT_MEAN_SQ: // root mean square
			score = []float64{
				math.Sqrt((math.Pow(count, 2) + math.Pow(pkgCount, 2) + math.Pow(length, 2)) / 3),
				length * pkgCount,
				count,
			}
		}

		scores[k] = score
	}

	return scores
}

func (v *visitor) orderLiterals(scores map[string][]float64) []string {
	// Create a slice of keys
	keys := make([]string, 0, len(scores))
	for k := range scores {
		keys = append(keys, k)
	}

	// Sort the keys based on the scores
	sort.Slice(keys, func(i, j int) bool { // nosemgrep:ci.semgrep.stdlib.prefer-slices-sortfunc
		if scores[keys[i]][0] != scores[keys[j]][0] {
			return scores[keys[i]][0] > scores[keys[j]][0]
		}
		if scores[keys[i]][1] != scores[keys[j]][1] {
			return scores[keys[i]][1] > scores[keys[j]][1]
		}
		return scores[keys[i]][2] > scores[keys[j]][2]
	})

	return keys
}

func (v *visitor) output() {
	var out io.Writer

	out = os.Stdout

	if opts.OutputFile != "" && opts.OutputFile != "stdout" {
		file, err := os.Create(opts.OutputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %s", err)
		}
		defer file.Close()
		out = file
	}

	// Write the header
	if !opts.IncludePkgs {
		fmt.Fprintln(out, "Literal,Count,Test Count,Total Packages,Score")
	}

	if opts.IncludePkgs {
		fmt.Fprintln(out, "Literal,Count,Test Count,Total Packages,Packages,Score")
	}

	scores := v.scoreLiterals()
	keys := v.orderLiterals(scores)

	// 6. Output the literals
	for _, k := range keys {
		v := v.literals[k]

		if v.count < opts.MinCount {
			continue
		}

		if len(v.packages) < opts.MinPkgCount {
			continue
		}

		packageKeys := make([]string, 0, len(v.packages))
		for k := range v.packages {
			packageKeys = append(packageKeys, k)
		}

		slices.Sort(packageKeys)

		if !opts.IncludePkgs {
			fmt.Fprintf(out, "%s,%d,%d,%d,%.2f\n", k, v.count, v.testCount, len(v.packages), scores[k][0])
			continue
		}

		fmt.Fprintf(out, "%s,%d,%d,%d,%s,%.2f\n", k, v.count, v.testCount, len(v.packages), strings.Join(packageKeys, "|"), scores[k][0])
	}
}

func main() {
	// Parse the command line flags
	parseFlags()

	// Handle replace mode
	if opts.Replace {
		if err := runReplace(); err != nil {
			log.Fatalf("Replace failed: %v", err)
		}
		return
	}

	fset := token.NewFileSet() // positions are relative to fset

	v := &visitor{
		literals: make(map[string]literal),
	}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		pkgs, err := parser.ParseDir(fset, path, nil, 0)
		if err != nil {
			return err
		}

		v.path = path
		for _, pkg := range pkgs {
			v.isTest = false
			if strings.HasSuffix(pkg.Name, "_test") {
				v.isTest = true
			}
			ast.Walk(v, pkg)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("filepath.Walk error: %s\n", err)
		return
	}

	v.output()
}

// parseConstantsFile parses the constants file and returns the AST
// If the file doesn't exist, it creates a new one with the basic structure
func parseConstantsFile(path string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()

	// Check if file exists
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		// Create new constants file
		dir := filepath.Dir(path)
		pkgName := filepath.Base(dir)

		content := fmt.Sprintf(`// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package %s

// Schema attribute name constants used across package
const ()
`, pkgName)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, nil, fmt.Errorf("creating constants file: %w", err)
		}
		fmt.Printf("Created new constants file: %s\n", path)
	}

	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing constants file: %w", err)
	}
	return fset, f, nil
}

// hasConstant checks if a constant with the given name exists in the file
func hasConstant(f *ast.File, name string) bool {
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, ident := range valueSpec.Names {
				if ident.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// addConstant adds a new constant to the constants file in alphabetical order
func addConstant(f *ast.File, name, value string) error {
	// Find the const block
	var constDecl *ast.GenDecl
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && genDecl.Tok == token.CONST {
			constDecl = genDecl
			break
		}
	}

	if constDecl == nil {
		return fmt.Errorf("no const block found in constants file")
	}

	// Create new constant spec
	newSpec := &ast.ValueSpec{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Values: []ast.Expr{&ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(value),
		}},
	}

	// Find insertion point (alphabetical order)
	insertIdx := len(constDecl.Specs)
	for i, spec := range constDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) == 0 {
			continue
		}
		if valueSpec.Names[0].Name > name {
			insertIdx = i
			break
		}
	}

	// Insert at the correct position
	constDecl.Specs = append(constDecl.Specs[:insertIdx], append([]ast.Spec{newSpec}, constDecl.Specs[insertIdx:]...)...)

	return nil
}

// writeConstantsFile writes the modified AST back to the file
func writeConstantsFile(fset *token.FileSet, f *ast.File, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	if err := format.Node(file, fset, f); err != nil {
		return fmt.Errorf("formatting file: %w", err)
	}

	return nil
}

// replacementVisitor replaces string literals with constant identifiers,
// skipping literals that are values of const declarations.
type replacementVisitor struct {
	literal      string
	constantName string
	replacements int
	parents      []ast.Node
}

func (v *replacementVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		v.parents = v.parents[:len(v.parents)-1]
		return v
	}

	v.parents = append(v.parents, n)

	lit, ok := n.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return v
	}

	str, err := strconv.Unquote(lit.Value)
	if err != nil || str != v.literal {
		return v
	}

	// Skip if this literal is the value of a const declaration
	if v.isConstValue(n) {
		return v
	}

	lit.Kind = token.IDENT
	lit.Value = v.constantName
	v.replacements++

	return v
}

func (v *replacementVisitor) isConstValue(n ast.Node) bool {
	if len(v.parents) < 3 {
		return false
	}
	valueSpec, ok := v.parents[len(v.parents)-2].(*ast.ValueSpec)
	if !ok {
		return false
	}
	genDecl, ok := v.parents[len(v.parents)-3].(*ast.GenDecl)
	if !ok || genDecl.Tok != token.CONST {
		return false
	}
	for _, val := range valueSpec.Values {
		if val == n {
			return true
		}
	}
	return false
}

// replaceInFile replaces string literals in a single file
func replaceInFile(path string, literal, constantName string) (int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("parsing file: %w", err)
	}

	v := &replacementVisitor{
		literal:      literal,
		constantName: constantName,
	}
	ast.Walk(v, f)

	if v.replacements == 0 {
		return 0, nil
	}

	if !opts.DryRun {
		file, err := os.Create(path)
		if err != nil {
			return 0, fmt.Errorf("creating file: %w", err)
		}
		defer file.Close()

		if err := format.Node(file, fset, f); err != nil {
			return 0, fmt.Errorf("formatting file: %w", err)
		}
	}

	return v.replacements, nil
}

// runReplace executes the replacement operation
func runReplace() error {
	// Parse constants file
	fset, constFile, err := parseConstantsFile(opts.ConstantsFile)
	if err != nil {
		return err
	}

	// Check if constant exists, add if not
	if !hasConstant(constFile, opts.ConstantName) {
		fmt.Printf("Adding constant %s = %q to %s\n", opts.ConstantName, opts.Literal, opts.ConstantsFile)
		if err := addConstant(constFile, opts.ConstantName, opts.Literal); err != nil {
			return err
		}
		if !opts.DryRun {
			if err := writeConstantsFile(fset, constFile, opts.ConstantsFile); err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("Constant %s already exists in %s\n", opts.ConstantName, opts.ConstantsFile)
	}

	// Find all .go files in package (excluding _test.go and the constants file itself)
	var files []string
	err = filepath.Walk(opts.PackagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != opts.PackagePath {
			return filepath.SkipDir // Don't recurse
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking package: %w", err)
	}

	// Replace in each file
	totalReplacements := 0
	for _, file := range files {
		count, err := replaceInFile(file, opts.Literal, opts.ConstantName)
		if err != nil {
			return fmt.Errorf("replacing in %s: %w", file, err)
		}
		if count > 0 {
			fmt.Printf("  %s: %d replacement(s)\n", file, count)
			totalReplacements += count
		}
	}

	if opts.DryRun {
		fmt.Printf("\nDry run: would make %d replacement(s)\n", totalReplacements)
	} else {
		fmt.Printf("\nMade %d replacement(s)\n", totalReplacements)
	}

	return nil
}
