-- Seed plugin marketplace data

-- Insert plugin categories
INSERT INTO mp_categories (id, name, description, icon, color, sort_order, is_active)
VALUES
  ('11111111-2222-3333-4444-555555555555', '开发工具', '提升开发效率的工具插件', 'Code', '#3B82F6', 1, TRUE),
  ('22222222-3333-4444-5555-666666666666', '数据处理', '数据分析和处理相关插件', 'Database', '#10B981', 2, TRUE),
  ('33333333-4444-5555-6666-777777777777', '图像处理', '图像编辑和处理工具', 'Image', '#F59E0B', 3, TRUE),
  ('44444444-5555-6666-7777-888888888888', '文本分析', '文本处理和分析工具', 'FileText', '#8B5CF6', 4, TRUE),
  ('55555555-6666-7777-8888-999999999999', 'API连接器', 'API集成和连接工具', 'Link', '#EF4444', 5, TRUE)
ON CONFLICT (id) DO NOTHING;

-- Insert sample plugins
INSERT INTO mp_plugins (id, name, description, short_description, author, developer_id, version, icon_url, price, rating, download_count, is_featured, category_id, status, review_status)
VALUES
  ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', 'Code Formatter Pro', '强大的代码格式化工具，支持多种编程语言的自动格式化和代码美化。提供智能缩进、语法高亮、代码折叠等功能，让您的代码更加整洁和易读。', '专业的代码格式化工具', 'DevTools Team', '11111111-1111-1111-1111-111111111111', '1.2.0', 'https://cdn.example.com/icons/code-formatter.png', 29.99, 4.8, 1250, TRUE, '11111111-2222-3333-4444-555555555555', 'active', 'approved'),
  ('bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'Data Analyzer', '高效的数据分析插件，提供数据可视化、统计分析和报表生成功能。支持多种数据格式，包括CSV、JSON、XML等，帮助您快速洞察数据价值。', '数据分析和可视化工具', 'Analytics Pro', '11111111-1111-1111-1111-111111111111', '2.1.5', 'https://cdn.example.com/icons/data-analyzer.png', 49.99, 4.6, 890, TRUE, '22222222-3333-4444-5555-666666666666', 'active', 'approved'),
  ('cccccccc-dddd-eeee-ffff-000000000000', 'Image Filter Studio', '专业的图像滤镜工具，提供丰富的滤镜效果和图像处理功能。包括模糊、锐化、色彩调整、特效滤镜等，让您的图片更加精美。', '图像滤镜和处理工具', 'ImageTech', '11111111-1111-1111-1111-111111111111', '1.8.3', 'https://cdn.example.com/icons/image-filter.png', 19.99, 4.5, 2100, FALSE, '33333333-4444-5555-6666-777777777777', 'active', 'approved'),
  ('dddddddd-eeee-ffff-0000-111111111111', 'Text Sentiment Analyzer', '智能文本情感分析工具，使用先进的自然语言处理技术分析文本情感倾向。支持多语言分析，提供详细的情感报告和可视化图表。', '文本情感分析工具', 'NLP Solutions', '11111111-1111-1111-1111-111111111111', '1.5.2', 'https://cdn.example.com/icons/sentiment-analyzer.png', 34.99, 4.7, 680, TRUE, '44444444-5555-6666-7777-888888888888', 'active', 'approved'),
  ('eeeeeeee-ffff-0000-1111-222222222222', 'REST API Connector', '通用的REST API连接器，简化API集成和数据交换。支持多种认证方式，提供请求构建器、响应解析器和错误处理机制。', 'REST API集成工具', 'API Solutions', '11111111-1111-1111-1111-111111111111', '3.0.1', 'https://cdn.example.com/icons/api-connector.png', 0.00, 4.4, 3200, FALSE, '55555555-6666-7777-8888-999999999999', 'active', 'approved'),
  ('ffffffff-0000-1111-2222-333333333333', 'Quick Debugger', '快速调试工具，提供断点设置、变量监控和性能分析功能。支持多种编程语言，帮助开发者快速定位和解决代码问题。', '快速调试和性能分析', 'Debug Masters', '11111111-1111-1111-1111-111111111111', '2.3.0', 'https://cdn.example.com/icons/debugger.png', 24.99, 4.9, 1800, TRUE, '11111111-2222-3333-4444-555555555555', 'active', 'approved'),
  ('00000000-1111-2222-3333-444444444444', 'CSV Data Processor', '专业的CSV数据处理工具，支持大文件处理和数据转换。提供数据清洗、格式转换、统计分析等功能，让数据处理变得简单高效。', 'CSV数据处理和转换', 'Data Tools Inc', '11111111-1111-1111-1111-111111111111', '1.4.7', 'https://cdn.example.com/icons/csv-processor.png', 15.99, 4.3, 950, FALSE, '22222222-3333-4444-5555-666666666666', 'active', 'approved')
ON CONFLICT (id) DO NOTHING;