package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 统一日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
}

// Config 日志配置
type Config struct {
	Level       string     `yaml:"level" env:"LOG_LEVEL" config:"log.level" default:"info"`
	Format      string     `yaml:"format" env:"LOG_FORMAT" config:"log.format" default:"json"`
	Output      string     `yaml:"output" env:"LOG_OUTPUT" config:"log.output" default:"stdout"`
	File        FileConfig `yaml:"file"`
	Service     string     `yaml:"service" env:"SERVICE_NAME" config:"log.service" default:"unknown"`
	Version     string     `yaml:"version" env:"SERVICE_VERSION" config:"log.version" default:"unknown"`
	Environment string     `yaml:"environment" env:"ENVIRONMENT" config:"log.environment" default:"development"`
}

// FileConfig 文件日志配置
type FileConfig struct {
	Filename   string `yaml:"filename" env:"LOG_FILE" config:"log.file.filename" default:""`
	MaxSize    int    `yaml:"max_size" env:"LOG_FILE_MAX_SIZE" config:"log.file.max_size" default:"100"`
	MaxBackups int    `yaml:"max_backups" env:"LOG_FILE_MAX_BACKUPS" config:"log.file.max_backups" default:"3"`
	MaxAge     int    `yaml:"max_age" env:"LOG_FILE_MAX_AGE" config:"log.file.max_age" default:"28"`
	Compress   bool   `yaml:"compress" env:"LOG_FILE_COMPRESS" config:"log.file.compress" default:"true"`
}

// loggerImpl logrus实现
type loggerImpl struct {
	logger *logrus.Logger
	entry  *logrus.Entry
	config Config
}

// New 创建新的日志实例
func New(config Config) Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	case "custom":
		logger.SetFormatter(&CustomFormatter{})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置输出
	switch strings.ToLower(config.Output) {
	case "file":
		// 如果没有指定文件名，使用默认路径（相对于当前服务目录）
		filename := config.File.Filename
		if filename == "" {
			filename = fmt.Sprintf("./logs/%s.log", config.Service)
		}
		
		// 确保日志目录存在
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.WithError(err).Error("创建日志目录失败")
		}

		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    config.File.MaxSize,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,
			Compress:   config.File.Compress,
		}
		logger.SetOutput(fileWriter)
	case "both":
		// 如果没有指定文件名，使用默认路径（相对于当前服务目录）
		filename := config.File.Filename
		if filename == "" {
			filename = fmt.Sprintf("./logs/%s.log", config.Service)
		}
		
		// 同时输出到文件和控制台
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.WithError(err).Error("创建日志目录失败")
		}

		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    config.File.MaxSize,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,
			Compress:   config.File.Compress,
		}
		logger.SetOutput(io.MultiWriter(os.Stdout, fileWriter))
	default:
		logger.SetOutput(os.Stdout)
	}

	// 添加调用信息
	logger.SetReportCaller(true)

	// 创建基础entry，包含服务信息
	entry := logger.WithFields(logrus.Fields{
		"service":     config.Service,
		"version":     config.Version,
		"environment": config.Environment,
		"hostname":    getHostname(),
	})

	return &loggerImpl{
		logger: logger,
		entry:  entry,
		config: config,
	}
}

// Debug 调试日志
func (l *loggerImpl) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf 格式化调试日志
func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Info 信息日志
func (l *loggerImpl) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof 格式化信息日志
func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warn 警告日志
func (l *loggerImpl) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf 格式化警告日志
func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Error 错误日志
func (l *loggerImpl) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf 格式化错误日志
func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatal 致命错误日志
func (l *loggerImpl) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// WithField 添加字段
func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	return &loggerImpl{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
		config: l.config,
	}
}

// WithFields 添加多个字段
func (l *loggerImpl) WithFields(fields map[string]interface{}) Logger {
	return &loggerImpl{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
		config: l.config,
	}
}

// WithContext 添加上下文信息
func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	entry := l.entry.WithContext(ctx)

	// 从上下文中提取常用字段
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}
	if spanID := ctx.Value("span_id"); spanID != nil {
		entry = entry.WithField("span_id", spanID)
	}

	return &loggerImpl{
		logger: l.logger,
		entry:  entry,
		config: l.config,
	}
}

// WithError 添加错误信息
func (l *loggerImpl) WithError(err error) Logger {
	entry := l.entry.WithError(err)

	// 添加错误堆栈信息
	if pc, file, line, ok := runtime.Caller(1); ok {
		entry = entry.WithFields(logrus.Fields{
			"error_file": file,
			"error_line": line,
			"error_func": runtime.FuncForPC(pc).Name(),
		})
	}

	// 获取调用信息
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry = entry.WithFields(logrus.Fields{
			"caller_file": file,
			"caller_line": line,
			"caller_func": runtime.FuncForPC(pc).Name(),
		})
	}

	return &loggerImpl{
		logger: l.logger,
		entry:  entry,
		config: l.config,
	}
}

// CustomFormatter 自定义格式化日志
type CustomFormatter struct{}

// Format 格式化日志
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(time.RFC3339)

	// 获取调用信息
	caller := ""
	if entry.HasCaller() {
		caller = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
	}

	// 构建字段字符串
	fields := ""
	for k, v := range entry.Data {
		fields += fmt.Sprintf(" %s=%v", k, v)
	}

	log := fmt.Sprintf("[%s] %s %s %s%s\n",
		timestamp, strings.ToUpper(entry.Level.String()), caller, entry.Message, fields)

	return []byte(log), nil
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// DefaultLogger 默认日志实例
var DefaultLogger Logger

// init 初始化默认日志实例
func init() {
	DefaultLogger = New(Config{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		Service:     "default",
		Version:     "1.0.0",
		Environment: "development",
	})
}

// 全局日志方法
func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	DefaultLogger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	DefaultLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	DefaultLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return DefaultLogger.WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
	return DefaultLogger.WithFields(fields)
}

func WithContext(ctx context.Context) Logger {
	return DefaultLogger.WithContext(ctx)
}

func WithError(err error) Logger {
	return DefaultLogger.WithError(err)
}
