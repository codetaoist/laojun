import React from 'react';
import { Outlet } from 'react-router-dom';
import { Layout as AntLayout, FloatButton } from 'antd';
import { UpOutlined } from '@ant-design/icons';
import Header from './Header';
import Footer from './Footer';
import './index.css';

const { Content } = AntLayout;

const Layout: React.FC = () => {
  return (
    <AntLayout className="marketplace-layout">
      <Header />
      
      <Content className="marketplace-content">
        <div className="content-wrapper">
          <Outlet />
        </div>
      </Content>
      
      <Footer />
      
      <FloatButton.BackTop>
        <div className="back-top-button">
          <UpOutlined />
        </div>
      </FloatButton.BackTop>
    </AntLayout>
  );
};

export default Layout;