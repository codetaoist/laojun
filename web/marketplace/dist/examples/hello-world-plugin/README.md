# Hello World 插件示例

这是一个简单的 Hello World 插件示例，展示了如何创建一个基本的插件。

## 功能特性

- 显示 Hello World 消息
- 演示基本的插件结构
- 包含完整的配置文件

## 文件结构

```
hello-world-plugin/
├── plugin.json      # 插件配置文件
├── index.js         # 主要代码文件
└── README.md        # 说明文档
```

## 使用方法

1. 下载插件包
2. 在插件市场中上传
3. 安装并激活插件
4. 使用 `hello` 命令查看效果

## 开发说明

### plugin.json 配置

- `name`: 插件唯一标识符
- `displayName`: 显示名称
- `version`: 版本号
- `type`: 插件类型（in_process/microservice）
- `runtime`: 运行时环境
- `entry_point`: 入口文件

### 主要方法

- `initialize()`: 插件初始化
- `sayHello()`: Hello 命令实现
- `dispose()`: 插件卸载

## 扩展开发

基于此示例，您可以：

1. 添加更多命令
2. 实现事件监听
3. 调用外部 API
4. 创建用户界面
5. 处理用户输入

## 许可证

MIT License