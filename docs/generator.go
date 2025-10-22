package docs

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIDocGenerator API 文档生成器
type APIDocGenerator struct {
	config     SwaggerConfig
	routes     []RouteInfo
	models     map[string]ModelInfo
	fileSet    *token.FileSet
	outputPath string
}

// RouteInfo 路由信息
type RouteInfo struct {
	Method      string                  `json:"method"`
	Path        string                  `json:"path"`
	Handler     string                  `json:"handler"`
	Summary     string                  `json:"summary"`
	Description string                  `json:"description"`
	Tags        []string                `json:"tags"`
	Parameters  []ParameterInfo         `json:"parameters"`
	Responses   map[string]ResponseInfo `json:"responses"`
	Security    []SecurityInfo          `json:"security"`
}

// ParameterInfo 参数信息
type ParameterInfo struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, body
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
	Schema      *SchemaInfo `json:"schema,omitempty"`
}

// ResponseInfo 响应信息
type ResponseInfo struct {
	Description string                 `json:"description"`
	Schema      *SchemaInfo            `json:"schema,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// SecurityInfo 安全信息
type SecurityInfo struct {
	Type   string   `json:"type"`
	Name   string   `json:"name"`
	In     string   `json:"in"`
	Scopes []string `json:"scopes,omitempty"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Properties  map[string]PropertyInfo `json:"properties"`
	Required    []string                `json:"required"`
	Description string                  `json:"description"`
}

