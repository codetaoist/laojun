-- Create sample plugin data
INSERT INTO plugins (
    name, description, short_description, author, version, developer_id, category_id,
    price, is_free, is_featured, is_active, download_count, rating, review_count,
    icon_url, banner_url, tags, requirements
) VALUES 
-- Development Tools
('Code Formatter', 'A powerful code formatting tool that supports multiple programming languages and automatically formats your code for better readability.', 'Smart code formatting tool', 'TaiShang Studio', '1.2.0', '05f59353-2dce-4dfd-95a0-fe279a213418', '1ac15c1d-8f2b-4671-9693-15f7d4b06b80', 0.00, true, true, true, 1250, 4.8, 45, '/icons/code-formatter.svg', '/banners/code-formatter.jpg', ARRAY['code', 'format', 'dev'], '{"min_version": "1.0.0", "dependencies": []}'),

('API Tester', 'Professional API testing tool supporting REST, GraphQL and other API types with intuitive interface and detailed reports.', 'Convenient API testing tool', 'TaiShang Studio', '2.1.3', '05f59353-2dce-4dfd-95a0-fe279a213418', '1ac15c1d-8f2b-4671-9693-15f7d4b06b80', 29.99, false, true, true, 890, 4.6, 32, '/icons/api-tester.svg', '/banners/api-tester.jpg', ARRAY['api', 'test', 'dev'], '{"min_version": "1.0.0", "dependencies": []}'),

-- UI Components
('Responsive Table', 'Feature-rich responsive table component with sorting, filtering, pagination, export and more, adapts to all screen sizes.', 'High-performance responsive data table', 'TaiShang Studio', '3.0.1', '05f59353-2dce-4dfd-95a0-fe279a213418', '23cf7876-bfd2-4fd3-bdd9-a93cfe4eb5c7', 19.99, false, false, true, 2100, 4.7, 78, '/icons/responsive-table.svg', '/banners/responsive-table.jpg', ARRAY['table', 'ui', 'responsive'], '{"min_version": "1.0.0", "dependencies": []}'),

('Chart Suite', 'Visualization suite with 30+ chart types, supporting real-time data updates, interactive operations and custom themes.', 'Professional data visualization charts', 'TaiShang Studio', '1.5.2', '05f59353-2dce-4dfd-95a0-fe279a213418', '23cf7876-bfd2-4fd3-bdd9-a93cfe4eb5c7', 0.00, true, true, true, 3200, 4.9, 156, '/icons/chart-suite.svg', '/banners/chart-suite.jpg', ARRAY['chart', 'visualization', 'ui'], '{"min_version": "1.0.0", "dependencies": []}'),

-- Productivity Tools
('Smart Notes', 'AI-integrated note management tool with smart categorization, content search, tag management and multi-device sync.', 'AI-powered smart note system', 'TaiShang Studio', '2.3.0', '05f59353-2dce-4dfd-95a0-fe279a213418', 'a4c79735-dba7-4f11-9d07-895bd41a429f', 39.99, false, true, true, 1680, 4.5, 89, '/icons/smart-notes.svg', '/banners/smart-notes.jpg', ARRAY['notes', 'ai', 'productivity'], '{"min_version": "1.0.0", "dependencies": []}'),

('Task Master', 'Professional task management tool with Gantt charts, Kanban views, time tracking and team collaboration features.', 'Efficient project task management', 'TaiShang Studio', '1.8.5', '05f59353-2dce-4dfd-95a0-fe279a213418', 'a4c79735-dba7-4f11-9d07-895bd41a429f', 0.00, true, false, true, 950, 4.4, 67, '/icons/task-master.svg', '/banners/task-master.jpg', ARRAY['task', 'management', 'productivity'], '{"min_version": "1.0.0", "dependencies": []}');