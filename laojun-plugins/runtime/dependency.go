package runtime

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// DependencyManager 依赖管理器接口
type DependencyManager interface {
	// RegisterDependency 注册依赖
	RegisterDependency(dep *Dependency) error

	// UnregisterDependency 注销依赖
	UnregisterDependency(name, version string) error

	// ResolveDependencies 解析依赖
	ResolveDependencies(requirements []*DependencyRequirement) ([]*Dependency, error)

	// CheckDependencies 检查依赖
	CheckDependencies(pluginID string, requirements []*DependencyRequirement) error

	// GetDependency 获取依赖
	GetDependency(name, version string) (*Dependency, error)

	// ListDependencies 列出所有依赖
	ListDependencies() []*Dependency

	// GetDependencyGraph 获取依赖图
	GetDependencyGraph(pluginID string) (*DependencyGraph, error)

	// ValidateCircularDependency 验证循环依赖
	ValidateCircularDependency(requirements []*DependencyRequirement) error
}

// Dependency 依赖定义
type Dependency struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        DependencyType    `json:"type"`
	Provider    string            `json:"provider"`
	Description string            `json:"description"`
	Repository  string            `json:"repository,omitempty"`
	License     string            `json:"license,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// DependencyRequirement 依赖需求
type DependencyRequirement struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`      // 支持语义化版本，如 ">=1.0.0", "~1.2.0", "^1.0.0"
	Type         DependencyType `json:"type"`
	Optional     bool           `json:"optional"`
	Scope        DependencyScope `json:"scope"`
	Constraints  []string       `json:"constraints,omitempty"`
}

// DependencyType 依赖类型
type DependencyType string

const (
	DependencyTypePlugin    DependencyType = "plugin"
	DependencyTypeLibrary   DependencyType = "library"
	DependencyTypeService   DependencyType = "service"
	DependencyTypeResource  DependencyType = "resource"
	DependencyTypeRuntime   DependencyType = "runtime"
)

// DependencyScope 依赖范围
type DependencyScope string

const (
	DependencyScopeRuntime     DependencyScope = "runtime"
	DependencyScopeCompile     DependencyScope = "compile"
	DependencyScopeTest        DependencyScope = "test"
	DependencyScopeDevelopment DependencyScope = "development"
)

// DependencyGraph 依赖图
type DependencyGraph struct {
	PluginID     string                    `json:"plugin_id"`
	Dependencies map[string]*DependencyNode `json:"dependencies"`
	Resolved     bool                      `json:"resolved"`
	Conflicts    []*DependencyConflict     `json:"conflicts,omitempty"`
}

// DependencyNode 依赖节点
type DependencyNode struct {
	Dependency   *Dependency             `json:"dependency"`
	Children     []*DependencyNode       `json:"children,omitempty"`
	Parent       *DependencyNode         `json:"-"`
	Depth        int                     `json:"depth"`
	Required     bool                    `json:"required"`
	Resolved     bool                    `json:"resolved"`
}

// DependencyConflict 依赖冲突
type DependencyConflict struct {
	Name         string   `json:"name"`
	Versions     []string `json:"versions"`
	RequiredBy   []string `json:"required_by"`
	ConflictType string   `json:"conflict_type"` // version, type, constraint
	Severity     string   `json:"severity"`      // low, medium, high, critical
}

// DefaultDependencyManager 默认依赖管理器实现
type DefaultDependencyManager struct {
	dependencies map[string]map[string]*Dependency // name -> version -> dependency
	logger       *logrus.Logger
	mu           sync.RWMutex
}

// NewDefaultDependencyManager 创建默认依赖管理器
func NewDefaultDependencyManager(logger *logrus.Logger) *DefaultDependencyManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &DefaultDependencyManager{
		dependencies: make(map[string]map[string]*Dependency),
		logger:       logger,
	}
}

