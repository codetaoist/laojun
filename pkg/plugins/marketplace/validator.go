package marketplace

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ValidationConfig 验证配置
type ValidationConfig struct {
	EnableSecurityScan    bool     `json:"enable_security_scan"`
	EnableSizeCheck       bool     `json:"enable_size_check"`
	EnableFormatCheck     bool     `json:"enable_format_check"`
	EnableDependencyCheck bool     `json:"enable_dependency_check"`
	MaxFileSize           int64    `json:"max_file_size"`
	AllowedExtensions     []string `json:"allowed_extensions"`
	BlockedPatterns       []string `json:"blocked_patterns"`
	RequiredFiles         []string `json:"required_files"`
	TrustedAuthors        []string `json:"trusted_authors"`
	ScanTimeout           int      `json:"scan_timeout"` // seconds
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid        bool                   `json:"valid"`
	Score        int                    `json:"score"` // 0-100
	Issues       []ValidationIssue      `json:"issues"`
	Warnings     []ValidationWarning    `json:"warnings"`
	Metadata     map[string]interface{} `json:"metadata"`
	ScanDuration time.Duration          `json:"scan_duration"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ValidationIssue 验证问题
type ValidationIssue struct {
	Type       string `json:"type"`     // security, format, dependency, size
	Severity   string `json:"severity"` // critical, high, medium, low
	Message    string `json:"message"`
	File       string `json:"file,omitempty"`
	Line       int    `json:"line,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
}

// SecurityRule 安全规则
type SecurityRule struct {
	Name      string   `json:"name"`
	Pattern   string   `json:"pattern"`
	Severity  string   `json:"severity"`
	Message   string   `json:"message"`
	FileTypes []string `json:"file_types"`
	Enabled   bool     `json:"enabled"`
}

// Validator 插件验证器
type Validator struct {
	config        ValidationConfig
	securityRules []SecurityRule
}

// NewValidator 创建新的验证器
func NewValidator(config ValidationConfig) *Validator {
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	if config.ScanTimeout <= 0 {
		config.ScanTimeout = 300 // 5 minutes
	}

	validator := &Validator{
		config: config,
	}

	// 初始化默认安全规则
	validator.initDefaultSecurityRules()

	return validator
}

// ValidatePlugin 验证插件
func (v *Validator) ValidatePlugin(pluginPath string, manifest *PluginEntry) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:     true,
		Score:     100,
		Issues:    []ValidationIssue{},
		Warnings:  []ValidationWarning{},
		Metadata:  make(map[string]interface{}),
		Timestamp: startTime,
	}

	// 基本文件检查
	if err := v.validateBasicFile(pluginPath, result); err != nil {
		return result, err
	}

	// 大小检查
	if v.config.EnableSizeCheck {
		v.validateSize(pluginPath, result)
	}

	// 格式检查
	if v.config.EnableFormatCheck {
		v.validateFormat(pluginPath, result)
	}

	// 安全扫描
	if v.config.EnableSecurityScan {
		if err := v.performSecurityScan(pluginPath, result); err != nil {
			v.addIssue(result, "security", "high", fmt.Sprintf("Security scan failed: %v", err), "", 0, "")
		}
	}

	// 清单验证
	if manifest != nil {
		v.validateManifest(manifest, result)
	}

	// 依赖检查
	if v.config.EnableDependencyCheck && manifest != nil {
		v.validateDependencies(manifest, result)
	}

	// 计算最终分数和有效性
	v.calculateFinalScore(result)

	result.ScanDuration = time.Since(startTime)
	return result, nil
}

// ValidateManifest 验证清单文件
func (v *Validator) ValidateManifest(manifest *PluginEntry) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		Score:     100,
		Issues:    []ValidationIssue{},
		Warnings:  []ValidationWarning{},
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	v.validateManifest(manifest, result)
	v.calculateFinalScore(result)

	return result, nil
}

// ValidateArchive 验证压缩包
func (v *Validator) ValidateArchive(archivePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		Score:     100,
		Issues:    []ValidationIssue{},
		Warnings:  []ValidationWarning{},
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// 检查是否为ZIP文件
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		v.addIssue(result, "format", "critical", "Invalid ZIP archive", archivePath, 0, "Ensure the file is a valid ZIP archive")
		return result, nil
	}
	defer reader.Close()

	// 验证压缩包内文件
	v.validateArchiveContents(&reader.Reader, result)
	v.calculateFinalScore(result)

	return result, nil
}

