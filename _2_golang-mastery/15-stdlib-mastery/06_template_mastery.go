//go:build ignore

// =============================================================================
// LESSON 15.6: text/template & html/template — Go's Template Engine
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Template syntax: actions, pipelines, variables
// - text/template vs html/template (and why you MUST use html for web)
// - Control flow: if, range, with, block, define
// - Functions: built-in, custom FuncMap
// - Template composition: nested templates, layouts, inheritance
// - html/template auto-escaping and security model
// - Production patterns: email templates, code generation, config files
//
// THE KEY INSIGHT:
// Go templates are LOGIC-LIGHT by design. They prevent you from putting
// business logic in templates (unlike Jinja2, ERB, etc.). This feels
// limiting at first but forces clean separation of concerns.
// html/template is contextually auto-escaped — it knows whether you're
// in HTML, CSS, JS, or URL context and escapes accordingly.
//
// RUN: go run 06_template_mastery.go
// =============================================================================

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	texttemplate "text/template"
	"time"
)

func main() {
	fmt.Println("=== TEMPLATE MASTERY ===")
	fmt.Println()

	templateBasics()
	controlFlow()
	templateFunctions()
	templateComposition()
	htmlTemplateSecurity()
	productionPatterns()
}

// =============================================================================
// PART 1: Template Basics
// =============================================================================
func templateBasics() {
	fmt.Println("--- TEMPLATE BASICS ---")

	// TEMPLATE SYNTAX:
	// ─────────────────
	// {{ . }}           — current value (the "dot")
	// {{ .Name }}       — field access
	// {{ .Method }}     — method call (must return 1 or 2 values)
	// {{ $var := . }}   — variable assignment
	// {{ /* comment */}} — comment (stripped from output)
	// {{- .Name -}}    — trim whitespace (- left trims before, - right trims after)

	// ─── Simple template ───
	t := texttemplate.Must(texttemplate.New("hello").Parse(
		"Hello, {{ .Name }}! You have {{ .Count }} messages.\n",
	))
	t.Execute(os.Stdout, struct {
		Name  string
		Count int
	}{"Vikram", 42})

	// ─── Dot (.) is the current context ───
	// When you pass a string, dot IS the string
	t2 := texttemplate.Must(texttemplate.New("dot").Parse(
		"  The value is: {{ . }}\n",
	))
	t2.Execute(os.Stdout, "just a string")

	// ─── Variables ───
	t3 := texttemplate.Must(texttemplate.New("vars").Parse(
		"  {{ $name := .Name }}{{ $name }} (saved to $name)\n",
	))
	t3.Execute(os.Stdout, struct{ Name string }{"Vikram"})

	// ─── Pipelines (like Unix pipes) ───
	// {{ .Name | printf "%q" }}  — pipe .Name into printf
	t4 := texttemplate.Must(texttemplate.New("pipe").Parse(
		"  Quoted: {{ .Name | printf \"%q\" }}\n",
	))
	t4.Execute(os.Stdout, struct{ Name string }{"Vikram"})

	// ─── Whitespace control ───
	// {{- trims whitespace before, -}} trims after
	t5 := texttemplate.Must(texttemplate.New("ws").Parse(
		"  items: [{{- range .Items }} {{ . }} {{- end -}}]\n",
	))
	t5.Execute(os.Stdout, struct{ Items []string }{[]string{"a", "b", "c"}})

	fmt.Println()
}