// RegisterDependency 注册依赖
func (dm *DefaultDependencyManager) RegisterDependency(dep *Dependency) error {
	if dep == nil {
		return fmt.Errorf("dependency cannot be nil")
	}

	if dep.Name == "" {
		return fmt.Errorf("dependency name cannot be empty")
	}

	if dep.Version == "" {
		return fmt.Errorf("dependency version cannot be empty")
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.dependencies[dep.Name]; !exists {
		dm.dependencies[dep.Name] = make(map[string]*Dependency)
	}

	dm.dependencies[dep.Name][dep.Version] = dep
	dm.logger.Infof("Registered dependency: %s@%s", dep.Name, dep.Version)
	return nil
}

// UnregisterDependency 注销依赖
func (dm *DefaultDependencyManager) UnregisterDependency(name, version string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	versions, exists := dm.dependencies[name]
	if !exists {
		return fmt.Errorf("dependency %s not found", name)
	}

	if _, exists := versions[version]; !exists {
		return fmt.Errorf("dependency %s@%s not found", name, version)
	}

	delete(versions, version)
	if len(versions) == 0 {
		delete(dm.dependencies, name)
	}

	dm.logger.Infof("Unregistered dependency: %s@%s", name, version)
	return nil
}

// ResolveDependencies 解析依赖
func (dm *DefaultDependencyManager) ResolveDependencies(requirements []*DependencyRequirement) ([]*Dependency, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var resolved []*Dependency
	var conflicts []*DependencyConflict

	for _, req := range requirements {
		dep, err := dm.findBestMatch(req)
		if err != nil {
			if req.Optional {
				dm.logger.Warnf("Optional dependency not found: %s@%s", req.Name, req.Version)
				continue
			}
			return nil, fmt.Errorf("failed to resolve dependency %s@%s: %w", req.Name, req.Version, err)
		}

		// 检查冲突
		if conflict := dm.checkConflict(dep, resolved); conflict != nil {
			conflicts = append(conflicts, conflict)
		}

		resolved = append(resolved, dep)
	}

	if len(conflicts) > 0 {
		return nil, fmt.Errorf("dependency conflicts detected: %v", conflicts)
	}

	dm.logger.Infof("Resolved %d dependencies", len(resolved))
	return resolved, nil
}

// CheckDependencies 检查依赖
func (dm *DefaultDependencyManager) CheckDependencies(pluginID string, requirements []*DependencyRequirement) error {
	_, err := dm.ResolveDependencies(requirements)
	if err != nil {
		return fmt.Errorf("dependency check failed for plugin %s: %w", pluginID, err)
	}
	return nil
}

// GetDependency 获取依赖
func (dm *DefaultDependencyManager) GetDependency(name, version string) (*Dependency, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	versions, exists := dm.dependencies[name]
	if !exists {
		return nil, fmt.Errorf("dependency %s not found", name)
	}

	dep, exists := versions[version]
	if !exists {
		return nil, fmt.Errorf("dependency %s@%s not found", name, version)
	}

	return dep, nil
}

// ListDependencies 列出所有依赖
func (dm *DefaultDependencyManager) ListDependencies() []*Dependency {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var deps []*Dependency
	for _, versions := range dm.dependencies {
		for _, dep := range versions {
			deps = append(deps, dep)
		}
	}

	// 按名称和版本排序
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Name != deps[j].Name {
			return deps[i].Name < deps[j].Name
		}
		return dm.compareVersions(deps[i].Version, deps[j].Version) < 0
	})

	return deps
}

// GetDependencyGraph 获取依赖图
func (dm *DefaultDependencyManager) GetDependencyGraph(pluginID string) (*DependencyGraph, error) {
	// 这里应该从插件注册表获取插件的依赖需求
	// 为了演示，我们返回一个空的依赖图
	return &DependencyGraph{
		PluginID:     pluginID,
		Dependencies: make(map[string]*DependencyNode),
		Resolved:     true,
		Conflicts:    nil,
	}, nil
}

// ValidateCircularDependency 验证循环依赖
func (dm *DefaultDependencyManager) ValidateCircularDependency(requirements []*DependencyRequirement) error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for _, req := range requirements {
		if dm.hasCircularDependency(req.Name, visited, recursionStack, requirements) {
			return fmt.Errorf("circular dependency detected involving %s", req.Name)
		}
	}

	return nil
}