// GetSecurityRules 获取安全规则
func (v *Validator) GetSecurityRules() []SecurityRule {
	return v.securityRules
}

// AddSecurityRule 添加安全规则
func (v *Validator) AddSecurityRule(rule SecurityRule) {
	v.securityRules = append(v.securityRules, rule)
}

// RemoveSecurityRule 移除安全规则
func (v *Validator) RemoveSecurityRule(name string) {
	for i, rule := range v.securityRules {
		if rule.Name == name {
			v.securityRules = append(v.securityRules[:i], v.securityRules[i+1:]...)
			break
		}
	}
}

// UpdateSecurityRule 更新安全规则
func (v *Validator) UpdateSecurityRule(name string, newRule SecurityRule) {
	for i, rule := range v.securityRules {
		if rule.Name == name {
			v.securityRules[i] = newRule
			break
		}
	}
}

// 私有方法

// validateBasicFile 基本文件验证
func (v *Validator) validateBasicFile(pluginPath string, result *ValidationResult) error {
	// 检查文件是否存在
	info, err := os.Stat(pluginPath)
	if err != nil {
		v.addIssue(result, "format", "critical", "File not found", pluginPath, 0, "Ensure the plugin file exists")
		return nil
	}

	// 记录文件信息
	result.Metadata["file_size"] = info.Size()
	result.Metadata["file_mode"] = info.Mode().String()
	result.Metadata["modified_time"] = info.ModTime()

	return nil
}

// validateSize 验证文件大小
func (v *Validator) validateSize(pluginPath string, result *ValidationResult) {
	info, err := os.Stat(pluginPath)
	if err != nil {
		return
	}

	if info.Size() > v.config.MaxFileSize {
		v.addIssue(result, "size", "high",
			fmt.Sprintf("File size (%d bytes) exceeds maximum allowed size (%d bytes)",
				info.Size(), v.config.MaxFileSize),
			pluginPath, 0,
			fmt.Sprintf("Reduce file size to under %d bytes", v.config.MaxFileSize))
	}

	// 检查是否过小
	if info.Size() < 1024 {
		v.addWarning(result, "size", "File is very small, may be incomplete", pluginPath)
	}
}

// validateFormat 验证文件格式
func (v *Validator) validateFormat(pluginPath string, result *ValidationResult) {
	ext := strings.ToLower(filepath.Ext(pluginPath))

	// 检查扩展名
	if len(v.config.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range v.config.AllowedExtensions {
			if ext == allowedExt {
				allowed = true
				break
			}
		}

		if !allowed {
			v.addIssue(result, "format", "medium",
				fmt.Sprintf("File extension '%s' is not allowed", ext),
				pluginPath, 0,
				fmt.Sprintf("Use one of the allowed extensions: %v", v.config.AllowedExtensions))
		}
	}

	// 验证ZIP格式
	if ext == ".zip" {
		if _, err := zip.OpenReader(pluginPath); err != nil {
			v.addIssue(result, "format", "high", "Invalid ZIP format", pluginPath, 0, "Ensure the file is a valid ZIP archive")
		}
	}
}

// performSecurityScan 执行安全扫描
func (v *Validator) performSecurityScan(pluginPath string, result *ValidationResult) error {
	// 如果是ZIP文件，扫描内部文件
	if strings.ToLower(filepath.Ext(pluginPath)) == ".zip" {
		return v.scanZipFile(pluginPath, result)
	}

	// 扫描单个文件
	return v.scanFile(pluginPath, result)
}

// scanZipFile 扫描ZIP文件
func (v *Validator) scanZipFile(zipPath string, result *ValidationResult) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// 检查文件路径是否可疑
		if v.isPathSuspicious(file.Name) {
			v.addIssue(result, "security", "high",
				fmt.Sprintf("Suspicious file path: %s", file.Name),
				file.Name, 0, "Remove or rename suspicious files")
		}

		// 扫描文件内容
		if !file.FileInfo().IsDir() {
			rc, err := file.Open()
			if err != nil {
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()

			if err != nil {
				continue
			}

			v.scanContent(string(content), file.Name, result)
		}
	}

	return nil
}

// scanFile 扫描单个文件
func (v *Validator) scanFile(filePath string, result *ValidationResult) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	v.scanContent(string(content), filePath, result)
	return nil
}

