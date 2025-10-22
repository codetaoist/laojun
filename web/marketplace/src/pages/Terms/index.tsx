import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Typography,
  Card,
  Divider,
  Breadcrumb,
  Anchor,
  Row,
  Col,
  Alert,
  Space,
  Button,
  Tag,
} from 'antd';
import {
  HomeOutlined,
  FileTextOutlined,
  MailOutlined,
  PhoneOutlined,
  CalendarOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const Terms: React.FC = () => {
  const navigate = useNavigate();

  const lastUpdated = '2024年1月15日';
  const effectiveDate = '2024年1月1日';

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* 面包屑导航 */}
      <Breadcrumb style={{ marginBottom: '24px' }}>
        <Breadcrumb.Item>
          <HomeOutlined />
          <span onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
            首页
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <FileTextOutlined />
          服务条款
        </Breadcrumb.Item>
      </Breadcrumb>

      <Row gutter={24}>
        {/* 侧边导航 */}
        <Col xs={24} lg={6}>
          <Card style={{ position: 'sticky', top: '24px' }}>
            <Anchor
              affix={false}
              offsetTop={80}
              items={[
                {
                  key: 'acceptance',
                  href: '#acceptance',
                  title: '条款接受',
                },
                {
                  key: 'services',
                  href: '#services',
                  title: '服务描述',
                },
                {
                  key: 'accounts',
                  href: '#accounts',
                  title: '用户账户',
                },
                {
                  key: 'conduct',
                  href: '#conduct',
                  title: '用户行为',
                },
                {
                  key: 'content',
                  href: '#content',
                  title: '内容政策',
                },
                {
                  key: 'payment',
                  href: '#payment',
                  title: '付费服务',
                },
                {
                  key: 'intellectual',
                  href: '#intellectual',
                  title: '知识产权',
                },
                {
                  key: 'privacy',
                  href: '#privacy',
                  title: '隐私保护',
                },
                {
                  key: 'disclaimer',
                  href: '#disclaimer',
                  title: '免责声明',
                },
                {
                  key: 'limitation',
                  href: '#limitation',
                  title: '责任限制',
                },
                {
                  key: 'termination',
                  href: '#termination',
                  title: '服务终止',
                },
                {
                  key: 'governing',
                  href: '#governing',
                  title: '适用法律',
                },
                {
                  key: 'changes',
                  href: '#changes',
                  title: '条款变更',
                },
                {
                  key: 'contact',
                  href: '#contact',
                  title: '联系我们',
                },
              ]}
            />
          </Card>
        </Col>

        {/* 主要内容 */}
        <Col xs={24} lg={18}>
          <Card>
            {/* 页面标题 */}
            <div style={{ textAlign: 'center', marginBottom: '32px' }}>
              <Title level={1}>
                <FileTextOutlined style={{ marginRight: '12px', color: '#1890ff' }} />
                服务条款
              </Title>
              <Space direction="vertical">
                <Space>
                  <CalendarOutlined />
                  <Text type="secondary">最后更新: {lastUpdated}</Text>
                </Space>
                <Space>
                  <Tag color="green">生效日期: {effectiveDate}</Tag>
                </Space>
              </Space>
            </div>

            <Alert
              message="重要提示"
              description="请仔细阅读本服务条款。使用我们的服务即表示您同意遵守这些条款。如果您不同意这些条款，请不要使用我们的服务。"
              type="warning"
              showIcon
              style={{ marginBottom: '32px' }}
            />

            {/* 条款接受 */}
            <div id="acceptance">
              <Title level={2}>1. 条款接受</Title>
              <Paragraph>
                欢迎使用插件市场平台（以下简称"本平台"、"我们"或"服务"）。
                本服务条款（以下简称"条款"）构成您与本平台之间具有法律约束力的协议。
              </Paragraph>
              <Paragraph>
                通过访问或使用我们的服务，您确认：
              </Paragraph>
              <ul>
                <li>您已阅读、理解并同意受本条款约束</li>
                <li>您具有签订本协议的法律能力</li>
                <li>如果您代表组织使用服务，您有权代表该组织接受这些条款</li>
              </ul>
            </div>

            <Divider />

            {/* 服务描述 */}
            <div id="services">
              <Title level={2}>2. 服务描述</Title>
              <Paragraph>
                本平台是一个插件市场，为用户提供以下服务：
              </Paragraph>
              <ul>
                <li><strong>插件浏览：</strong>浏览、搜索和发现各类插件</li>
                <li><strong>插件下载：</strong>下载免费和付费插件</li>
                <li><strong>插件管理：</strong>管理已安装的插件</li>
                <li><strong>开发者服务：</strong>为开发者提供插件发布和管理工具</li>
                <li><strong>社区功能：</strong>用户评价、评论和反馈</li>
                <li><strong>支付服务：</strong>安全的在线支付处理</li>
              </ul>
              <Paragraph>
                我们保留随时修改、暂停或终止任何服务功能的权利，恕不另行通知。
              </Paragraph>
            </div>

            <Divider />

            {/* 用户账户 */}
            <div id="accounts">
              <Title level={2}>3. 用户账户</Title>
              
              <Title level={3}>3.1 账户注册</Title>
              <Paragraph>
                要使用某些服务功能，您需要创建账户。注册时，您必须：
              </Paragraph>
              <ul>
                <li>提供准确、完整和最新的信息</li>
                <li>选择安全的密码</li>
                <li>及时更新账户信息</li>
                <li>对账户活动承担责任</li>
              </ul>

              <Title level={3}>3.2 账户安全</Title>
              <Paragraph>
                您有责任：
              </Paragraph>
              <ul>
                <li>保护账户凭据的机密性</li>
                <li>立即通知我们任何未经授权的使用</li>
                <li>确保账户信息的准确性</li>
                <li>遵守我们的安全政策</li>
              </ul>

              <Title level={3}>3.3 账户限制</Title>
              <Paragraph>
                每个用户只能创建一个账户。我们保留暂停或终止违反条款的账户的权利。
              </Paragraph>
            </div>

            <Divider />

            {/* 用户行为 */}
            <div id="conduct">
              <Title level={2}>4. 用户行为规范</Title>
              <Paragraph>
                使用我们的服务时，您同意不会：
              </Paragraph>
              
              <Title level={3}>4.1 禁止行为</Title>
              <ul>
                <li>违反任何适用的法律法规</li>
                <li>侵犯他人的知识产权</li>
                <li>上传恶意软件或病毒</li>
                <li>进行欺诈或误导性活动</li>
                <li>骚扰、威胁或诽谤他人</li>
                <li>干扰或破坏服务的正常运行</li>
                <li>尝试未经授权访问系统</li>
                <li>创建虚假账户或身份</li>
              </ul>

              <Title level={3}>4.2 内容标准</Title>
              <Paragraph>
                您提交的所有内容必须：
              </Paragraph>
              <ul>
                <li>准确且不误导</li>
                <li>不包含非法或有害内容</li>
                <li>尊重他人的权利和尊严</li>
                <li>符合社区准则</li>
              </ul>
            </div>

            <Divider />

            {/* 内容政策 */}
            <div id="content">
              <Title level={2}>5. 内容政策</Title>
              
              <Title level={3}>5.1 用户生成内容</Title>
              <Paragraph>
                您对提交到平台的内容负责，包括但不限于：
              </Paragraph>
              <ul>
                <li>插件代码和文档</li>
                <li>评论和评价</li>
                <li>个人资料信息</li>
                <li>支持请求和反馈</li>
              </ul>

              <Title level={3}>5.2 内容审核</Title>
              <Paragraph>
                我们保留审核、编辑或删除任何内容的权利，特别是：
              </Paragraph>
              <ul>
                <li>违反法律法规的内容</li>
                <li>侵犯知识产权的内容</li>
                <li>包含恶意代码的插件</li>
                <li>不当或有害的内容</li>
              </ul>

              <Title level={3}>5.3 内容许可</Title>
              <Paragraph>
                通过提交内容，您授予我们非独占、全球性、免版税的许可，
                以使用、复制、修改、分发和展示该内容，仅限于提供和改进服务。
              </Paragraph>
            </div>

            <Divider />

            {/* 付费服务 */}
            <div id="payment">
              <Title level={2}>6. 付费服务</Title>
              
              <Title level={3}>6.1 定价和付款</Title>
              <ul>
                <li>所有价格均以人民币显示，除非另有说明</li>
                <li>付款通过第三方支付处理商处理</li>
                <li>您负责支付所有适用的税费</li>
                <li>价格可能随时变更，恕不另行通知</li>
              </ul>

              <Title level={3}>6.2 退款政策</Title>
              <Paragraph>
                我们提供有限的退款政策：
              </Paragraph>
              <ul>
                <li>技术问题导致的无法使用：全额退款</li>
                <li>购买后14天内且未下载：可申请退款</li>
                <li>开发者主动下架的插件：按比例退款</li>
                <li>违规使用不予退款</li>
              </ul>

              <Title level={3}>6.3 订阅服务</Title>
              <Paragraph>
                对于订阅服务：
              </Paragraph>
              <ul>
                <li>自动续费，除非您取消订阅</li>
                <li>可随时取消，在当前计费周期结束时生效</li>
                <li>取消后仍可使用至期满</li>
              </ul>
            </div>

            <Divider />

            {/* 知识产权 */}
            <div id="intellectual">
              <Title level={2}>7. 知识产权</Title>
              
              <Title level={3}>7.1 平台知识产权</Title>
              <Paragraph>
                本平台及其内容（不包括用户生成内容）的所有知识产权归我们所有，包括：
              </Paragraph>
              <ul>
                <li>网站设计和布局</li>
                <li>软件代码和算法</li>
                <li>商标和标识</li>
                <li>文档和帮助内容</li>
              </ul>

              <Title level={3}>7.2 插件知识产权</Title>
              <Paragraph>
                插件的知识产权归其开发者所有。通过平台分发插件，
                开发者授予用户根据插件许可证使用插件的权利。
              </Paragraph>

              <Title level={3}>7.3 侵权处理</Title>
              <Paragraph>
                如果您认为有内容侵犯了您的知识产权，请联系我们。
                我们将根据适用法律处理侵权投诉。
              </Paragraph>
            </div>

            <Divider />

            {/* 隐私保护 */}
            <div id="privacy">
              <Title level={2}>8. 隐私保护</Title>
              <Paragraph>
                您的隐私对我们很重要。我们的隐私政策详细说明了我们如何收集、
                使用和保护您的个人信息。使用我们的服务即表示您同意我们的隐私政策。
              </Paragraph>
              <Paragraph>
                <Button type="link" onClick={() => navigate('/privacy')}>
                  查看完整隐私政策
                </Button>
              </Paragraph>
            </div>

            <Divider />

            {/* 免责声明 */}
            <div id="disclaimer">
              <Title level={2}>9. 免责声明</Title>
              <Alert
                message="重要声明"
                description="本服务按「现状」提供，我们不提供任何明示或暗示的保证。"
                type="warning"
                showIcon
                style={{ marginBottom: '16px' }}
              />
              <Paragraph>
                我们明确声明不对以下内容承担责任：
              </Paragraph>
              <ul>
                <li>服务的可用性、可靠性或及时性</li>
                <li>第三方插件的质量、安全性或功能</li>
                <li>用户生成内容的准确性或合法性</li>
                <li>因使用服务而导致的任何损失或损害</li>
                <li>第三方网站或服务的内容或行为</li>
              </ul>
            </div>

            <Divider />

            {/* 责任限制 */}
            <div id="limitation">
              <Title level={2}>10. 责任限制</Title>
              <Paragraph>
                在法律允许的最大范围内，我们的总责任不超过：
              </Paragraph>
              <ul>
                <li>您在过去12个月内支付给我们的费用总额</li>
                <li>或人民币1000元，以较高者为准</li>
              </ul>
              <Paragraph>
                我们不对以下损失承担责任：
              </Paragraph>
              <ul>
                <li>间接、特殊、偶然或后果性损失</li>
                <li>利润损失或业务中断</li>
                <li>数据丢失或损坏</li>
                <li>声誉损害</li>
              </ul>
            </div>

            <Divider />

            {/* 服务终止 */}
            <div id="termination">
              <Title level={2}>11. 服务终止</Title>
              
              <Title level={3}>11.1 用户终止</Title>
              <Paragraph>
                您可以随时停止使用我们的服务并删除您的账户。
              </Paragraph>

              <Title level={3}>11.2 我们的终止权</Title>
              <Paragraph>
                我们可能在以下情况下暂停或终止您的账户：
              </Paragraph>
              <ul>
                <li>违反本条款</li>
                <li>涉嫌欺诈或非法活动</li>
                <li>长期不活跃</li>
                <li>技术或安全原因</li>
              </ul>

              <Title level={3}>11.3 终止后果</Title>
              <Paragraph>
                账户终止后：
              </Paragraph>
              <ul>
                <li>您将失去对服务的访问权限</li>
                <li>您的数据可能被删除</li>
                <li>未使用的付费服务可能不予退款</li>
                <li>本条款的相关条款仍然有效</li>
              </ul>
            </div>

            <Divider />

            {/* 适用法律 */}
            <div id="governing">
              <Title level={2}>12. 适用法律和争议解决</Title>
              
              <Title level={3}>12.1 适用法律</Title>
              <Paragraph>
                本条款受中华人民共和国法律管辖，不考虑法律冲突原则。
              </Paragraph>

              <Title level={3}>12.2 争议解决</Title>
              <Paragraph>
                因本条款引起的任何争议，双方应首先通过友好协商解决。
                协商不成的，应提交至上海市有管辖权的人民法院解决。
              </Paragraph>
            </div>

            <Divider />

            {/* 条款变更 */}
            <div id="changes">
              <Title level={2}>13. 条款变更</Title>
              <Paragraph>
                我们可能会不时更新本条款。重大变更时，我们会：
              </Paragraph>
              <ul>
                <li>在网站上发布显著通知</li>
                <li>向注册用户发送邮件通知</li>
                <li>提供至少30天的通知期</li>
              </ul>
              <Paragraph>
                继续使用服务即表示您接受更新后的条款。
                如果您不同意新条款，请停止使用服务。
              </Paragraph>
            </div>

            <Divider />

            {/* 联系我们 */}
            <div id="contact">
              <Title level={2}>14. 联系我们</Title>
              <Paragraph>
                如果您对本服务条款有任何疑问，请联系我们：
              </Paragraph>
              
              <Card style={{ background: '#f9f9f9', marginTop: '16px' }}>
                <Space direction="vertical" size="middle">
                  <Space>
                    <MailOutlined style={{ color: '#1890ff' }} />
                    <Text><strong>邮箱：</strong>legal@pluginmarket.com</Text>
                  </Space>
                  <Space>
                    <PhoneOutlined style={{ color: '#1890ff' }} />
                    <Text><strong>电话：</strong>+86 400-123-4567</Text>
                  </Space>
                  <Text>
                    <strong>邮寄地址：</strong>中国上海市浦东新区张江高科技园区
                    插件市场平台 法务部门
                  </Text>
                </Space>
              </Card>
            </div>

            {/* 底部操作 */}
            <div style={{ textAlign: 'center', marginTop: '48px', paddingTop: '24px', borderTop: '1px solid #f0f0f0' }}>
              <Space size="large">
                <Button type="primary" onClick={() => navigate('/')}>
                  返回首页
                </Button>
                <Button onClick={() => navigate('/privacy')}>
                  查看隐私政策
                </Button>
                <Button onClick={() => navigate('/about')}>
                  关于我们
                </Button>
              </Space>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Terms;