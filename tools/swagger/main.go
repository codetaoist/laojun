package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SwaggerInfo represents the basic information of the API
type SwaggerInfo struct {
	Title          string            `json:"title" yaml:"title"`
	Description    string            `json:"description" yaml:"description"`
	Version        string            `json:"version" yaml:"version"`
	TermsOfService string            `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact          `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License          `json:"license,omitempty" yaml:"license,omitempty"`
	Extensions     map[string]string `json:"-" yaml:"-"`
}

// Contact represents the contact information
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License represents the license information
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents a server
type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable represents a server variable
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Swagger represents the swagger object
type Swagger struct {
	OpenAPI      string                 `json:"openapi" yaml:"openapi"`
	Info         SwaggerInfo            `json:"info" yaml:"info"`
	Servers      []Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        map[string]PathItem    `json:"paths" yaml:"paths"`
	Components   *Components            `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// PathItem represents a path item
type PathItem struct {
	Ref         string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation  `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation  `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation  `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation represents an operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
	Callbacks   map[string]Callback   `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Servers     []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Parameter represents a parameter
type Parameter struct {
	Name            string             `json:"name" yaml:"name"`
	In              string             `json:"in" yaml:"in"`
	Description     string             `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool               `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         *bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// Response represents a response
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// MediaType represents a media type
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Schema represents a schema
type Schema struct {
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	AllOf                []*Schema              `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*Schema              `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*Schema              `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty" yaml:"not,omitempty"`
	Items                *Schema                `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Title                string                 `json:"title,omitempty" yaml:"title,omitempty"`
	MultipleOf           *float64               `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     *bool                  `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     *bool                  `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            *int64                 `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            *int64                 `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems             *int64                 `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             *int64                 `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        *int64                 `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int64                 `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required             []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Example              interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Discriminator        *Discriminator         `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                   `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	XML                  *XML                   `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Deprecated           bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

// Components represents the components object
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement map[string][]string

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// Tag represents a tag
type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocumentation represents external documentation
type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// Header represents a header
type Header struct {
	Description     string             `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool               `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         *bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Example represents an example
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Link represents a link
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
}

// Callback represents a callback
type Callback map[string]PathItem

// Discriminator represents a discriminator
type Discriminator struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// XML represents XML metadata
type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}

// Encoding represents encoding
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string            `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       *bool             `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Config represents the configuration
type Config struct {
	SourceDir    string                `yaml:"source_dir"`
	OutputDir    string                `yaml:"output_dir"`
	OutputFormat string                `yaml:"output_format"` // json, yaml, html
	IncludeFiles []string              `yaml:"include_files"`
	ExcludeFiles []string              `yaml:"exclude_files"`
	Info         SwaggerInfo           `yaml:"info"`
	Servers      []Server              `yaml:"servers"`
	Security     []SecurityRequirement `yaml:"security"`
}

var (
	config  Config
	swagger Swagger
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "laojun-swagger",
		Short: "Laojun Swagger/OpenAPI Documentation Generator",
		Long: `A comprehensive tool for generating Swagger/OpenAPI documentation from Go source code.
This tool analyzes Go source files and generates API documentation automatically.`,
	}

	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(initCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateCmd() *cobra.Command {
	var configFile, output, format string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Swagger documentation",
		Long:  "Generate Swagger/OpenAPI documentation from Go source code",
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateSwagger(configFile, output, format); err != nil {
				log.Fatalf("Error generating swagger: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "swagger.yaml", "Configuration file path")
	cmd.Flags().StringVarP(&output, "output", "o", "./docs/swagger.json", "Output file path")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, yaml, html)")

	return cmd
}

func generateSwagger(configFile, output, format string) error {
	fmt.Printf("🔧 Loading configuration from: %s\n", configFile)

	if err := loadConfig(configFile); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("📂 Scanning source directory: %s\n", config.SourceDir)

	swagger = Swagger{
		OpenAPI: "3.0.3",
		Info:    config.Info,
		Servers: config.Servers,
		Paths:   make(map[string]PathItem),
		Components: &Components{
			Schemas: make(map[string]*Schema),
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "JWT Bearer token authentication",
				},
			},
		},
		Security: config.Security,
	}

	if err := scanSourceFiles(); err != nil {
		return fmt.Errorf("failed to scan source files: %w", err)
	}

	fmt.Printf("📝 Generating %s documentation...\n", format)

	if err := generateOutput(output, format); err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	fmt.Printf("✅ Documentation generated successfully: %s\n", output)
	return nil
}

func loadConfig(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("⚠️  Configuration file not found, using defaults\n")
		config = getDefaultConfig()
		return nil
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults
	if config.SourceDir == "" {
		config.SourceDir = "./"
	}
	if config.OutputDir == "" {
		config.OutputDir = "./docs"
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "json"
	}
	if len(config.IncludeFiles) == 0 {
		config.IncludeFiles = []string{"**/*.go"}
	}

	return nil
}

func getDefaultConfig() Config {
	return Config{
		SourceDir:    "./",
		OutputDir:    "./docs",
		OutputFormat: "json",
		IncludeFiles: []string{"**/*.go"},
		ExcludeFiles: []string{"*_test.go", "vendor/**", "node_modules/**"},
		Info: SwaggerInfo{
			Title:       "Laojun API",
			Description: "Laojun Platform API Documentation",
			Version:     "1.0.0",
			Contact: &Contact{
				Name:  "Laojun Team",
				Email: "team@laojun.com",
			},
			License: &License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []Server{
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Security: []SecurityRequirement{
			{"bearerAuth": []string{}},
		},
	}
}

func scanSourceFiles() error {
	fset := token.NewFileSet()

	return filepath.Walk(config.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files and vendor
		if strings.Contains(path, "_test.go") || strings.Contains(path, "vendor/") {
			return nil
		}

		// Parse file
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		// Extract API information
		extractAPIInfo(file, string(src))

		return nil
	})
}

func extractAPIInfo(file *ast.File, source string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Doc != nil {
				extractRouteInfo(node, source)
			}
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				extractTypeInfo(node)
			}
		}
		return true
	})
}

func extractRouteInfo(funcDecl *ast.FuncDecl, source string) {
	comments := funcDecl.Doc.Text()

	// Look for swagger annotations
	lines := strings.Split(comments, "\n")
	var operation Operation
	var path, method string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse swagger annotations
		if strings.HasPrefix(line, "@Summary") {
			operation.Summary = strings.TrimSpace(strings.TrimPrefix(line, "@Summary"))
		} else if strings.HasPrefix(line, "@Description") {
			operation.Description = strings.TrimSpace(strings.TrimPrefix(line, "@Description"))
		} else if strings.HasPrefix(line, "@Tags") {
			tags := strings.TrimSpace(strings.TrimPrefix(line, "@Tags"))
			operation.Tags = strings.Split(tags, ",")
			for i, tag := range operation.Tags {
				operation.Tags[i] = strings.TrimSpace(tag)
			}
		} else if strings.HasPrefix(line, "@Router") {
			// Parse @Router /path [method]
			re := regexp.MustCompile(`@Router\s+(\S+)\s+\[(\w+)\]`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				path = matches[1]
				method = strings.ToLower(matches[2])
			}
		} else if strings.HasPrefix(line, "@Param") {
			// Parse @Param name in type required description
			parts := strings.Fields(strings.TrimPrefix(line, "@Param"))
			if len(parts) >= 4 {
				param := Parameter{
					Name:        parts[0],
					In:          parts[1],
					Required:    parts[3] == "true",
					Description: strings.Join(parts[4:], " "),
					Schema: &Schema{
						Type: parts[2],
					},
				}
				operation.Parameters = append(operation.Parameters, param)
			}
		} else if strings.HasPrefix(line, "@Success") {
			// Parse @Success 200 {object} ResponseType "description"
			parts := strings.Fields(strings.TrimPrefix(line, "@Success"))
			if len(parts) >= 2 {
				code := parts[0]
				if operation.Responses == nil {
					operation.Responses = make(map[string]Response)
				}
				operation.Responses[code] = Response{
					Description: "Success",
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type: "object",
							},
						},
					},
				}
			}
		}
	}

	// Add default response if none specified
	if operation.Responses == nil {
		operation.Responses = make(map[string]Response)
		operation.Responses["200"] = Response{
			Description: "Success",
		}
	}

	// Add to swagger paths
	if path != "" && method != "" {
		if swagger.Paths[path].Get == nil && swagger.Paths[path].Post == nil &&
			swagger.Paths[path].Put == nil && swagger.Paths[path].Delete == nil {
			swagger.Paths[path] = PathItem{}
		}

		pathItem := swagger.Paths[path]
		switch method {
		case "get":
			pathItem.Get = &operation
		case "post":
			pathItem.Post = &operation
		case "put":
			pathItem.Put = &operation
		case "delete":
			pathItem.Delete = &operation
		case "patch":
			pathItem.Patch = &operation
		}
		swagger.Paths[path] = pathItem
	}
}

func extractTypeInfo(genDecl *ast.GenDecl) {
	for _, spec := range genDecl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				schema := extractStructSchema(structType)
				swagger.Components.Schemas[typeSpec.Name.Name] = schema
			}
		}
	}
}

func extractStructSchema(structType *ast.StructType) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name
			fieldSchema := extractFieldSchema(field.Type)

			// Extract JSON tag
			if field.Tag != nil {
				tag := field.Tag.Value
				if jsonTag := extractJSONTag(tag); jsonTag != "" {
					fieldName = jsonTag
				}
			}

			schema.Properties[fieldName] = fieldSchema
		}
	}

	return schema
}

func extractFieldSchema(expr ast.Expr) *Schema {
	switch t := expr.(type) {
	case *ast.Ident:
		return &Schema{Type: goTypeToSwaggerType(t.Name)}
	case *ast.ArrayType:
		return &Schema{
			Type:  "array",
			Items: extractFieldSchema(t.Elt),
		}
	case *ast.StarExpr:
		return extractFieldSchema(t.X)
	case *ast.SelectorExpr:
		return &Schema{Type: "object"}
	default:
		return &Schema{Type: "object"}
	}
}

func goTypeToSwaggerType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64":
		return "integer"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "object"
	}
}

func extractJSONTag(tag string) string {
	// Remove quotes
	tag = strings.Trim(tag, "`")

	// Find json tag
	re := regexp.MustCompile(`json:"([^"]*)"`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) > 1 {
		jsonTag := matches[1]
		// Handle json:"-" or json:"name,omitempty"
		if jsonTag == "-" {
			return ""
		}
		parts := strings.Split(jsonTag, ",")
		return parts[0]
	}

	return ""
}

