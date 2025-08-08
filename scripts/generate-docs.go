package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Component represents a code component to document
type Component struct {
	Name        string
	Package     string
	Type        string // struct, interface, function
	Description string
	File        string
	Line        int
	Methods     []Method
	Fields      []Field
}

// Method represents a method or function
type Method struct {
	Name        string
	Description string
	Parameters  []string
	Returns     []string
	Receiver    string
}

// Field represents a struct field
type Field struct {
	Name        string
	Type        string
	Description string
	Tags        string
}

// Architecture represents the overall architecture
type Architecture struct {
	Components []Component
	Packages   map[string]PackageInfo
}

// PackageInfo represents package-level information
type PackageInfo struct {
	Name        string
	Path        string
	Description string
	Imports     []string
	Components  []Component
}

var (
	srcDir    = flag.String("src", ".", "Source directory to scan")
	outputDir = flag.String("out", "docs/generated", "Output directory")
	format    = flag.String("format", "markdown", "Output format (markdown, json, html)")
	verbose   = flag.Bool("v", false, "Verbose output")
)

// mustFprintf is a helper that panics on fprintf errors (for simplicity in this tool)
func mustFprintf(w io.Writer, format string, args ...interface{}) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}
}

func main() {
	flag.Parse()

	if *verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Scan source code
	arch := &Architecture{
		Components: []Component{},
		Packages:   make(map[string]PackageInfo),
	}

	if err := scanDirectory(*srcDir, arch); err != nil {
		log.Fatalf("Failed to scan directory: %v", err)
	}

	// Generate documentation
	switch *format {
	case "markdown":
		if err := generateMarkdown(arch); err != nil {
			log.Fatalf("Failed to generate markdown: %v", err)
		}
	case "json":
		if err := generateJSON(arch); err != nil {
			log.Fatalf("Failed to generate JSON: %v", err)
		}
	case "html":
		if err := generateHTML(arch); err != nil {
			log.Fatalf("Failed to generate HTML: %v", err)
		}
	default:
		log.Fatalf("Unknown format: %s", *format)
	}

	fmt.Printf("✅ Documentation generated in %s\n", *outputDir)
}

func scanDirectory(dir string, arch *Architecture) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "_test.go") {
			return nil
		}

		// Skip vendor and hidden directories
		if strings.Contains(path, "vendor/") || strings.Contains(path, "/.") {
			return nil
		}

		if *verbose {
			log.Printf("Scanning %s", path)
		}

		// Parse Go file
		if err := parseGoFile(path, arch); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", path, err)
		}

		return nil
	})
}

func parseGoFile(filename string, arch *Architecture) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return err
	}

	// Extract package info
	pkgName := node.Name.Name
	if _, exists := arch.Packages[pkgName]; !exists {
		arch.Packages[pkgName] = PackageInfo{
			Name:       pkgName,
			Path:       filepath.Dir(filename),
			Components: []Component{},
		}
	}

	// Extract package description from comments
	if node.Doc != nil {
		pkg := arch.Packages[pkgName]
		pkg.Description = node.Doc.Text()
		arch.Packages[pkgName] = pkg
	}

	// Extract components
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			// Extract struct/interface definitions
			if structType, ok := x.Type.(*ast.StructType); ok {
				comp := extractStruct(x, structType, fset, filename)
				arch.Components = append(arch.Components, comp)

				pkg := arch.Packages[pkgName]
				pkg.Components = append(pkg.Components, comp)
				arch.Packages[pkgName] = pkg
			}
		case *ast.FuncDecl:
			// Extract function definitions
			if x.Name.IsExported() {
				method := extractFunction(x)
				// Add to receiver's component if applicable
				if x.Recv != nil {
					updateComponentMethod(arch, method, x.Recv)
				}
			}
		}
		return true
	})

	return nil
}

func extractStruct(spec *ast.TypeSpec, structType *ast.StructType, fset *token.FileSet, filename string) Component {
	comp := Component{
		Name:    spec.Name.Name,
		Type:    "struct",
		File:    filename,
		Fields:  []Field{},
		Methods: []Method{},
	}

	// Extract struct comment
	if spec.Doc != nil {
		comp.Description = spec.Doc.Text()
	}

	// Extract position
	pos := fset.Position(spec.Pos())
	comp.Line = pos.Line

	// Extract fields
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.IsExported() {
				f := Field{
					Name: name.Name,
					Type: exprToString(field.Type),
				}

				if field.Doc != nil {
					f.Description = field.Doc.Text()
				}

				if field.Tag != nil {
					f.Tags = field.Tag.Value
				}

				comp.Fields = append(comp.Fields, f)
			}
		}
	}

	return comp
}

func extractFunction(fn *ast.FuncDecl) Method {
	method := Method{
		Name:       fn.Name.Name,
		Parameters: []string{},
		Returns:    []string{},
	}

	// Extract documentation
	if fn.Doc != nil {
		method.Description = fn.Doc.Text()
	}

	// Extract parameters
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			paramType := exprToString(param.Type)
			for _, name := range param.Names {
				method.Parameters = append(method.Parameters, fmt.Sprintf("%s %s", name.Name, paramType))
			}
		}
	}

	// Extract returns
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			method.Returns = append(method.Returns, exprToString(result.Type))
		}
	}

	// Extract receiver
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		method.Receiver = exprToString(recv.Type)
	}

	return method
}