// =============================================================================
// PART 2: Control Flow
// =============================================================================
func controlFlow() {
	fmt.Println("--- CONTROL FLOW ---")

	// ─── if/else ───
	// "truthy" in templates: non-zero, non-nil, non-empty, non-false
	ifTmpl := texttemplate.Must(texttemplate.New("if").Parse(
		`  {{ if .IsAdmin }}Admin{{ else if .IsMod }}Moderator{{ else }}User{{ end }}
`))
	ifTmpl.Execute(os.Stdout, struct{ IsAdmin, IsMod bool }{false, true})

	// ─── Comparison operators ───
	// eq, ne, lt, le, gt, ge — these are FUNCTIONS, not operators!
	// {{ if eq .Status "active" }} ... {{ end }}
	// {{ if gt .Age 18 }} ... {{ end }}
	// {{ if and .IsAdmin (ne .Status "banned") }} ... {{ end }}
	// {{ if or .IsAdmin .IsMod }} ... {{ end }}
	// {{ if not .Disabled }} ... {{ end }}
	cmpTmpl := texttemplate.Must(texttemplate.New("cmp").Parse(
		`  {{ if eq .Status "active" }}Active!{{ end }} {{ if gt .Score 90 }}High score!{{ end }}
`))
	cmpTmpl.Execute(os.Stdout, struct {
		Status string
		Score  int
	}{"active", 95})

	// ─── range: iterate over slices, maps, channels ───
	rangeTmpl := texttemplate.Must(texttemplate.New("range").Parse(
		`  Users:{{ range . }}
    - {{ .Name }} ({{ .Age }}){{ end }}
`))
	rangeTmpl.Execute(os.Stdout, []struct {
		Name string
		Age  int
	}{{"Alice", 30}, {"Bob", 25}, {"Charlie", 35}})

	// ─── range with index ───
	idxTmpl := texttemplate.Must(texttemplate.New("idx").Parse(
		`  Indexed:{{ range $i, $v := . }}
    {{ $i }}: {{ $v }}{{ end }}
`))
	idxTmpl.Execute(os.Stdout, []string{"Go", "Rust", "Python"})

	// ─── range over map ───
	mapTmpl := texttemplate.Must(texttemplate.New("map").Parse(
		`  Map:{{ range $k, $v := . }}
    {{ $k }} = {{ $v }}{{ end }}
`))
	mapTmpl.Execute(os.Stdout, map[string]int{"apples": 5, "bananas": 3})

	// ─── range with else (empty collection) ───
	emptyTmpl := texttemplate.Must(texttemplate.New("empty").Parse(
		`  {{ range .Items }}{{ .}}{{ else }}No items found{{ end }}
`))
	emptyTmpl.Execute(os.Stdout, struct{ Items []string }{nil})

	// ─── with: rebind dot to a sub-value ───
	// with also acts as a nil check (skips block if nil/zero/empty)
	withTmpl := texttemplate.Must(texttemplate.New("with").Parse(
		`  {{ with .Address }}City: {{ .City }}, State: {{ .State }}{{ else }}No address{{ end }}
`))
	type Address struct{ City, State string }
	withTmpl.Execute(os.Stdout, struct{ Address *Address }{&Address{"SF", "CA"}})
	withTmpl.Execute(os.Stdout, struct{ Address *Address }{nil})

	fmt.Println()
}

// =============================================================================
// PART 3: Custom Functions
// =============================================================================
func templateFunctions() {
	fmt.Println("--- TEMPLATE FUNCTIONS ---")

	// BUILT-IN FUNCTIONS:
	// ───────────────────
	// and, or, not     — boolean logic
	// eq, ne, lt, le, gt, ge — comparison
	// len              — length of array, slice, map, string
	// index            — index into array/slice/map: {{ index .Map "key" }}
	// print, printf, println — fmt equivalents
	// call             — call a func value: {{ call .Func arg1 arg2 }}
	// html, js, urlquery — escaping (text/template only; html/template auto-escapes)
	// slice            — sub-slice: {{ slice .Items 1 3 }}

	// ─── Custom FuncMap ───
	// Add your own functions to templates
	funcMap := texttemplate.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title, //nolint
		"join":     strings.Join,
		"contains": strings.Contains,
		"repeat":   strings.Repeat,
		"add":      func(a, b int) int { return a + b },
		"sub":      func(a, b int) int { return a - b },
		"mul":      func(a, b int) int { return a * b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"now": func() string {
			return time.Now().Format("2006-01-02")
		},
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 || val == false {
				return def
			}
			return val
		},
	}

	t := texttemplate.Must(texttemplate.New("funcs").Funcs(funcMap).Parse(
		`  Upper: {{ .Name | upper }}
  Add: {{ add 10 20 }}
  Join: {{ .Tags | join ", " }}
  Seq: {{ range seq 3 }}[{{ . }}]{{ end }}
  Default: {{ default "N/A" .Missing }}
`))
	t.Execute(os.Stdout, struct {
		Name    string
		Tags    []string
		Missing string
	}{"Vikram", []string{"go", "kafka", "k8s"}, ""})

	// ─── IMPORTANT: FuncMap must be added BEFORE Parse() ───
	// template.New("x").Funcs(funcMap).Parse(...)  ✓
	// template.New("x").Parse(...).Funcs(funcMap)  ✗ panic!

	fmt.Println()
}