func generateOutput(output, format string) error {
	switch format {
	case "yaml":
		return generateYAML(output)
	case "html":
		return generateHTML(output)
	default:
		return generateJSON(output)
	}
}

func generateJSON(output string) error {
	data, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, 0644)
}

func generateYAML(output string) error {
	data, err := yaml.Marshal(swagger)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, 0644)
}

func generateHTML(output string) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Info.Title}} - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: './swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	t, err := template.New("swagger").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, swagger)
}

func serveCmd() *cobra.Command {
	var port int
	var dir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve Swagger documentation",
		Long:  "Start a web server to serve the Swagger documentation",
		Run: func(cmd *cobra.Command, args []string) {
			serveSwagger(port, dir)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
	cmd.Flags().StringVarP(&dir, "dir", "d", "./docs", "Documentation directory")

	return cmd
}

func serveSwagger(port int, dir string) {
	fmt.Printf("🚀 Starting Swagger documentation server on port %d...\n", port)
	fmt.Printf("📖 Documentation available at: http://localhost:%d\n", port)

	// Serve static files
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	// Serve swagger.json if it exists
	swaggerPath := filepath.Join(dir, "swagger.json")
	if _, err := os.Stat(swaggerPath); err == nil {
		http.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, swaggerPath)
		})
	}

	// Serve swagger UI
	http.HandleFunc("/swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		swaggerUIHTML := generateSwaggerUIHTML()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerUIHTML))
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"service": "laojun-swagger",
		})
	})

	fmt.Printf("🌐 Swagger UI available at: http://localhost:%d/swagger-ui\n", port)
	fmt.Printf("🔍 Health check at: http://localhost:%d/health\n", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func generateSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Laojun API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1f2937;
        }
        .swagger-ui .topbar .download-url-wrapper {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: './swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                filter: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                },
                onFailure: function(error) {
                    console.error('Failed to load Swagger UI:', error);
                }
            });
        };
    </script>
