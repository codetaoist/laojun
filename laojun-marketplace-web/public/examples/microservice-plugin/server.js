const express = require('express');
const cors = require('cors');

const app = express();
const PORT = process.env.PORT || 8080;

// 中间件
app.use(cors());
app.use(express.json());

// 健康检查端点
app.get('/health', (req, res) => {
  res.json({ 
    status: 'healthy', 
    timestamp: new Date().toISOString(),
    service: 'microservice-example'
  });
});

// 数据处理API
app.post('/api/process', (req, res) => {
  try {
    const { data } = req.body;
    
    if (!data) {
      return res.status(400).json({
        error: 'Missing required parameter: data'
      });
    }

    // 模拟数据处理
    const processedData = {
      original: data,
      processed: data.toUpperCase(),
      timestamp: new Date().toISOString(),
      length: data.length,
      wordCount: data.split(' ').length
    };

    // 发送处理完成事件（在实际应用中，这里会发送到事件总线）
    console.log('Data processed event:', {
      type: 'data_processed',
      data: processedData
    });

    res.json({
      success: true,
      result: processedData
    });
  } catch (error) {
    console.error('Processing error:', error);
    res.status(500).json({
      error: 'Internal server error',
      message: error.message
    });
  }
});

// 获取插件信息
app.get('/api/info', (req, res) => {
  res.json({
    name: 'microservice-example',
    version: '1.0.0',
    type: 'microservice',
    status: 'running',
    uptime: process.uptime()
  });
});

// 错误处理中间件
app.use((err, req, res, next) => {
  console.error('Unhandled error:', err);
  res.status(500).json({
    error: 'Internal server error'
  });
});

// 404 处理
app.use((req, res) => {
  res.status(404).json({
    error: 'Not found',
    path: req.path
  });
});

// 启动服务器
app.listen(PORT, () => {
  console.log(`Microservice plugin running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
  console.log(`API endpoint: http://localhost:${PORT}/api/process`);
});