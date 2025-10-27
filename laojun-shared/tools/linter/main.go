package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Rule 规范检查规则
type Rule struct {
	Name        string
	Description string
	Check       func(*ast.File, *token.FileSet) []Issue
}

// Issue 检查发现的问题
type Issue struct {
	File        string
	Line        int
	Column      int
	Rule        string
	Message     string
	Severity    string // error, warning, info
}

// Linter API规范检查器
type Linter struct {
	rules []Rule
}

// NewLinter 创建新的检查器
func NewLinter() *Linter {
	return &Linter{
		rules: []Rule{
			{
				Name:        "interface-naming",
				Description: "接口命名应该使用名词或形容词",
				Check:       checkInterfaceNaming,
			},
			{
				Name:        "config-struct",
				Description: "每个包应该有Config结构体",
				Check:       checkConfigStruct,
			},
			{
				Name:        "error-handling",
				Description: "错误应该定义为包级变量",
				Check:       checkErrorHandling,
			},
			{
				Name:        "context-usage",
				Description: "公共方法应该接受context.Context参数",
				Check:       checkContextUsage,
			},
			{
				Name:        "validation-method",
				Description: "Config结构体应该有Validate方法",
				Check:       checkValidationMethod,
			},
		},
	}
}

// CheckDirectory 检查目录
func (l *Linter) CheckDirectory(dir string) ([]Issue, error) {
	var allIssues []Issue

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		issues, err := l.CheckFile(path)
		if err != nil {
			return err
		}

		allIssues = append(allIssues, issues...)
		return nil
	})

	return allIssues, err
}

// CheckFile 检查单个文件
func (l *Linter) CheckFile(filename string) ([]Issue, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	for _, rule := range l.rules {
		ruleIssues := rule.Check(node, fset)
		for i := range ruleIssues {
			ruleIssues[i].File = filename
			ruleIssues[i].Rule = rule.Name
		}
		issues = append(issues, ruleIssues...)
	}

	return issues, nil
}

// 检查接口命名
func checkInterfaceNaming(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				name := typeSpec.Name.Name
				if strings.HasSuffix(name, "er") && len(name) > 2 {
					// 允许以er结尾的接口名（如Reader, Writer）
					return true
				}
				if !isValidInterfaceName(name) {
					pos := fset.Position(typeSpec.Pos())
					issues = append(issues, Issue{
						Line:     pos.Line,
						Column:   pos.Column,
						Message:  fmt.Sprintf("接口名 '%s' 应该使用名词或形容词", name),
						Severity: "warning",
					})
				}
			}
		}
		return true
	})

	return issues
}

// 检查Config结构体
func checkConfigStruct(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue
	hasConfig := false

	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				if typeSpec.Name.Name == "Config" {
					hasConfig = true
				}
			}
		}
		return true
	})

	if !hasConfig && file.Name.Name != "main" {
		issues = append(issues, Issue{
			Line:     1,
			Column:   1,
			Message:  "包应该定义Config结构体",
			Severity: "info",
		})
	}

	return issues
}

// 检查错误处理
func checkErrorHandling(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if strings.HasPrefix(name.Name, "Err") {
							// 检查是否使用errors.New或fmt.Errorf
							if len(valueSpec.Values) > 0 {
								if callExpr, ok := valueSpec.Values[0].(*ast.CallExpr); ok {
									if isErrorsNew(callExpr) || isFmtErrorf(callExpr) {
										continue
									}
								}
							}
							pos := fset.Position(name.Pos())
							issues = append(issues, Issue{
								Line:     pos.Line,
								Column:   pos.Column,
								Message:  fmt.Sprintf("错误变量 '%s' 应该使用errors.New()或fmt.Errorf()定义", name.Name),
								Severity: "warning",
							})
						}
					}
				}
			}
		}
		return true
	})

	return issues
}

// 检查Context使用
func checkContextUsage(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			// 检查公共方法（首字母大写）
			if funcDecl.Name.IsExported() && funcDecl.Recv != nil {
				// 检查是否有context.Context参数
				hasContext := false
				if funcDecl.Type.Params != nil {
					for _, field := range funcDecl.Type.Params.List {
						if isContextType(field.Type) {
							hasContext = true
							break
						}
					}
				}

				if !hasContext {
					pos := fset.Position(funcDecl.Pos())
					issues = append(issues, Issue{
						Line:     pos.Line,
						Column:   pos.Column,
						Message:  fmt.Sprintf("公共方法 '%s' 应该接受context.Context参数", funcDecl.Name.Name),
						Severity: "info",
					})
				}
			}
		}
		return true
	})

	return issues
}

// 检查Validate方法
func checkValidationMethod(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue
	hasConfig := false
	hasValidate := false

	// 检查是否有Config结构体
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				if typeSpec.Name.Name == "Config" {
					hasConfig = true
				}
			}
		}
		return true
	})

	// 检查是否有Validate方法
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "Validate" && funcDecl.Recv != nil {
				// 检查接收者是否是Config
				if len(funcDecl.Recv.List) > 0 {
					if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == "Config" {
							hasValidate = true
						}
					}
					if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok && ident.Name == "Config" {
						hasValidate = true
					}
				}
			}
		}
		return true
	})

	if hasConfig && !hasValidate {
		issues = append(issues, Issue{
			Line:     1,
			Column:   1,
			Message:  "Config结构体应该有Validate()方法",
			Severity: "warning",
		})
	}

	return issues
}

// 辅助函数
func isValidInterfaceName(name string) bool {
	// 简单的接口命名检查
	return len(name) > 0 && strings.ToUpper(name[:1]) == name[:1]
}

func isErrorsNew(expr *ast.CallExpr) bool {
	if selExpr, ok := expr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			return ident.Name == "errors" && selExpr.Sel.Name == "New"
		}
	}
	return false
}

func isFmtErrorf(expr *ast.CallExpr) bool {
	if selExpr, ok := expr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			return ident.Name == "fmt" && selExpr.Sel.Name == "Errorf"
		}
	}
	return false
}

func isContextType(expr ast.Expr) bool {
	if selExpr, ok := expr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			return ident.Name == "context" && selExpr.Sel.Name == "Context"
		}
	}
	return false
}

func main() {
	var (
		dir    = flag.String("dir", ".", "Directory to check")
		format = flag.String("format", "text", "Output format: text, json")
	)
	flag.Parse()

	linter := NewLinter()
	issues, err := linter.CheckDirectory(*dir)
	if err != nil {
		log.Fatalf("检查失败: %v", err)
	}

	if len(issues) == 0 {
		fmt.Println("✅ 没有发现API规范问题")
		return
	}

	switch *format {
	case "json":
		// TODO: 实现JSON输出
		fmt.Println("JSON格式输出暂未实现")
	default:
		printTextReport(issues)
	}

	// 统计
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}

	fmt.Printf("\n📊 检查结果: %d 错误, %d 警告, %d 信息\n", errorCount, warningCount, infoCount)
}

func printTextReport(issues []Issue) {
	fmt.Println("🔍 API规范检查报告:")
	fmt.Println(strings.Repeat("=", 50))

	for _, issue := range issues {
		severity := ""
		switch issue.Severity {
		case "error":
			severity = "❌"
		case "warning":
			severity = "⚠️"
		case "info":
			severity = "ℹ️"
		}

		fmt.Printf("%s %s:%d:%d [%s] %s\n",
			severity,
			filepath.Base(issue.File),
			issue.Line,
			issue.Column,
			issue.Rule,
			issue.Message,
		)
	}
}