// findBestMatch 查找最佳匹配的依赖
func (dm *DefaultDependencyManager) findBestMatch(req *DependencyRequirement) (*Dependency, error) {
	versions, exists := dm.dependencies[req.Name]
	if !exists {
		return nil, fmt.Errorf("dependency %s not found", req.Name)
	}

	// 如果指定了确切版本
	if !strings.ContainsAny(req.Version, ">=<~^*") {
		if dep, exists := versions[req.Version]; exists {
			return dep, nil
		}
		return nil, fmt.Errorf("exact version %s not found for %s", req.Version, req.Name)
	}

	// 查找满足版本约束的最佳版本
	var candidates []*Dependency
	for version, dep := range versions {
		if dm.satisfiesVersionConstraint(version, req.Version) {
			candidates = append(candidates, dep)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no version satisfies constraint %s for %s", req.Version, req.Name)
	}

	// 返回最新版本
	sort.Slice(candidates, func(i, j int) bool {
		return dm.compareVersions(candidates[i].Version, candidates[j].Version) > 0
	})

	return candidates[0], nil
}

// checkConflict 检查依赖冲突
func (dm *DefaultDependencyManager) checkConflict(dep *Dependency, resolved []*Dependency) *DependencyConflict {
	for _, existing := range resolved {
		if existing.Name == dep.Name && existing.Version != dep.Version {
			return &DependencyConflict{
				Name:         dep.Name,
				Versions:     []string{existing.Version, dep.Version},
				RequiredBy:   []string{"plugin1", "plugin2"}, // 这里应该记录实际的依赖者
				ConflictType: "version",
				Severity:     "medium",
			}
		}
	}
	return nil
}

// satisfiesVersionConstraint 检查版本是否满足约束
func (dm *DefaultDependencyManager) satisfiesVersionConstraint(version, constraint string) bool {
	// 简化的版本约束检查，实际应该使用更完善的语义化版本库
	if constraint == "*" || constraint == "" {
		return true
	}

	if strings.HasPrefix(constraint, ">=") {
		targetVersion := strings.TrimPrefix(constraint, ">=")
		return dm.compareVersions(version, targetVersion) >= 0
	}

	if strings.HasPrefix(constraint, ">") {
		targetVersion := strings.TrimPrefix(constraint, ">")
		return dm.compareVersions(version, targetVersion) > 0
	}

	if strings.HasPrefix(constraint, "<=") {
		targetVersion := strings.TrimPrefix(constraint, "<=")
		return dm.compareVersions(version, targetVersion) <= 0
	}

	if strings.HasPrefix(constraint, "<") {
		targetVersion := strings.TrimPrefix(constraint, "<")
		return dm.compareVersions(version, targetVersion) < 0
	}

	if strings.HasPrefix(constraint, "~") {
		// 兼容版本约束，如 ~1.2.0 匹配 >=1.2.0 <1.3.0
		targetVersion := strings.TrimPrefix(constraint, "~")
		return dm.isCompatibleVersion(version, targetVersion)
	}

	if strings.HasPrefix(constraint, "^") {
		// 兼容版本约束，如 ^1.2.0 匹配 >=1.2.0 <2.0.0
		targetVersion := strings.TrimPrefix(constraint, "^")
		return dm.isCaretCompatible(version, targetVersion)
	}

	return version == constraint
}

// compareVersions 比较版本号
func (dm *DefaultDependencyManager) compareVersions(v1, v2 string) int {
	// 简化的版本比较，实际应该使用更完善的语义化版本库
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}

	return 0
}

// isCompatibleVersion 检查兼容版本
func (dm *DefaultDependencyManager) isCompatibleVersion(version, target string) bool {
	// 简化实现，实际应该根据语义化版本规则
	vParts := strings.Split(version, ".")
	tParts := strings.Split(target, ".")

	if len(vParts) < 2 || len(tParts) < 2 {
		return false
	}

	// 主版本和次版本必须相同
	return vParts[0] == tParts[0] && vParts[1] == tParts[1] &&
		dm.compareVersions(version, target) >= 0
}

// isCaretCompatible 检查插入符兼容版本
func (dm *DefaultDependencyManager) isCaretCompatible(version, target string) bool {
	// 简化实现，实际应该根据语义化版本规则
	vParts := strings.Split(version, ".")
	tParts := strings.Split(target, ".")

	if len(vParts) < 1 || len(tParts) < 1 {
		return false
	}

	// 主版本必须相同
	return vParts[0] == tParts[0] && dm.compareVersions(version, target) >= 0
}

// hasCircularDependency 检查循环依赖
func (dm *DefaultDependencyManager) hasCircularDependency(
	name string,
	visited map[string]bool,
	recursionStack map[string]bool,
	requirements []*DependencyRequirement,
) bool {
	visited[name] = true
	recursionStack[name] = true

	// 查找当前依赖的子依赖
	for _, req := range requirements {
		if req.Name == name {
			// 这里应该递归检查子依赖，简化实现
			continue
		}
	}

	recursionStack[name] = false
	return false
}

// DependencyInjector 依赖注入器
type DependencyInjector struct {
	dependencies map[string]interface{}
	mu           sync.RWMutex
}

// NewDependencyInjector 创建依赖注入器
func NewDependencyInjector() *DependencyInjector {
	return &DependencyInjector{
		dependencies: make(map[string]interface{}),
	}
}

// Register 注册依赖实例
func (di *DependencyInjector) Register(name string, instance interface{}) {
	di.mu.Lock()
	defer di.mu.Unlock()
	di.dependencies[name] = instance
}

// Get 获取依赖实例
func (di *DependencyInjector) Get(name string) (interface{}, bool) {
	di.mu.RLock()
	defer di.mu.RUnlock()
	instance, exists := di.dependencies[name]
	return instance, exists
}

// Inject 注入依赖到插件
func (di *DependencyInjector) Inject(plugin interface{}, dependencies []string) error {
	// 这里应该使用反射来注入依赖，简化实现
	return nil
}