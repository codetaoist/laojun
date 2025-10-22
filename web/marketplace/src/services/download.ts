import { ApiResponse, Plugin, DownloadRecord, InstallationStatus } from '@/types';
import { message } from 'antd';

// 下载进度回调类型
export type DownloadProgressCallback = (progress: number, status: string) => void;

// 安装状态回调类型
export type InstallationCallback = (status: InstallationStatus, message?: string) => void;

// 下载配置
interface DownloadConfig {
  pluginId: string;
  version?: string;
  onProgress?: DownloadProgressCallback;
  onComplete?: (filePath: string) => void;
  onError?: (error: Error) => void;
}

// 安装配置
interface InstallationConfig {
  pluginId: string;
  filePath: string;
  onProgress?: InstallationCallback;
  onComplete?: () => void;
  onError?: (error: Error) => void;
}

class DownloadService {
  private downloadQueue: Map<string, DownloadConfig> = new Map();
  private installationQueue: Map<string, InstallationConfig> = new Map();
  private downloadRecords: DownloadRecord[] = [];

  /**
   * 下载插件
   */
  async downloadPlugin(config: DownloadConfig): Promise<string> {
    const { pluginId, version = 'latest', onProgress, onComplete, onError } = config;
    
    try {
      // 检查是否已在下载队列中
      if (this.downloadQueue.has(pluginId)) {
        throw new Error('插件正在下载中，请稍候');
      }

      // 添加到下载队列
      this.downloadQueue.set(pluginId, config);

      // 模拟下载进度
      onProgress?.(0, '开始下载...');
      
      // 获取下载链接
      const downloadUrl = await this.getDownloadUrl(pluginId, version);
      
      onProgress?.(10, '获取下载链接成功');

      // 模拟下载过程
      const filePath = await this.simulateDownload(pluginId, downloadUrl, onProgress);
      
      onProgress?.(100, '下载完成');

      // 记录下载历史
      const downloadRecord: DownloadRecord = {
        id: `download-${Date.now()}`,
        pluginId,
        version,
        downloadedAt: new Date().toISOString(),
        filePath,
        fileSize: Math.floor(Math.random() * 10000000) + 1000000, // 模拟文件大小
        status: 'completed',
      };
      
      this.downloadRecords.push(downloadRecord);

      // 从下载队列中移除
      this.downloadQueue.delete(pluginId);

      onComplete?.(filePath);
      message.success('插件下载完成');
      
      return filePath;
    } catch (error) {
      this.downloadQueue.delete(pluginId);
      const errorMessage = error instanceof Error ? error.message : '下载失败';
      onError?.(new Error(errorMessage));
      message.error(errorMessage);
      throw error;
    }
  }

  /**
   * 安装插件
   */
  async installPlugin(config: InstallationConfig): Promise<void> {
    const { pluginId, filePath, onProgress, onComplete, onError } = config;
    
    try {
      // 检查是否已在安装队列中
      if (this.installationQueue.has(pluginId)) {
        throw new Error('插件正在安装中，请稍候');
      }

      // 添加到安装队列
      this.installationQueue.set(pluginId, config);

      // 模拟安装过程
      onProgress?.('installing', '正在安装插件...');
      
      // 验证插件文件
      await this.validatePluginFile(filePath);
      onProgress?.('installing', '验证插件文件...');
      
      // 解压插件
      await this.extractPlugin(filePath);
      onProgress?.('installing', '解压插件文件...');
      
      // 安装依赖
      await this.installDependencies(pluginId);
      onProgress?.('installing', '安装依赖...');
      
      // 注册插件
      await this.registerPlugin(pluginId);
      onProgress?.('installed', '插件安装完成');

      // 从安装队列中移除
      this.installationQueue.delete(pluginId);

      onComplete?.();
      message.success('插件安装成功');
    } catch (error) {
      this.installationQueue.delete(pluginId);
      const errorMessage = error instanceof Error ? error.message : '安装失败';
      onProgress?.('failed', errorMessage);
      onError?.(new Error(errorMessage));
      message.error(errorMessage);
      throw error;
    }
  }

  /**
   * 卸载插件
   */
  async uninstallPlugin(pluginId: string): Promise<void> {
    try {
      // 停止插件运行
      await this.stopPlugin(pluginId);
      
      // 移除插件文件
      await this.removePluginFiles(pluginId);
      
      // 清理注册信息
      await this.unregisterPlugin(pluginId);
      
      message.success('插件卸载成功');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '卸载失败';
      message.error(errorMessage);
      throw error;
    }
  }

