/**
 * Hello World 插件示例
 * 这是一个简单的插件，展示如何创建基本的插件功能
 */

class HelloWorld {
  constructor() {
    this.name = 'Hello World Plugin';
    this.version = '1.0.0';
  }

  /**
   * 插件初始化方法
   */
  initialize() {
    console.log('Hello World 插件已初始化');
    this.registerCommands();
  }

  /**
   * 注册插件命令
   */
  registerCommands() {
    // 注册 hello 命令
    this.registerCommand('hello', this.sayHello.bind(this));
  }

  /**
   * Hello 命令实现
   */
  sayHello() {
    const message = 'Hello, World! 这是来自插件的问候！';
    
    // 显示消息给用户
    this.showMessage(message);
    
    return {
      success: true,
      message: message
    };
  }

  /**
   * 显示消息的辅助方法
   */
  showMessage(message) {
    // 这里可以调用宿主应用的 API 来显示消息
    if (typeof window !== 'undefined' && window.showNotification) {
      window.showNotification(message);
    } else {
      console.log(message);
    }
  }

  /**
   * 注册命令的辅助方法
   */
  registerCommand(name, handler) {
    // 这里可以调用宿主应用的 API 来注册命令
    if (typeof window !== 'undefined' && window.registerPluginCommand) {
      window.registerPluginCommand(name, handler);
    } else {
      console.log(`命令 "${name}" 已注册`);
    }
  }

  /**
   * 插件卸载方法
   */
  dispose() {
    console.log('Hello World 插件已卸载');
  }
}

// 导出插件类
if (typeof module !== 'undefined' && module.exports) {
  module.exports = HelloWorld;
} else if (typeof window !== 'undefined') {
  window.HelloWorld = HelloWorld;
}