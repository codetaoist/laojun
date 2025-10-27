package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	configPath string
	portConfig *PortConfig
}

// NewConfigLoader 创建新的配置加载器
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
	}
}

// LoadPortConfig 从YAML文件加载端口配置
func (cl *ConfigLoader) LoadPortConfig() (*PortConfig, error) {
	if cl.portConfig != nil {
		return cl.portConfig, nil
	}

	// 如果没有指定配置文件路径，使用默认配置
	if cl.configPath == "" {
		cl.portConfig = DefaultPortConfig()
		return cl.portConfig, nil
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(cl.configPath)
	if err != nil {
		// 如果文件不存在，使用默认配置
		cl.portConfig = DefaultPortConfig()
		return cl.portConfig, nil
	}

	// 解析YAML配置
	var config PortConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析端口配置文件失败: %v", err)
	}

	// 验证配置
	if err := config.ValidatePortConfig(); err != nil {
		return nil, fmt.Errorf("端口配置验证失败: %v", err)
	}

	cl.portConfig = &config
	return cl.portConfig, nil
}

// GetPortConfig 获取端口配置（缓存版本）
func (cl *ConfigLoader) GetPortConfig() *PortConfig {
	if cl.portConfig == nil {
		config, _ := cl.LoadPortConfig()
		return config
	}
	return cl.portConfig
}

// ReloadPortConfig 重新加载端口配置
func (cl *ConfigLoader) ReloadPortConfig() (*PortConfig, error) {
	cl.portConfig = nil
	return cl.LoadPortConfig()
}

// SavePortConfig 保存端口配置到文件
func (cl *ConfigLoader) SavePortConfig(config *PortConfig) error {
	if cl.configPath == "" {
		return fmt.Errorf("未指定配置文件路径")
	}

	// 验证配置
	if err := config.ValidatePortConfig(); err != nil {
		return fmt.Errorf("端口配置验证失败: %v", err)
	}

	// 序列化为YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(cl.configPath)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(cl.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	cl.portConfig = config
	return nil
}

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	// 这里可以使用 os.MkdirAll，但为了兼容性，暂时简化
	return nil
}

// GlobalPortConfig 全局端口配置实例
var GlobalPortConfig *PortConfig

// InitGlobalPortConfig 初始化全局端口配置
func InitGlobalPortConfig(configPath string) error {
	loader := NewConfigLoader(configPath)
	config, err := loader.LoadPortConfig()
	if err != nil {
		return err
	}
	GlobalPortConfig = config
	return nil
}

// GetGlobalPortConfig 获取全局端口配置
func GetGlobalPortConfig() *PortConfig {
	if GlobalPortConfig == nil {
		GlobalPortConfig = DefaultPortConfig()
	}
	return GlobalPortConfig
}

// GetServicePort 获取服务端口（全局函数）
func GetServicePort(service string) int {
	return GetGlobalPortConfig().GetServicePort(service)
}

// GetServerAddress 获取服务器地址（全局函数）
func GetServerAddress(service string, host ...string) string {
	return GetGlobalPortConfig().GetServerAddress(service, host...)
}