  /**
   * 更新插件
   */
  async updatePlugin(pluginId: string, newVersion: string): Promise<void> {
    try {
      // 下载新版本
      const filePath = await this.downloadPlugin({
        pluginId,
        version: newVersion,
      });
      
      // 备份当前版本
      await this.backupCurrentVersion(pluginId);
      
      // 安装新版本
      await this.installPlugin({
        pluginId,
        filePath,
      });
      
      message.success('插件更新成功');
    } catch (error) {
      // 如果更新失败，尝试恢复备份
      await this.restoreBackup(pluginId);
      const errorMessage = error instanceof Error ? error.message : '更新失败';
      message.error(errorMessage);
      throw error;
    }
  }

  /**
   * 获取下载历史
   */
  getDownloadHistory(): DownloadRecord[] {
    return [...this.downloadRecords].sort((a, b) => 
      new Date(b.downloadedAt).getTime() - new Date(a.downloadedAt).getTime()
    );
  }

  /**
   * 获取下载状态
   */
  getDownloadStatus(pluginId: string): 'downloading' | 'completed' | 'failed' | 'idle' {
    if (this.downloadQueue.has(pluginId)) {
      return 'downloading';
    }
    
    const record = this.downloadRecords.find(r => r.pluginId === pluginId);
    return record?.status || 'idle';
  }

  /**
   * 获取安装状态
   */
  getInstallationStatus(pluginId: string): InstallationStatus {
    if (this.installationQueue.has(pluginId)) {
      return 'installing';
    }
    
    // 这里应该从本地存储或API获取实际的安装状态
    // 暂时返回模拟状态
    return 'not_installed';
  }

  /**
   * 取消下载
   */
  cancelDownload(pluginId: string): void {
    if (this.downloadQueue.has(pluginId)) {
      this.downloadQueue.delete(pluginId);
      message.info('下载已取消');
    }
  }

  // 私有方法

  private async getDownloadUrl(pluginId: string, version: string): Promise<string> {
    // 模拟API调用获取下载链接
    await new Promise(resolve => setTimeout(resolve, 500));
    return `https://api.marketplace.com/plugins/${pluginId}/download?version=${version}`;
  }

  private async simulateDownload(
    pluginId: string, 
    url: string, 
    onProgress?: DownloadProgressCallback
  ): Promise<string> {
    // 模拟下载进度
    for (let progress = 10; progress <= 90; progress += 10) {
      await new Promise(resolve => setTimeout(resolve, 200));
      onProgress?.(progress, `下载中... ${progress}%`);
    }
    
    // 模拟文件保存路径
    return `/downloads/${pluginId}-${Date.now()}.zip`;
  }

  private async validatePluginFile(filePath: string): Promise<void> {
    // 模拟文件验证
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // 随机模拟验证失败
    if (Math.random() < 0.1) {
      throw new Error('插件文件损坏或格式不正确');
    }
  }

  private async extractPlugin(filePath: string): Promise<void> {
    // 模拟解压过程
    await new Promise(resolve => setTimeout(resolve, 1000));
  }

  private async installDependencies(pluginId: string): Promise<void> {
    // 模拟安装依赖
    await new Promise(resolve => setTimeout(resolve, 1500));
  }

  private async registerPlugin(pluginId: string): Promise<void> {
    // 模拟注册插件
    await new Promise(resolve => setTimeout(resolve, 500));
  }

  private async stopPlugin(pluginId: string): Promise<void> {
    // 模拟停止插件
    await new Promise(resolve => setTimeout(resolve, 300));
  }

  private async removePluginFiles(pluginId: string): Promise<void> {
    // 模拟删除插件文件
    await new Promise(resolve => setTimeout(resolve, 500));
  }

  private async unregisterPlugin(pluginId: string): Promise<void> {
    // 模拟取消注册
    await new Promise(resolve => setTimeout(resolve, 300));
  }

  private async backupCurrentVersion(pluginId: string): Promise<void> {
    // 模拟备份当前版本
    await new Promise(resolve => setTimeout(resolve, 800));
  }

  private async restoreBackup(pluginId: string): Promise<void> {
    // 模拟恢复备份
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
}

// 导出单例实例
export const downloadService = new DownloadService();

// 导出便捷函数
export const downloadPlugin = (config: DownloadConfig) => downloadService.downloadPlugin(config);
export const installPlugin = (config: InstallationConfig) => downloadService.installPlugin(config);
export const uninstallPlugin = (pluginId: string) => downloadService.uninstallPlugin(pluginId);
export const updatePlugin = (pluginId: string, version: string) => downloadService.updatePlugin(pluginId, version);

export default downloadService;