</body>
</html>`
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate Swagger documentation",
		Long:  "Validate the generated Swagger/OpenAPI documentation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateSwagger(args[0]); err != nil {
				log.Fatalf("Validation failed: %v", err)
			}
		},
	}
}

func validateSwagger(file string) error {
	fmt.Printf("🔍 Validating Swagger documentation: %s\n", file)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var swagger Swagger
	if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		err = yaml.Unmarshal(data, &swagger)
	} else {
		err = json.Unmarshal(data, &swagger)
	}

	if err != nil {
		return fmt.Errorf("error parsing file: %w", err)
	}

	// Comprehensive validation
	errors := validateSwaggerSpec(&swagger)

	if len(errors) > 0 {
		fmt.Println("❌ Validation failed:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("validation failed with %d errors", len(errors))
	}

	fmt.Println("✅ Swagger documentation is valid!")
	return nil
}

func validateSwaggerSpec(swagger *Swagger) []string {
	var errors []string

	// Required fields validation
	if swagger.OpenAPI == "" {
		errors = append(errors, "OpenAPI version is required")
	}

	if swagger.Info.Title == "" {
		errors = append(errors, "Info.title is required")
	}

	if swagger.Info.Version == "" {
		errors = append(errors, "Info.version is required")
	}

	if len(swagger.Paths) == 0 {
		errors = append(errors, "At least one path is required")
	}

	// Validate paths
	for path, pathItem := range swagger.Paths {
		if !strings.HasPrefix(path, "/") {
			errors = append(errors, fmt.Sprintf("Path '%s' must start with '/'", path))
		}

		// Validate operations
		operations := []*Operation{
			pathItem.Get, pathItem.Post, pathItem.Put, pathItem.Delete,
			pathItem.Patch, pathItem.Options, pathItem.Head, pathItem.Trace,
		}

		for _, op := range operations {
			if op != nil {
				if len(op.Responses) == 0 {
					errors = append(errors, fmt.Sprintf("Operation in path '%s' must have at least one response", path))
				}
			}
		}
	}

	// Validate servers
	for i, server := range swagger.Servers {
		if server.URL == "" {
			errors = append(errors, fmt.Sprintf("Server[%d].url is required", i))
		}
	}

	return errors
}

func initCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Swagger configuration",
		Long:  "Create a default swagger.yaml configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initSwagger(force); err != nil {
				log.Fatalf("Error initializing swagger: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing configuration")

	return cmd
}

func initSwagger(force bool) error {
	configFile := "swagger.yaml"

	if !force {
		if _, err := os.Stat(configFile); err == nil {
			return fmt.Errorf("configuration file already exists. Use --force to overwrite")
		}
	}

	fmt.Println("🚀 Initializing Swagger configuration...")

	config := getDefaultConfig()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("✅ Configuration file created: %s\n", configFile)
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Edit swagger.yaml to customize your API documentation")
	fmt.Println("   2. Add swagger annotations to your Go code:")
	fmt.Println("      // @Summary Get user by ID")
	fmt.Println("      // @Description Get a user by their unique identifier")
	fmt.Println("      // @Tags users")
	fmt.Println("      // @Param id path int true \"User ID\"")
	fmt.Println("      // @Success 200 {object} User")
	fmt.Println("      // @Router /users/{id} [get]")
	fmt.Println("   3. Run 'laojun-swagger generate' to generate documentation")
	fmt.Println("   4. Run 'laojun-swagger serve' to start the documentation server")

	return nil
}