func exprToString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + exprToString(x.X)
	case *ast.ArrayType:
		return "[]" + exprToString(x.Elt)
	case *ast.SelectorExpr:
		return exprToString(x.X) + "." + x.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func updateComponentMethod(arch *Architecture, method Method, recv *ast.FieldList) {
	if len(recv.List) == 0 {
		return
	}

	recvType := strings.TrimPrefix(exprToString(recv.List[0].Type), "*")

	for i, comp := range arch.Components {
		if comp.Name == recvType {
			arch.Components[i].Methods = append(arch.Components[i].Methods, method)
			break
		}
	}
}

func generateMarkdown(arch *Architecture) error {
	// Main documentation file
	mainFile := filepath.Join(*outputDir, "code-architecture.md")
	f, err := os.Create(mainFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	w := bufio.NewWriter(f)
	defer func() {
		if err := w.Flush(); err != nil {
			log.Printf("Error flushing writer: %v", err)
		}
	}()

	// Write header
	mustFprintf(w, "# Code Architecture Documentation\n\n")
	mustFprintf(w, "Auto-generated from source code\n\n")
	mustFprintf(w, "## Table of Contents\n\n")

	// Write package index
	mustFprintf(w, "### Packages\n\n")
	for name, pkg := range arch.Packages {
		mustFprintf(w, "- [%s](%s) - %s\n", name, filepath.Join(pkg.Path, "README.md"),
			strings.Split(pkg.Description, "\n")[0])
	}

	mustFprintf(w, "\n## Components\n\n")

	// Group components by package
	for pkgName, pkg := range arch.Packages {
		if len(pkg.Components) == 0 {
			continue
		}

		mustFprintf(w, "### Package: %s\n\n", pkgName)

		for _, comp := range pkg.Components {
			mustFprintf(w, "#### %s\n\n", comp.Name)
			mustFprintf(w, "**Type**: %s  \n", comp.Type)
			mustFprintf(w, "**File**: `%s:%d`  \n\n", comp.File, comp.Line)

			if comp.Description != "" {
				mustFprintf(w, "%s\n\n", comp.Description)
			}

			// Write fields
			if len(comp.Fields) > 0 {
				mustFprintf(w, "**Fields**:\n\n")
				mustFprintf(w, "| Field | Type | Description |\n")
				mustFprintf(w, "|-------|------|-------------|\n")
				for _, field := range comp.Fields {
					desc := strings.ReplaceAll(field.Description, "\n", " ")
					mustFprintf(w, "| %s | `%s` | %s |\n", field.Name, field.Type, desc)
				}
				mustFprintf(w, "\n")
			}

			// Write methods
			if len(comp.Methods) > 0 {
				mustFprintf(w, "**Methods**:\n\n")
				for _, method := range comp.Methods {
					mustFprintf(w, "- `%s(%s)`", method.Name, strings.Join(method.Parameters, ", "))
					if len(method.Returns) > 0 {
						mustFprintf(w, " → (%s)", strings.Join(method.Returns, ", "))
					}
					mustFprintf(w, "`\n")
					if method.Description != "" {
						mustFprintf(w, "  %s\n", strings.ReplaceAll(method.Description, "\n", "\n  "))
					}
				}
				mustFprintf(w, "\n")
			}
		}
	}

	return nil
}

func generateJSON(arch *Architecture) error {
	outputFile := filepath.Join(*outputDir, "architecture.json")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(arch)
}

func generateHTML(arch *Architecture) error {
	tmplStr := `<!DOCTYPE html>
<html>
<head>
    <title>Architecture Documentation</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 40px; }
        h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }
        h2 { color: #4CAF50; margin-top: 30px; }
        h3 { color: #666; }
        .component { background: #f5f5f5; padding: 15px; margin: 15px 0; border-radius: 5px; }
        .field { margin: 5px 0; padding: 5px; background: white; }
        .method { margin: 5px 0; padding: 8px; background: #e8f5e9; border-left: 3px solid #4CAF50; }
        code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #4CAF50; color: white; }
    </style>
</head>
<body>
    <h1>Architecture Documentation</h1>
    {{range $pkg := .Packages}}
    <h2>Package: {{$pkg.Name}}</h2>
    <p>{{$pkg.Description}}</p>
    {{range $comp := $pkg.Components}}
    <div class="component">
        <h3>{{$comp.Name}}</h3>
        <p><strong>Type:</strong> {{$comp.Type}}</p>
        <p>{{$comp.Description}}</p>
        {{if $comp.Fields}}
        <h4>Fields</h4>
        <table>
            <tr><th>Name</th><th>Type</th><th>Description</th></tr>
            {{range $field := $comp.Fields}}
            <tr><td>{{$field.Name}}</td><td><code>{{$field.Type}}</code></td><td>{{$field.Description}}</td></tr>
            {{end}}
        </table>
        {{end}}
        {{if $comp.Methods}}
        <h4>Methods</h4>
        {{range $method := $comp.Methods}}
        <div class="method">
            <code>{{$method.Name}}({{range $i, $p := $method.Parameters}}{{if $i}}, {{end}}{{$p}}{{end}})</code>
            {{if $method.Returns}}<code> → ({{range $i, $r := $method.Returns}}{{if $i}}, {{end}}{{$r}}{{end}})</code>{{end}}
            <p>{{$method.Description}}</p>
        </div>
        {{end}}
        {{end}}
    </div>
    {{end}}
    {{end}}
</body>
</html>`

	outputFile := filepath.Join(*outputDir, "architecture.html")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	tmpl, err := template.New("html").Parse(tmplStr)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, arch)
}
