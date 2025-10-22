import React from 'react';
import { Layout, Row, Col, Space, Divider } from 'antd';
import { Link } from 'react-router-dom';
import {
  GithubOutlined,
  TwitterOutlined,
  WechatOutlined,
  MailOutlined,
} from '@ant-design/icons';

const { Footer: AntFooter } = Layout;

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear();

  return (
    <AntFooter className="marketplace-footer">
      <div className="footer-container">
        <Row gutter={[32, 32]}>
          {/* 产品信息 */}
          <Col xs={24} sm={12} md={6}>
            <div className="footer-section">
              <h4>太上老君插件市场</h4>
              <p className="footer-description">
                发现和安装各种功能插件，扩展您的太上老君系统功能。
              </p>
              <Space>
                <a href="#" className="social-link" title="GitHub">
                  <GithubOutlined />
                </a>
                <a href="#" className="social-link" title="Twitter">
                  <TwitterOutlined />
                </a>
                <a href="#" className="social-link" title="微信">
                  <WechatOutlined />
                </a>
                <a href="#" className="social-link" title="邮箱">
                  <MailOutlined />
                </a>
              </Space>
            </div>
          </Col>

          {/* 产品链接 */}
          <Col xs={24} sm={12} md={6}>
            <div className="footer-section">
              <h4>产品</h4>
              <ul className="footer-links">
                <li><Link to="/categories">插件分类</Link></li>
                <li><Link to="/developers">开发者</Link></li>
                <li><Link to="/my-plugins">我的插件</Link></li>
                <li><Link to="/favorites">我的收藏</Link></li>
              </ul>
            </div>
          </Col>

          {/* 开发者资源 */}
          <Col xs={24} sm={12} md={6}>
            <div className="footer-section">
              <h4>开发者</h4>
              <ul className="footer-links">
                <li><a href="#" target="_blank" rel="noopener noreferrer">开发文档</a></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">API 参考</a></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">插件模板</a></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">发布指南</a></li>
              </ul>
            </div>
          </Col>

          {/* 支持与帮助 */}
          <Col xs={24} sm={12} md={6}>
            <div className="footer-section">
              <h4>支持</h4>
              <ul className="footer-links">
                <li><Link to="/about">关于我们</Link></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">帮助中心</a></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">联系我们</a></li>
                <li><a href="#" target="_blank" rel="noopener noreferrer">反馈建议</a></li>
              </ul>
            </div>
          </Col>
        </Row>

        <Divider />

        {/* 底部信息 */}
        <div className="footer-bottom">
          <Row justify="space-between" align="middle">
            <Col xs={24} md={12}>
              <div className="copyright">
                © {currentYear} 太上老君插件市场. 保留所有权利.
              </div>
            </Col>
            <Col xs={24} md={12}>
              <div className="footer-legal">
                <Space split={<span className="separator">|</span>}>
                  <Link to="/privacy">隐私政策</Link>
                  <Link to="/terms">服务条款</Link>
                  <a href="#" target="_blank" rel="noopener noreferrer">
                    许可协议
                  </a>
                </Space>
              </div>
            </Col>
          </Row>
        </div>
      </div>
    </AntFooter>
  );
};

export default Footer;