// scanContent 扫描文件内容
func (v *Validator) scanContent(content, filePath string, result *ValidationResult) {
	lines := strings.Split(content, "\n")

	for _, rule := range v.securityRules {
		if !rule.Enabled {
			continue
		}

		// 检查文件类型匹配
		if len(rule.FileTypes) > 0 {
			ext := strings.ToLower(filepath.Ext(filePath))
			matched := false
			for _, fileType := range rule.FileTypes {
				if ext == fileType {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 编译正则表达式
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		// 扫描每一行
		for lineNum, line := range lines {
			if regex.MatchString(line) {
				v.addIssue(result, "security", rule.Severity, rule.Message, filePath, lineNum+1, "")
			}
		}
	}

	// 检查阻止的模式
	for _, pattern := range v.config.BlockedPatterns {
		if strings.Contains(content, pattern) {
			v.addIssue(result, "security", "high",
				fmt.Sprintf("Blocked pattern found: %s", pattern),
				filePath, 0, "Remove blocked content")
		}
	}
}

// validateManifest 验证清单
func (v *Validator) validateManifest(manifest *PluginEntry, result *ValidationResult) {
	// 必需字段检查
	if manifest.ID == "" {
		v.addIssue(result, "format", "critical", "Plugin ID is required", "", 0, "Add a unique plugin ID")
	}

	if manifest.Name == "" {
		v.addIssue(result, "format", "critical", "Plugin name is required", "", 0, "Add a plugin name")
	}

	if manifest.Version == "" {
		v.addIssue(result, "format", "critical", "Plugin version is required", "", 0, "Add a version number")
	}

	if manifest.Author == "" {
		v.addIssue(result, "format", "medium", "Plugin author is recommended", "", 0, "Add author information")
	}

	// 版本格式检查
	if manifest.Version != "" && !v.isValidVersion(manifest.Version) {
		v.addIssue(result, "format", "medium", "Invalid version format", "", 0, "Use semantic versioning (e.g., 1.0.0)")
	}

	// 检查可信作者
	if len(v.config.TrustedAuthors) > 0 {
		trusted := false
		for _, author := range v.config.TrustedAuthors {
			if manifest.Author == author {
				trusted = true
				break
			}
		}
		if !trusted {
			v.addWarning(result, "security", "Author is not in trusted list", "")
		}
	}

	// 检查描述长度
	if len(manifest.Description) > 1000 {
		v.addWarning(result, "format", "Description is very long", "")
	}

	// 检查标签数量
	if len(manifest.Tags) > 10 {
		v.addWarning(result, "format", "Too many tags", "")
	}
}

// validateDependencies 验证依赖
func (v *Validator) validateDependencies(manifest *PluginEntry, result *ValidationResult) {
	for _, dep := range manifest.Dependencies {
		if dep.ID == "" {
			v.addIssue(result, "dependency", "medium", "Dependency ID is required", "", 0, "Add dependency ID")
		}

		if dep.Version != "" && !v.isValidVersion(dep.Version) {
			v.addIssue(result, "dependency", "low",
				fmt.Sprintf("Invalid dependency version format: %s", dep.Version),
				"", 0, "Use semantic versioning")
		}

		// 检查循环依赖（简单检查）
		if dep.ID == manifest.ID {
			v.addIssue(result, "dependency", "high", "Plugin cannot depend on itself", "", 0, "Remove self-dependency")
		}
	}
}

// validateArchiveContents 验证压缩包内文件
func (v *Validator) validateArchiveContents(reader *zip.Reader, result *ValidationResult) {
	var hasManifest bool
	var hasMainFile bool

	for _, file := range reader.File {
		fileName := filepath.Base(file.Name)

		// 检查必需文件
		if fileName == "manifest.json" {
			hasManifest = true
		}

		for _, required := range v.config.RequiredFiles {
			if fileName == required {
				hasMainFile = true
				break
			}
		}

		// 检查文件大小
		if file.UncompressedSize64 > uint64(v.config.MaxFileSize) {
			v.addIssue(result, "size", "high",
				fmt.Sprintf("File %s is too large", file.Name),
				file.Name, 0, "Reduce file size")
		}

		// 检查路径遍历
		if strings.Contains(file.Name, "..") {
			v.addIssue(result, "security", "critical",
				fmt.Sprintf("Path traversal detected: %s", file.Name),
				file.Name, 0, "Remove path traversal attempts")
		}
	}

	if !hasManifest {
		v.addIssue(result, "format", "critical", "Missing manifest.json", "", 0, "Add manifest.json file")
	}

	if len(v.config.RequiredFiles) > 0 && !hasMainFile {
		v.addIssue(result, "format", "high",
			fmt.Sprintf("Missing required files: %v", v.config.RequiredFiles),
			"", 0, "Add required files")
	}
}

// addIssue 添加问题
func (v *Validator) addIssue(result *ValidationResult, issueType, severity, message, file string, line int, suggestion string) {
	issue := ValidationIssue{
		Type:       issueType,
		Severity:   severity,
		Message:    message,
		File:       file,
		Line:       line,
		Suggestion: suggestion,
	}

	result.Issues = append(result.Issues, issue)
}

// addWarning 添加警告
func (v *Validator) addWarning(result *ValidationResult, warningType, message, file string) {
	warning := ValidationWarning{
		Type:    warningType,
		Message: message,
		File:    file,
	}

	result.Warnings = append(result.Warnings, warning)
}

// calculateFinalScore 计算最终分数
func (v *Validator) calculateFinalScore(result *ValidationResult) {
	score := 100

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			score -= 25
			result.Valid = false
		case "high":
			score -= 15
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	for range result.Warnings {
		score -= 2
	}

	if score < 0 {
		score = 0
	}

	result.Score = score

	// 如果分数太低，标记为无效
	if score < 50 {
		result.Valid = false
	}
}

// isValidVersion 检查版本格式
func (v *Validator) isValidVersion(version string) bool {
	// 简单的语义版本检查
	pattern := `^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`
	matched, _ := regexp.MatchString(pattern, version)
	return matched
}

// isPathSuspicious 检查路径是否可疑
func (v *Validator) isPathSuspicious(path string) bool {
	suspicious := []string{
		"../",
		"..\\",
		"/etc/",
		"/bin/",
		"/usr/",
		"C:\\Windows\\",
		"C:\\Program Files\\",
	}

	for _, pattern := range suspicious {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// initDefaultSecurityRules 初始化默认安全规则
func (v *Validator) initDefaultSecurityRules() {
	v.securityRules = []SecurityRule{
		{
			Name:      "SQL Injection",
			Pattern:   `(?i)(union\s+select|drop\s+table|delete\s+from|insert\s+into)`,
			Severity:  "high",
			Message:   "Potential SQL injection pattern detected",
			FileTypes: []string{".js", ".py", ".php", ".go"},
			Enabled:   true,
		},
		{
			Name:      "Command Injection",
			Pattern:   `(?i)(exec\s*\(|system\s*\(|shell_exec|passthru)`,
			Severity:  "high",
			Message:   "Potential command injection pattern detected",
			FileTypes: []string{".js", ".py", ".php", ".go"},
			Enabled:   true,
		},
		{
			Name:      "File System Access",
			Pattern:   `(?i)(file_get_contents|fopen|readfile|include|require)`,
			Severity:  "medium",
			Message:   "File system access detected",
			FileTypes: []string{".php", ".py"},
			Enabled:   true,
		},
		{
			Name:      "Network Access",
			Pattern:   `(?i)(curl|wget|http_request|socket|tcp)`,
			Severity:  "medium",
			Message:   "Network access detected",
			FileTypes: []string{".js", ".py", ".php", ".go"},
			Enabled:   true,
		},
		{
			Name:      "Eval Usage",
			Pattern:   `(?i)(eval\s*\(|Function\s*\(|setTimeout\s*\(.*string)`,
			Severity:  "high",
			Message:   "Dynamic code execution detected",
			FileTypes: []string{".js", ".py", ".php"},
			Enabled:   true,
		},
		{
			Name:      "Hardcoded Secrets",
			Pattern:   `(?i)(password\s*=\s*['"]\w+['"]|api_key\s*=\s*['"]\w+['"]|secret\s*=\s*['"]\w+['"])`,
			Severity:  "critical",
			Message:   "Hardcoded secrets detected",
			FileTypes: []string{".js", ".py", ".php", ".go", ".json", ".yaml", ".yml"},
			Enabled:   true,
		},
	}
}

// CalculateChecksum 计算文件校验和
func (v *Validator) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ValidateChecksum 验证校验和
func (v *Validator) ValidateChecksum(filePath, expectedChecksum string) error {
	actualChecksum, err := v.CalculateChecksum(filePath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}
