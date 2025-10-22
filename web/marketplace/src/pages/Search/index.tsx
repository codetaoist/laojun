import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Row, Col, Card, Input, Select, Pagination, Spin, Empty, message } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { Plugin, SearchParams } from '@/types';
import { pluginService } from '@/services/plugin';
import './index.css';

const { Option } = Select;

const Search: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [total, setTotal] = useState(0);
  const [searchForm, setSearchForm] = useState<SearchParams>({
    keyword: searchParams.get('q') || '',
    category: searchParams.get('category') || '',
    sortBy: (searchParams.get('sortBy') as any) || 'relevance',
    page: parseInt(searchParams.get('page') || '1'),
    pageSize: parseInt(searchParams.get('pageSize') || '12'),
  });

  const sortOptions = [
    { value: 'relevance', label: '相关度' },
    { value: 'downloads', label: '下载量' },
    { value: 'rating', label: '评分' },
    { value: 'updated', label: '更新时间' },
    { value: 'created', label: '发布时间' },
    { value: 'name', label: '名称' },
  ];

  const categories = [
    { value: '', label: '全部分类' },
    { value: 'development', label: '开发工具' },
    { value: 'productivity', label: '效率工具' },
    { value: 'design', label: '设计工具' },
    { value: 'entertainment', label: '娱乐工具' },
    { value: 'utility', label: '实用工具' },
  ];

  useEffect(() => {
    searchPlugins();
  }, [searchForm]);

  const searchPlugins = async () => {
    try {
      setLoading(true);
      const response = await pluginService.searchPlugins(searchForm);
      setPlugins(response.data.items || []);
      setTotal(response.data.total || 0);
    } catch (error) {
      message.error('搜索失败，请稍后重试');
      setPlugins([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (value: string) => {
    const newForm = { ...searchForm, keyword: value, page: 1 };
    setSearchForm(newForm);
    updateURL(newForm);
  };

  const handleFilterChange = (key: string, value: any) => {
    const newForm = { ...searchForm, [key]: value, page: 1 };
    setSearchForm(newForm);
    updateURL(newForm);
  };

  const handlePageChange = (page: number, pageSize?: number) => {
    const newForm = { ...searchForm, page, pageSize: pageSize || searchForm.pageSize };
    setSearchForm(newForm);
    updateURL(newForm);
  };

  const updateURL = (form: SearchParams) => {
    const params = new URLSearchParams();
    if (form.keyword) params.set('q', form.keyword);
    if (form.category) params.set('category', form.category);
    if (form.sortBy !== 'relevance') params.set('sortBy', form.sortBy);
    if (form.page !== 1) params.set('page', form.page.toString());
    if (form.pageSize !== 12) params.set('pageSize', form.pageSize.toString());
    setSearchParams(params);
  };

  return (
    <div className="search-page">
      <div className="search-header">
        <div className="search-bar">
          <Input.Search
            placeholder="搜索插件..."
            value={searchForm.keyword}
            onChange={(e) => setSearchForm({ ...searchForm, keyword: e.target.value })}
            onSearch={handleSearch}
            size="large"
            prefix={<SearchOutlined />}
            enterButton="搜索"
          />
        </div>
        
        <div className="search-filters">
          <Select
            value={searchForm.category}
            onChange={(value) => handleFilterChange('category', value)}
            style={{ width: 120 }}
            placeholder="分类"
          >
            {categories.map(cat => (
              <Option key={cat.value} value={cat.value}>{cat.label}</Option>
            ))}
          </Select>
          
          <Select
            value={searchForm.sortBy}
            onChange={(value) => handleFilterChange('sortBy', value)}
            style={{ width: 120 }}
          >
            {sortOptions.map(option => (
              <Option key={option.value} value={option.value}>{option.label}</Option>
            ))}
          </Select>
        </div>
      </div>

      <div className="search-results">
        {loading ? (
          <div className="loading-container">
            <Spin size="large" />
            <span className="loading-text">搜索中...</span>
          </div>
        ) : (plugins && plugins.length > 0) ? (
          <>
            <div className="results-info">
              <span>找到 {total} 个插件</span>
            </div>
            
            <Row gutter={[16, 16]}>
              {plugins.map(plugin => (
                <Col key={plugin.id} xs={24} sm={12} md={8} lg={6}>
                  <Card
                    hoverable
                    cover={
                      <img
                        alt={plugin.name}
                        src={plugin.icon || '/placeholder-plugin.png'}
                        style={{ height: 120, objectFit: 'cover' }}
                      />
                    }
                    onClick={() => window.open(`/plugin/${plugin.id}`, '_blank')}
                  >
                    <Card.Meta
                      title={plugin.name}
                      description={
                        <div>
                          <div className="plugin-description">
                            {plugin.description}
                          </div>
                          <div className="plugin-meta">
                            <span>下载: {plugin.downloadCount}</span>
                            <span>评分: {plugin.rating}/5</span>
                          </div>
                        </div>
                      }
                    />
                  </Card>
                </Col>
              ))}
            </Row>
            
            <div className="pagination-container">
              <Pagination
                current={searchForm.page}
                total={total}
                pageSize={searchForm.pageSize}
                showSizeChanger
                showQuickJumper
                showTotal={(total, range) => `第 ${range[0]}-${range[1]} 项，共 ${total} 项`}
                onChange={handlePageChange}
                onShowSizeChange={handlePageChange}
              />
            </div>
          </>
        ) : (
          <Empty
            description="没有找到相关插件"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        )}
      </div>
    </div>
  );
};

export default Search;