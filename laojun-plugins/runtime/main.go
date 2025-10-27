package runtime

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// 命令行参数
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		logLevel   = flag.String("log-level", "info", "日志级别 (debug, info, warn, error)")
		port       = flag.Int("port", 8080, "HTTP服务端口")
		help       = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		fmt.Println("插件运行时环境")
		fmt.Println("用法:")
		flag.PrintDefaults()
		return
	}

	// 初始化日志
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.Fatalf("无效的日志级别: %v", err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.WithFields(logrus.Fields{
		"config": *configPath,
		"port":   *port,
	}).Info("启动插件运行时环境")

	// 创建插件加载器
	loader := NewDefaultPluginLoader(logger)

	// 创建沙箱
	sandbox := NewDefaultSandbox(logger)

	// 创建插件管理器
	manager := NewDefaultPluginManager(loader, sandbox, logger)

	// 创建插件注册中心
	_ = NewDefaultPluginRegistry(logger)

	// 创建监控器
	_ = NewDefaultMonitor(manager, 30*time.Second, logger)

	// 创建执行器
	_ = NewDefaultExecutor(manager, 10, 1000, logger)

	// 创建事件总线
	_ = NewDefaultEventBus(1000, 5, logger)

	// 创建依赖管理器
	_ = NewDefaultDependencyManager(logger)

	// 创建插件引擎
	engine, err := NewPluginEngine(nil, logger)
	if err != nil {
		logger.Fatalf("创建插件引擎失败: %v", err)
	}

	// 启动引擎
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		logger.Fatalf("启动插件引擎失败: %v", err)
	}

	logger.Info("插件运行时环境启动成功")

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("收到停止信号，正在关闭...")

	// 停止引擎
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := engine.Stop(shutdownCtx); err != nil {
		logger.Errorf("停止插件引擎失败: %v", err)
	}

	logger.Info("插件运行时环境已停止")
}