// =============================================================================
// PART 4: Template Composition — Nesting & Layouts
// =============================================================================
func templateComposition() {
	fmt.Println("--- TEMPLATE COMPOSITION ---")

	// ─── define + template: reusable blocks ───
	const layout = `
{{- define "header" -}}
=== {{ .Title }} ===
{{ end -}}

{{- define "footer" -}}
---
Generated: {{ .Date }}
{{ end -}}

{{- template "header" . -}}
Content: {{ .Body }}
{{ template "footer" . -}}
`

	t := texttemplate.Must(texttemplate.New("page").Parse(layout))
	t.Execute(os.Stdout, struct {
		Title, Body, Date string
	}{"My Page", "Hello from composed templates!", "2024-11-15"})

	// ─── block: define with default content (overridable) ───
	const baseTmpl = `
{{- define "base" -}}
<html>
<head><title>{{ block "title" . }}Default Title{{ end }}</title></head>
<body>{{ block "content" . }}Default content{{ end }}</body>
</html>
{{ end -}}
`
	// A child template can override "title" and "content" blocks
	// by defining them with the same name.

	// ─── ParseFiles: load templates from files ───
	// t, err := template.ParseFiles("base.html", "page.html")
	// t.ExecuteTemplate(w, "base", data)  // execute a specific named template
	//
	// ParseGlob: load all matching files
	// t, err := template.ParseGlob("templates/*.html")

	// ─── PRODUCTION LAYOUT PATTERN ───
	// templates/
	//   layouts/
	//     base.html       → {{ define "base" }}...{{ block "content" . }}{{ end }}
	//   pages/
	//     home.html       → {{ define "content" }}...{{ end }}
	//     about.html      → {{ define "content" }}...{{ end }}
	//   partials/
	//     header.html     → {{ define "header" }}...{{ end }}
	//     footer.html     → {{ define "footer" }}...{{ end }}
	//
	// // Load once at startup (not per request!)
	// baseTmpl := template.Must(template.ParseFiles(
	//     "templates/layouts/base.html",
	//     "templates/partials/header.html",
	//     "templates/partials/footer.html",
	// ))
	//
	// // Per page: clone base + add page content
	// pageTmpl, _ := template.Must(baseTmpl.Clone()).ParseFiles("templates/pages/home.html")
	// pageTmpl.ExecuteTemplate(w, "base", data)

	fmt.Println("  ParseFiles for file-based templates")
	fmt.Println("  Clone() + ParseFiles for layout inheritance")

	_ = baseTmpl // used for documentation

	fmt.Println()
}

// =============================================================================
// PART 5: html/template — Security & Auto-Escaping
// =============================================================================
func htmlTemplateSecurity() {
	fmt.Println("--- HTML TEMPLATE SECURITY ---")

	// html/template has the SAME API as text/template but with auto-escaping.
	// It CONTEXTUALLY escapes output based on where it appears:
	//
	// CONTEXT           ESCAPING
	// HTML body          < → &lt;  > → &gt;  & → &amp;
	// HTML attribute     " → &#34;  ' → &#39;  + standard HTML escaping
	// CSS                escapes for CSS context
	// JavaScript         escapes for JS string context
	// URL                escapes for URL context
	//
	// This prevents XSS (Cross-Site Scripting) attacks automatically!

	// ─── Auto-escaping in action ───
	tmpl := template.Must(template.New("safe").Parse(
		`  HTML: {{ .HTML }}
  URL:  <a href="{{ .URL }}">link</a>
  JS:   <script>var x = "{{ .JS }}";</script>
`))

	var buf bytes.Buffer
	tmpl.Execute(&buf, struct {
		HTML, URL, JS string
	}{
		HTML: `<script>alert("xss")</script>`,
		URL:  `javascript:alert('xss')`,
		JS:   `"; alert("xss"); "`,
	})
	fmt.Print(buf.String())
	// All injection attempts are safely escaped!

	// ─── Marking content as safe (when YOU trust it) ───
	// template.HTML("trusted HTML")  — renders raw HTML (DANGEROUS if from user input!)
	// template.CSS("trusted CSS")
	// template.JS("trusted JS")
	// template.URL("trusted URL")
	// template.HTMLAttr("trusted attr")
	//
	// ONLY use these for content YOU control (not user input)!
	safeTmpl := template.Must(template.New("trusted").Parse(
		"  Trusted HTML: {{ .TrustedHTML }}\n  Escaped HTML: {{ .UserHTML }}\n",
	))
	var buf2 bytes.Buffer
	safeTmpl.Execute(&buf2, struct {
		TrustedHTML template.HTML
		UserHTML    string
	}{
		TrustedHTML: template.HTML("<strong>bold</strong>"), // rendered as HTML
		UserHTML:    "<strong>this gets escaped</strong>",   // escaped
	})
	fmt.Print(buf2.String())

	// ─── WHEN TO USE WHICH PACKAGE ───
	// html/template: ANY output that goes to a web browser (HTML, JSON in HTML, etc.)
	// text/template: config files, code generation, emails (plain text), CLI output
	//
	// RULE: If the output could be displayed in a browser, use html/template.
	//       NEVER use text/template for HTML output!

	fmt.Println()
}

