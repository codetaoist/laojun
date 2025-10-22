// Hello JavaScript Plugin
// 这是一个示例JavaScript插件，演示插件系统的基本功能

const plugin = {
    // 插件元数据
    metadata: {
        id: "hello-js-plugin",
        name: "Hello JavaScript Plugin",
        version: "1.0.0",
        description: "A simple Hello World plugin written in JavaScript",
        type: "filter",
        runtime: "js",
        author: "Taishanglaojun Team",
        dependencies: [],
        permissions: ["network.http"],
        config: {
            greeting: "Hello",
            language: "en"
        }
    },

    // 插件状态
    _config: {},
    _startTime: null,
    _counter: 0,
    _status: "loaded",

    // 初始化插件
    initialize: function(configStr) {
        try {
            this._config = JSON.parse(configStr);
            this._startTime = new Date();
            this._counter = 0;
            this._status = "initialized";
            
            console.log("Hello JavaScript Plugin initialized with config:", this._config);
            return null; // 成功返回null，失败返回错误信息
        } catch (error) {
            return "Failed to initialize plugin: " + error.message;
        }
    },

    // 启动插件
    start: function() {
        try {
            this._status = "running";
            console.log("Hello JavaScript Plugin started");
            return null;
        } catch (error) {
            return "Failed to start plugin: " + error.message;
        }
    },

    // 停止插件
    stop: function() {
        try {
            this._status = "stopped";
            console.log("Hello JavaScript Plugin stopped");
            return null;
        } catch (error) {
            return "Failed to stop plugin: " + error.message;
        }
    },

    // 处理请求
    handleRequest: function(requestStr) {
        try {
            this._counter++;
            
            // 解析请求
            const request = JSON.parse(requestStr);
            const requestData = request.data || {};

            // 获取配置中的问候语
            const greeting = this._config.greeting || "Hello";

            // 获取请求中的名称
            const name = requestData.name || "World";

            // 计算运行时间
            const uptime = this._startTime ? 
                Math.floor((new Date() - this._startTime) / 1000) + " seconds" : 
                "unknown";

            // 构造响应数据
            const responseData = {
                message: greeting + ", " + name + "!",
                counter: this._counter,
                uptime: uptime,
                plugin_id: this.metadata.id,
                timestamp: new Date().toISOString(),
                request_id: request.id
            };

            // 如果请求包含特殊参数，添加额外信息
            if (requestData.include_stats === true) {
                responseData.stats = {
                    total_requests: this._counter,
                    start_time: this._startTime ? this._startTime.toISOString() : null,
                    config: this._config,
                    status: this._status
                };
            }

            // 如果请求包含计算参数，执行简单计算
            if (requestData.calculate && typeof requestData.calculate === "object") {
                const calc = requestData.calculate;
                if (calc.operation && calc.a !== undefined && calc.b !== undefined) {
                    let result;
                    switch (calc.operation) {
                        case "add":
                            result = calc.a + calc.b;
                            break;
                        case "subtract":
                            result = calc.a - calc.b;
                            break;
                        case "multiply":
                            result = calc.a * calc.b;
                            break;
                        case "divide":
                            result = calc.b !== 0 ? calc.a / calc.b : "Error: Division by zero";
                            break;
                        default:
                            result = "Error: Unknown operation";
                    }
                    responseData.calculation_result = result;
                }
            }

            // 构造响应
            const response = {
                success: true,
                data: responseData,
                message: "Request processed successfully"
            };

            return JSON.stringify(response);
        } catch (error) {
            // 错误响应
            const errorResponse = {
                success: false,
                data: null,
                message: "Error processing request: " + error.message,
                error: {
                    type: "processing_error",
                    details: error.message,
                    timestamp: new Date().toISOString()
                }
            };
            return JSON.stringify(errorResponse);
        }
    },

    // 处理事件
    handleEvent: function(eventStr) {
        try {
            const event = JSON.parse(eventStr);
            console.log("Hello JavaScript Plugin received event:", event.type, "from", event.source);
            
            // 根据事件类型执行不同的处理逻辑
            switch (event.type) {
                case "system.startup":
                    console.log("System startup event received");
                    break;
                case "plugin.loaded":
                    if (event.data && event.data.id) {
                        console.log("Another plugin loaded:", event.data.id);
                    }
                    break;
                case "config.updated":
                    console.log("Configuration updated event received");
                    // 这里可以重新加载配置
                    if (event.data && event.data.config) {
                        this._config = Object.assign(this._config, event.data.config);
                    }
                    break;
                default:
                    console.log("Unknown event type:", event.type);
            }

            return null; // 成功处理
        } catch (error) {
            return "Error handling event: " + error.message;
        }
    },

    // 获取插件状态
    getStatus: function() {
        return this._status;
    },

    // 获取插件健康状态
    getHealth: function() {
        try {
            const uptime = this._startTime ? 
                Math.floor((new Date() - this._startTime) / 1000) : 0;

            const health = {
                status: "healthy",
                message: "Plugin is running normally",
                timestamp: new Date().toISOString(),
                details: {
                    uptime: uptime + " seconds",
                    total_requests: this._counter,
                    memory_usage: "N/A", // JavaScript中难以准确获取内存使用情况
                    last_request: new Date().toISOString(),
                    plugin_status: this._status
                }
            };

            return JSON.stringify(health);
        } catch (error) {
            const errorHealth = {
                status: "unhealthy",
                message: "Error getting health status: " + error.message,
                timestamp: new Date().toISOString(),
                details: {
                    error: error.message
                }
            };
            return JSON.stringify(errorHealth);
        }
    },

    // 获取插件元数据
    getMetadata: function() {
        return JSON.stringify(this.metadata);
    },

    // 验证配置
    validateConfig: function(configStr) {
        try {
            const config = JSON.parse(configStr);
            
            // 验证必需的配置项
            if (config.greeting && typeof config.greeting !== "string") {
                return "greeting must be a string";
            }

            if (config.language && typeof config.language !== "string") {
                return "language must be a string";
            }

            // 验证支持的语言
            if (config.language) {
                const supportedLanguages = ["en", "zh", "es", "fr", "de"];
                if (supportedLanguages.indexOf(config.language) === -1) {
                    return "unsupported language: " + config.language;
                }
            }

            return null; // 验证成功
        } catch (error) {
            return "Invalid config format: " + error.message;
        }
    },

    // 获取插件信息
    getPluginInfo: function() {
        const info = {
            build_time: new Date().toISOString(),
            js_version: "ES5+",
            dependencies: [],
            capabilities: ["request_handling", "event_handling", "health_check", "calculation"],
            documentation: "https://docs.taishanglaojun.com/plugins/hello-js-plugin",
            examples: {
                basic_request: {
                    data: {
                        name: "World"
                    }
                },
                stats_request: {
                    data: {
                        name: "Admin",
                        include_stats: true
                    }
                },
                calculation_request: {
                    data: {
                        name: "Calculator",
                        calculate: {
                            operation: "add",
                            a: 10,
                            b: 5
                        }
                    }
                }
            }
        };
        return JSON.stringify(info);
    }
};

// 导出插件对象（在V8环境中会被自动识别）
if (typeof module !== 'undefined' && module.exports) {
    module.exports = plugin;
}