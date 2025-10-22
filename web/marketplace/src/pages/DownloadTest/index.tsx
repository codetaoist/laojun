import React from 'react';
import { Button, Card, Space, Typography, Divider } from 'antd';
import { DownloadOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { useDownloadStore } from '@/stores/download';

const { Title, Text } = Typography;

const DownloadTest: React.FC = () => {
  const { 
    downloadPlugin, 
    installPlugin, 
    updatePlugin, 
    uninstallPlugin,
    getActiveTasksCount,
    isPluginDownloading,
    isPluginInstalling 
  } = useDownloadStore();

  const testPlugins = [
    {
      id: 'test-plugin-1',
      name: '测试插件 1',
      version: '1.0.0',
      description: '这是一个用于测试下载功能的插件',
      price: 0
    },
    {
      id: 'test-plugin-2', 
      name: '测试插件 2',
      version: '2.1.0',
      description: '另一个测试插件，用于验证并发下载',
      price: 29.99
    },
    {
      id: 'test-plugin-3',
      name: '测试插件 3', 
      version: '1.5.2',
      description: '第三个测试插件，用于测试更新功能',
      price: 0
    }
  ];

  const handleDownload = (plugin: typeof testPlugins[0]) => {
    downloadPlugin(plugin.id, {
      name: plugin.name,
      version: plugin.version,
      description: plugin.description,
      price: plugin.price,
      downloadUrl: `https://example.com/plugins/${plugin.id}.zip`,
      size: Math.floor(Math.random() * 50 + 10) * 1024 * 1024, // 10-60MB
    });
  };

  const handleInstall = (pluginId: string) => {
    installPlugin(pluginId);
  };

  const handleUpdate = (plugin: typeof testPlugins[0]) => {
    updatePlugin(plugin.id, {
      name: plugin.name,
      version: plugin.version,
      description: plugin.description,
      price: plugin.price,
      downloadUrl: `https://example.com/plugins/${plugin.id}.zip`,
      size: Math.floor(Math.random() * 50 + 10) * 1024 * 1024,
    });
  };

  const handleUninstall = (pluginId: string) => {
    uninstallPlugin(pluginId);
  };

  const activeTasksCount = getActiveTasksCount();

  return (
    <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
      <Title level={2}>下载管理器功能测试</Title>
      <Text type="secondary">
        这个页面用于测试下载管理器的各种功能。当前活跃任务数: {activeTasksCount}
      </Text>
      
      <Divider />

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        {testPlugins.map(plugin => (
          <Card 
            key={plugin.id}
            title={plugin.name}
            extra={<Text type="secondary">v{plugin.version}</Text>}
          >
            <Text>{plugin.description}</Text>
            <br />
            <Text strong>价格: {plugin.price === 0 ? '免费' : `¥${plugin.price}`}</Text>
            
            <div style={{ marginTop: '16px' }}>
              <Space>
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  onClick={() => handleDownload(plugin)}
                  loading={isPluginDownloading(plugin.id)}
                  disabled={isPluginDownloading(plugin.id) || isPluginInstalling(plugin.id)}
                >
                  {isPluginDownloading(plugin.id) ? '下载中...' : '下载'}
                </Button>
                
                <Button
                  icon={<PlayCircleOutlined />}
                  onClick={() => handleInstall(plugin.id)}
                  loading={isPluginInstalling(plugin.id)}
                  disabled={isPluginDownloading(plugin.id) || isPluginInstalling(plugin.id)}
                >
                  {isPluginInstalling(plugin.id) ? '安装中...' : '安装'}
                </Button>
                
                <Button
                  onClick={() => handleUpdate(plugin)}
                  disabled={isPluginDownloading(plugin.id) || isPluginInstalling(plugin.id)}
                >
                  更新
                </Button>
                
                <Button
                  danger
                  onClick={() => handleUninstall(plugin.id)}
                  disabled={isPluginDownloading(plugin.id) || isPluginInstalling(plugin.id)}
                >
                  卸载
                </Button>
              </Space>
            </div>
          </Card>
        ))}
      </Space>

      <Divider />
      
      <Card title="测试说明">
        <Space direction="vertical">
          <Text>• 点击"下载"按钮开始下载插件</Text>
          <Text>• 点击"安装"按钮安装已下载的插件</Text>
          <Text>• 点击"更新"按钮更新插件到新版本</Text>
          <Text>• 点击"卸载"按钮移除已安装的插件</Text>
          <Text>• 查看Header中的下载按钮，应该显示活跃任务数量</Text>
          <Text>• 点击Header中的下载按钮打开下载管理器查看详细进度</Text>
        </Space>
      </Card>
    </div>
  );
};

export default DownloadTest;