// PropertyInfo 属性信息
type PropertyInfo struct {
	Type        string      `json:"type"`
	Format      string      `json:"format,omitempty"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Items       *SchemaInfo `json:"items,omitempty"`
}

// SchemaInfo 模式信息
type SchemaInfo struct {
	Type       string                  `json:"type,omitempty"`
	Format     string                  `json:"format,omitempty"`
	Ref        string                  `json:"$ref,omitempty"`
	Items      *SchemaInfo             `json:"items,omitempty"`
	Properties map[string]PropertyInfo `json:"properties,omitempty"`
	Required   []string                `json:"required,omitempty"`
}

// NewAPIDocGenerator 创建 API 文档生成器
func NewAPIDocGenerator(config SwaggerConfig, outputPath string) *APIDocGenerator {
	return &APIDocGenerator{
		config:     config,
		routes:     make([]RouteInfo, 0),
		models:     make(map[string]ModelInfo),
		fileSet:    token.NewFileSet(),
		outputPath: outputPath,
	}
}

// ScanRoutes 扫描路由
func (g *APIDocGenerator) ScanRoutes(engine *gin.Engine) error {
	routes := engine.Routes()

	for _, route := range routes {
		routeInfo := RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		}

		// 解析路由注释
		if err := g.parseRouteComments(&routeInfo); err != nil {
			continue // 跳过解析失败的路由
		}

		g.routes = append(g.routes, routeInfo)
	}

	return nil
}

// ScanModels 扫描模型
func (g *APIDocGenerator) ScanModels(packagePath string) error {
	return filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return g.parseGoFile(path)
	})
}

// GenerateSwaggerJSON 生成 Swagger JSON
func (g *APIDocGenerator) GenerateSwaggerJSON() ([]byte, error) {
	swagger := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       g.config.Title,
			"description": g.config.Description,
			"version":     g.config.Version,
			"contact": map[string]interface{}{
				"name":  g.config.Contact.Name,
				"url":   g.config.Contact.URL,
				"email": g.config.Contact.Email,
			},
			"license": map[string]interface{}{
				"name": g.config.License.Name,
				"url":  g.config.License.URL,
			},
		},
		"host":        g.config.Host,
		"basePath":    g.config.BasePath,
		"schemes":     g.config.Schemes,
		"consumes":    []string{"application/json"},
		"produces":    []string{"application/json"},
		"paths":       g.generatePaths(),
		"definitions": g.generateDefinitions(),
		"securityDefinitions": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "Bearer token authentication. Format: Bearer {token}",
			},
		},
	}

	return json.MarshalIndent(swagger, "", "  ")
}

// SaveSwaggerJSON 保存 Swagger JSON 文件
func (g *APIDocGenerator) SaveSwaggerJSON() error {
	data, err := g.GenerateSwaggerJSON()
	if err != nil {
		return fmt.Errorf("failed to generate swagger JSON: %w", err)
	}

	filename := filepath.Join(g.outputPath, "swagger.json")
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write swagger file: %w", err)
	}

	return nil
}

// GenerateMarkdown 生成 Markdown 文档
func (g *APIDocGenerator) GenerateMarkdown() (string, error) {
	var md strings.Builder

	// 标题和描述
	md.WriteString(fmt.Sprintf("# %s\n\n", g.config.Title))
	md.WriteString(fmt.Sprintf("%s\n\n", g.config.Description))
	md.WriteString(fmt.Sprintf("**版本:** %s\n\n", g.config.Version))
	md.WriteString(fmt.Sprintf("**基础路径:** %s\n\n", g.config.BasePath))

	// 目录
	md.WriteString("## 目录\n\n")
	tags := g.getUniqueTags()
	for _, tag := range tags {
		md.WriteString(fmt.Sprintf("- [%s](#%s)\n", tag, strings.ToLower(strings.ReplaceAll(tag, " ", "-"))))
	}
	md.WriteString("\n")

	// 按标签分组路由
	for _, tag := range tags {
		md.WriteString(fmt.Sprintf("## %s\n\n", tag))

		for _, route := range g.routes {
			if g.hasTag(route.Tags, tag) {
				md.WriteString(g.generateRouteMarkdown(route))
			}
		}
	}

	// 数据模型
	if len(g.models) > 0 {
		md.WriteString("## 数据模型\n\n")
		for _, model := range g.models {
			md.WriteString(g.generateModelMarkdown(model))
		}
	}

	return md.String(), nil
}

// SaveMarkdown 保存 Markdown 文档
func (g *APIDocGenerator) SaveMarkdown() error {
	content, err := g.GenerateMarkdown()
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	filename := filepath.Join(g.outputPath, "api.md")
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// 私有方法

// parseRouteComments 解析路由注释
func (g *APIDocGenerator) parseRouteComments(route *RouteInfo) error {
	// 这里需要实现从源码中解析注释的逻辑
	// 可以通过 AST 解析或者约定的注释格式来实现

	// 示例：设置默认值
	route.Summary = fmt.Sprintf("%s %s", route.Method, route.Path)
	route.Description = route.Summary
	route.Tags = []string{"API"}

	// 解析路径参数
	pathParams := g.extractPathParameters(route.Path)
	for _, param := range pathParams {
		route.Parameters = append(route.Parameters, ParameterInfo{
			Name:        param,
			In:          "path",
			Type:        "string",
			Required:    true,
			Description: fmt.Sprintf("%s parameter", param),
		})
	}

	// 设置默认响应
	route.Responses = map[string]ResponseInfo{
		"200": {
			Description: "Success",
		},
		"400": {
			Description: "Bad Request",
			Schema: &SchemaInfo{
				Ref: "#/definitions/ErrorResponse",
			},
		},
		"500": {
			Description: "Internal Server Error",
			Schema: &SchemaInfo{
				Ref: "#/definitions/ErrorResponse",
			},
		},
	}

	return nil
}

// parseGoFile 解析 Go 文件
func (g *APIDocGenerator) parseGoFile(filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	file, err := parser.ParseFile(g.fileSet, filename, src, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历文件中的类型声明
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						model := g.parseStruct(typeSpec.Name.Name, structType, genDecl.Doc)
						g.models[model.Name] = model
					}
				}
			}
		}
	}

	return nil
}

// parseStruct 解析结构体
func (g *APIDocGenerator) parseStruct(name string, structType *ast.StructType, doc *ast.CommentGroup) ModelInfo {
	model := ModelInfo{
		Name:       name,
		Type:       "object",
		Properties: make(map[string]PropertyInfo),
		Required:   make([]string, 0),
	}

	if doc != nil {
		model.Description = strings.TrimSpace(doc.Text())
	}

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.IsExported() {
				prop := g.parseField(field)
				jsonName := g.getJSONName(field)
				if jsonName != "" && jsonName != "-" {
					model.Properties[jsonName] = prop

					// 检查是否为必需字段
					if g.isRequiredField(field) {
						model.Required = append(model.Required, jsonName)
					}
				}
			}
		}
	}

	return model
}

// parseField 解析字段
func (g *APIDocGenerator) parseField(field *ast.Field) PropertyInfo {
	prop := PropertyInfo{}

	// 解析类型
	typeStr := g.getTypeString(field.Type)
	prop.Type, prop.Format = g.mapGoTypeToSwagger(typeStr)

	// 解析注释
	if field.Doc != nil {
		prop.Description = strings.TrimSpace(field.Doc.Text())
	}

	// 解析标签
	if field.Tag != nil {
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

		// 解析验证标签
		if validate := tag.Get("validate"); validate != "" {
			if strings.Contains(validate, "required") {
				// 标记为必需字段
			}

			// 解析枚举值
			if enumMatch := regexp.MustCompile(`oneof=([^,\s]+)`).FindStringSubmatch(validate); len(enumMatch) > 1 {
				prop.Enum = strings.Split(enumMatch[1], " ")
			}
		}

		// 解析示例值
		if example := tag.Get("example"); example != "" {
			prop.Example = example
		}
	}

	return prop
}

// generatePaths 生成路径信息
func (g *APIDocGenerator) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, route := range g.routes {
		if paths[route.Path] == nil {
			paths[route.Path] = make(map[string]interface{})
		}

		pathItem := paths[route.Path].(map[string]interface{})
		pathItem[strings.ToLower(route.Method)] = g.generateOperation(route)
	}

	return paths
}

// generateOperation 生成操作信息
func (g *APIDocGenerator) generateOperation(route RouteInfo) map[string]interface{} {
	operation := map[string]interface{}{
		"summary":     route.Summary,
		"description": route.Description,
		"tags":        route.Tags,
		"parameters":  g.generateParameters(route.Parameters),
		"responses":   g.generateResponses(route.Responses),
	}

	if len(route.Security) > 0 {
		operation["security"] = g.generateSecurity(route.Security)
	}

	return operation
}

// generateParameters 生成参数信息
func (g *APIDocGenerator) generateParameters(params []ParameterInfo) []map[string]interface{} {
	var parameters []map[string]interface{}

	for _, param := range params {
		p := map[string]interface{}{
			"name":        param.Name,
			"in":          param.In,
			"required":    param.Required,
			"description": param.Description,
		}

		if param.Schema != nil {
			p["schema"] = g.generateSchema(*param.Schema)
		} else {
			p["type"] = param.Type
		}

		if param.Example != nil {
			p["example"] = param.Example
		}

		parameters = append(parameters, p)
	}

	return parameters
}

// generateResponses 生成响应信息
func (g *APIDocGenerator) generateResponses(responses map[string]ResponseInfo) map[string]interface{} {
	result := make(map[string]interface{})

	for code, response := range responses {
		r := map[string]interface{}{
			"description": response.Description,
		}

		if response.Schema != nil {
			r["schema"] = g.generateSchema(*response.Schema)
		}

		if len(response.Examples) > 0 {
			r["examples"] = response.Examples
		}

		result[code] = r
	}

	return result
}

// generateSecurity 生成安全信息
func (g *APIDocGenerator) generateSecurity(securities []SecurityInfo) []map[string][]string {
	var result []map[string][]string

	for _, security := range securities {
		s := map[string][]string{
			security.Name: security.Scopes,
		}
		result = append(result, s)
	}

	return result
}

// generateSchema 生成模式信息
func (g *APIDocGenerator) generateSchema(schema SchemaInfo) map[string]interface{} {
	result := make(map[string]interface{})

	if schema.Ref != "" {
		result["$ref"] = schema.Ref
	} else {
		if schema.Type != "" {
			result["type"] = schema.Type
		}
		if schema.Format != "" {
			result["format"] = schema.Format
		}
		if schema.Items != nil {
			result["items"] = g.generateSchema(*schema.Items)
		}
		if len(schema.Properties) > 0 {
			props := make(map[string]interface{})
			for name, prop := range schema.Properties {
				props[name] = map[string]interface{}{
					"type":        prop.Type,
					"description": prop.Description,
				}
				if prop.Format != "" {
					props[name].(map[string]interface{})["format"] = prop.Format
				}
				if prop.Example != nil {
					props[name].(map[string]interface{})["example"] = prop.Example
				}
			}
			result["properties"] = props
		}
		if len(schema.Required) > 0 {
			result["required"] = schema.Required
		}
	}

	return result
}

// generateDefinitions 生成定义信息
func (g *APIDocGenerator) generateDefinitions() map[string]interface{} {
	definitions := make(map[string]interface{})

	for _, model := range g.models {
		def := map[string]interface{}{
			"type":        model.Type,
			"description": model.Description,
		}

		if len(model.Properties) > 0 {
			props := make(map[string]interface{})
			for name, prop := range model.Properties {
				p := map[string]interface{}{
					"type":        prop.Type,
					"description": prop.Description,
				}
				if prop.Format != "" {
					p["format"] = prop.Format
				}
				if prop.Example != nil {
					p["example"] = prop.Example
				}
				if len(prop.Enum) > 0 {
					p["enum"] = prop.Enum
				}
				props[name] = p
			}
			def["properties"] = props
		}

		if len(model.Required) > 0 {
			def["required"] = model.Required
		}

		definitions[model.Name] = def
	}

	return definitions
}

// 辅助方法

// extractPathParameters 提取路径参数
func (g *APIDocGenerator) extractPathParameters(path string) []string {
	re := regexp.MustCompile(`:([^/]+)`)
	matches := re.FindAllStringSubmatch(path, -1)

	var params []string
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	return params
}

// getJSONName 获取 JSON 字段名
func (g *APIDocGenerator) getJSONName(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}

	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
	jsonTag := tag.Get("json")

	if jsonTag == "" {
		return ""
	}

	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// isRequiredField 检查是否为必需字段
func (g *APIDocGenerator) isRequiredField(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}

	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
	validate := tag.Get("validate")

	return strings.Contains(validate, "required")
}

// getTypeString 获取类型字符串
func (g *APIDocGenerator) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", g.getTypeString(t.X), t.Sel.Name)
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", g.getTypeString(t.Elt))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", g.getTypeString(t.X))
	default:
		return "interface{}"
	}
}

// mapGoTypeToSwagger 映射 Go 类型到 Swagger 类型
func (g *APIDocGenerator) mapGoTypeToSwagger(goType string) (string, string) {
	switch goType {
	case "string":
		return "string", ""
	case "int", "int8", "int16", "int32":
		return "integer", "int32"
	case "int64":
		return "integer", "int64"
	case "uint", "uint8", "uint16", "uint32":
		return "integer", "int32"
	case "uint64":
		return "integer", "int64"
	case "float32":
		return "number", "float"
	case "float64":
		return "number", "double"
	case "bool":
		return "boolean", ""
	case "time.Time":
		return "string", "date-time"
	default:
		if strings.HasPrefix(goType, "[]") {
			return "array", ""
		}
		return "object", ""
	}
}

// getUniqueTags 获取唯一标签
func (g *APIDocGenerator) getUniqueTags() []string {
	tagSet := make(map[string]bool)

	for _, route := range g.routes {
		for _, tag := range route.Tags {
			tagSet[tag] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// hasTag 检查是否包含指定标签
func (g *APIDocGenerator) hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

// generateRouteMarkdown 生成路由 Markdown
func (g *APIDocGenerator) generateRouteMarkdown(route RouteInfo) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("### %s %s\n\n", route.Method, route.Path))
	md.WriteString(fmt.Sprintf("**描述:** %s\n\n", route.Description))

	if len(route.Parameters) > 0 {
		md.WriteString("**参数:**\n\n")
		md.WriteString("| 名称 | 位置 | 类型 | 必需 | 描述 |\n")
		md.WriteString("|------|------|------|------|------|\n")

		for _, param := range route.Parameters {
			required := "否"
			if param.Required {
				required = "是"
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				param.Name, param.In, param.Type, required, param.Description))
		}
		md.WriteString("\n")
	}

	if len(route.Responses) > 0 {
		md.WriteString("**响应:**\n\n")
		md.WriteString("| 状态码 | 描述 |\n")
		md.WriteString("|--------|------|\n")

		for code, response := range route.Responses {
			md.WriteString(fmt.Sprintf("| %s | %s |\n", code, response.Description))
		}
		md.WriteString("\n")
	}

	md.WriteString("---\n\n")

	return md.String()
}

// generateModelMarkdown 生成模型 Markdown
func (g *APIDocGenerator) generateModelMarkdown(model ModelInfo) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("### %s\n\n", model.Name))
	md.WriteString(fmt.Sprintf("**描述:** %s\n\n", model.Description))

	if len(model.Properties) > 0 {
		md.WriteString("**属性:**\n\n")
		md.WriteString("| 名称 | 类型 | 必需 | 描述 |\n")
		md.WriteString("|------|------|------|------|\n")

		for name, prop := range model.Properties {
			required := "否"
			for _, req := range model.Required {
				if req == name {
					required = "是"
					break
				}
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				name, prop.Type, required, prop.Description))
		}
		md.WriteString("\n")
	}

	md.WriteString("---\n\n")

	return md.String()
}