// =============================================================================
// PART 6: Production Patterns
// =============================================================================
func productionPatterns() {
	fmt.Println("--- PRODUCTION PATTERNS ---")

	// ─── PATTERN 1: Code generation with text/template ───
	// This is how `go generate` tools work (stringer, enumer, etc.)
	const codeTemplate = `// Code generated by tool; DO NOT EDIT.
package {{ .Package }}

type {{ .TypeName }} int

const (
{{- range $i, $v := .Values }}
	{{ $v }} {{ if eq $i 0 }}{{ $.TypeName }} = iota{{ end }}
{{- end }}
)

var _{{ .TypeName }}Names = map[{{ .TypeName }}]string{
{{- range .Values }}
	{{ . }}: "{{ . }}",
{{- end }}
}

func (t {{ .TypeName }}) String() string {
	if name, ok := _{{ .TypeName }}Names[t]; ok {
		return name
	}
	return "unknown"
}
`

	t := texttemplate.Must(texttemplate.New("codegen").Parse(codeTemplate))
	var buf bytes.Buffer
	t.Execute(&buf, struct {
		Package  string
		TypeName string
		Values   []string
	}{"mypackage", "Color", []string{"Red", "Green", "Blue"}})
	fmt.Printf("  Code generation output:\n%s\n", buf.String())

	// ─── PATTERN 2: Email templates ───
	// const emailTmpl = `Subject: {{ .Subject }}
	// Dear {{ .Name }},
	// {{ .Body }}
	// {{ if .HasCTA }}Visit: {{ .CTAURL }}{{ end }}
	// Best regards,
	// {{ .SenderName }}`
	//
	// IMPORTANT: Use text/template for emails (not html/template)
	// unless sending HTML emails.
	fmt.Println("  Pattern 2: text/template for email generation")

	// ─── PATTERN 3: Template caching ───
	// Parse templates ONCE at startup, not per request!
	//
	// var templates *template.Template
	//
	// func init() {
	//     templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
	// }
	//
	// func handler(w http.ResponseWriter, r *http.Request) {
	//     templates.ExecuteTemplate(w, "home.html", data)
	// }
	//
	// For development, reload templates on each request:
	// if isDev {
	//     templates = template.Must(template.ParseGlob("templates/*.html"))
	// }
	fmt.Println("  Pattern 3: Parse templates once at startup, not per request")

	// ─── PATTERN 4: Render to buffer first (check errors before writing response) ───
	// func render(w http.ResponseWriter, tmpl string, data interface{}) {
	//     var buf bytes.Buffer
	//     if err := templates.ExecuteTemplate(&buf, tmpl, data); err != nil {
	//         http.Error(w, "Internal Server Error", 500)
	//         return
	//     }
	//     buf.WriteTo(w)  // only write to response if template succeeded
	// }
	//
	// WHY: If you Execute directly to http.ResponseWriter and the template
	// fails halfway, you've already sent partial HTML + 200 status code.
	// By buffering first, you can send a proper 500 error instead.
	fmt.Println("  Pattern 4: Buffer template output, write to response only on success")

	// ─── PATTERN 5: Config file generation (YAML, TOML, Dockerfile, etc.) ───
	const dockerTmpl = `FROM golang:{{ .GoVersion }}-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/{{ .BinaryName }} ./cmd/{{ .BinaryName }}

FROM alpine:3.19
COPY --from=builder /app/{{ .BinaryName }} /usr/local/bin/
EXPOSE {{ .Port }}
CMD ["{{ .BinaryName }}"]
`
	dockerFile := texttemplate.Must(texttemplate.New("docker").Parse(dockerTmpl))
	var dockerBuf bytes.Buffer
	dockerFile.Execute(&dockerBuf, struct {
		GoVersion, BinaryName string
		Port                  int
	}{"1.22", "myservice", 8080})
	fmt.Printf("  Dockerfile generation:\n%s", dockerBuf.String())

	fmt.Println()
}
