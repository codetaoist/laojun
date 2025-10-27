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
} from 'antd';
import {
  HomeOutlined,
  SafetyOutlined,
  MailOutlined,
  PhoneOutlined,
  CalendarOutlined,
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;
const { Link } = Anchor;

const Privacy: React.FC = () => {
  const navigate = useNavigate();

  const lastUpdated = '2024年1月15日';

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
-          <ShieldOutlined />
+          <SafetyOutlined />
          隐私政策
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
                  key: 'overview',
                  href: '#overview',
                  title: '概述',
                },
                {
                  key: 'collection',
                  href: '#collection',
                  title: '信息收集',
                },
                {
                  key: 'usage',
                  href: '#usage',
                  title: '信息使用',
                },
                {
                  key: 'sharing',
                  href: '#sharing',
                  title: '信息共享',
                },
                {
                  key: 'security',
                  href: '#security',
                  title: '信息安全',
                },
                {
                  key: 'cookies',
                  href: '#cookies',
                  title: 'Cookie政策',
                },
                {
                  key: 'rights',
                  href: '#rights',
                  title: '用户权利',
                },
                {
                  key: 'children',
                  href: '#children',
                  title: '儿童隐私',
                },
                {
                  key: 'changes',
                  href: '#changes',
                  title: '政策变更',
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
-                <ShieldOutlined style={{ marginRight: '12px', color: '#52c41a' }} />
+                <SafetyOutlined style={{ marginRight: '12px', color: '#52c41a' }} />
                隐私政策
              </Title>
              <Space>
                <CalendarOutlined />
                <Text type="secondary">最后更新: {lastUpdated}</Text>
              </Space>
            </div>

            <Alert
              message="重要提示"
              description="请仔细阅读本隐私政策。使用我们的服务即表示您同意本政策中描述的信息收集和使用方式。"
              type="info"
              showIcon
              style={{ marginBottom: '32px' }}
            />

            {/* 概述 */}
            <div id="overview">
              <Title level={2}>1. 概述</Title>
              <Paragraph>
                插件市场平台（以下简称"我们"或"本平台"）非常重视用户的隐私保护。
                本隐私政策说明了我们如何收集、使用、存储和保护您的个人信息。
              </Paragraph>
              <Paragraph>
                本政策适用于您使用我们的网站、移动应用程序和相关服务时的信息处理活动。
                我们承诺按照适用的隐私法律法规处理您的个人信息。
              </Paragraph>
            </div>

            <Divider />

            {/* 信息收集 */}
            <div id="collection">
              <Title level={2}>2. 我们收集的信息</Title>
              
              <Title level={3}>2.1 您主动提供的信息</Title>
              <Paragraph>
                当您注册账户、使用我们的服务或与我们联系时，我们可能收集以下信息：
              </Paragraph>
              <ul>
                <li>个人身份信息：姓名、邮箱地址、电话号码</li>
                <li>账户信息：用户名、密码、个人资料</li>
                <li>支付信息：信用卡信息、账单地址（通过第三方支付处理商处理）</li>
                <li>通信信息：您发送给我们的消息、反馈和评论</li>
              </ul>

              <Title level={3}>2.2 自动收集的信息</Title>
              <Paragraph>
                当您使用我们的服务时，我们可能自动收集以下技术信息：
              </Paragraph>
              <ul>
                <li>设备信息：设备类型、操作系统、浏览器类型和版本</li>
                <li>使用信息：访问时间、页面浏览记录、点击行为</li>
                <li>网络信息：IP地址、网络连接类型</li>
                <li>位置信息：基于IP地址的大致地理位置</li>
              </ul>

              <Title level={3}>2.3 第三方信息</Title>
              <Paragraph>
                我们可能从第三方合作伙伴处获得关于您的信息，包括：
              </Paragraph>
              <ul>
                <li>社交媒体平台的公开信息（如果您选择关联账户）</li>
                <li>分析服务提供商的使用统计信息</li>
                <li>广告合作伙伴的兴趣和行为数据</li>
              </ul>
            </div>

            <Divider />

            {/* 信息使用 */}
            <div id="usage">
              <Title level={2}>3. 信息使用方式</Title>
              <Paragraph>
                我们使用收集的信息来：
              </Paragraph>
              <ul>
                <li><strong>提供服务：</strong>创建和管理您的账户，处理交易，提供客户支持</li>
                <li><strong>改进服务：</strong>分析使用模式，优化用户体验，开发新功能</li>
                <li><strong>个性化体验：</strong>推荐相关插件，定制内容和广告</li>
                <li><strong>安全保护：</strong>检测和防止欺诈、滥用和安全威胁</li>
                <li><strong>法律合规：</strong>遵守法律义务，保护我们的权利和利益</li>
                <li><strong>营销推广：</strong>发送产品更新、促销信息（您可以选择退订）</li>
              </ul>
            </div>

            <Divider />

            {/* 信息共享 */}
            <div id="sharing">
              <Title level={2}>4. 信息共享</Title>
              <Paragraph>
                我们不会出售您的个人信息。我们可能在以下情况下共享您的信息：
              </Paragraph>
              
              <Title level={3}>4.1 服务提供商</Title>
              <Paragraph>
                我们可能与第三方服务提供商共享信息，包括：
              </Paragraph>
              <ul>
                <li>云存储和托管服务提供商</li>
                <li>支付处理服务提供商</li>
                <li>分析和营销服务提供商</li>
                <li>客户支持服务提供商</li>
              </ul>

              <Title level={3}>4.2 法律要求</Title>
              <Paragraph>
                在以下情况下，我们可能披露您的信息：
              </Paragraph>
              <ul>
                <li>遵守法律义务或法院命令</li>
                <li>保护我们的权利、财产或安全</li>
                <li>防止或调查可能的违法行为</li>
                <li>保护用户或公众的安全</li>
              </ul>

              <Title level={3}>4.3 业务转让</Title>
              <Paragraph>
                如果我们参与合并、收购或资产出售，您的信息可能会作为业务资产的一部分被转让。
                我们会在此类转让发生前通知您。
              </Paragraph>
            </div>

            <Divider />

            {/* 信息安全 */}
            <div id="security">
              <Title level={2}>5. 信息安全</Title>
              <Paragraph>
                我们采用多种安全措施来保护您的个人信息：
              </Paragraph>
              <ul>
                <li><strong>加密传输：</strong>使用SSL/TLS加密保护数据传输</li>
                <li><strong>访问控制：</strong>限制员工对个人信息的访问权限</li>
                <li><strong>安全存储：</strong>使用安全的数据中心和存储系统</li>
                <li><strong>定期审计：</strong>定期进行安全评估和漏洞测试</li>
                <li><strong>事件响应：</strong>建立数据泄露应急响应机制</li>
              </ul>
              <Paragraph>
                尽管我们采取了合理的安全措施，但请注意，没有任何系统是100%安全的。
                我们建议您采取适当的预防措施保护您的账户安全。
              </Paragraph>
            </div>

            <Divider />

            {/* Cookie政策 */}
            <div id="cookies">
              <Title level={2}>6. Cookie和类似技术</Title>
              <Paragraph>
                我们使用Cookie和类似技术来改善您的体验：
              </Paragraph>
              
              <Title level={3}>6.1 Cookie类型</Title>
              <ul>
                <li><strong>必要Cookie：</strong>确保网站正常运行的基本功能</li>
                <li><strong>性能Cookie：</strong>收集网站使用统计信息</li>
                <li><strong>功能Cookie：</strong>记住您的偏好设置</li>
                <li><strong>广告Cookie：</strong>提供个性化广告内容</li>
              </ul>

              <Title level={3}>6.2 Cookie管理</Title>
              <Paragraph>
                您可以通过浏览器设置管理Cookie偏好。请注意，禁用某些Cookie可能影响网站功能。
              </Paragraph>
            </div>

            <Divider />

            {/* 用户权利 */}
            <div id="rights">
              <Title level={2}>7. 您的权利</Title>
              <Paragraph>
                根据适用的隐私法律，您享有以下权利：
              </Paragraph>
              <ul>
                <li><strong>访问权：</strong>请求查看我们持有的关于您的个人信息</li>
                <li><strong>更正权：</strong>请求更正不准确或不完整的信息</li>
                <li><strong>删除权：</strong>请求删除您的个人信息</li>
                <li><strong>限制权：</strong>请求限制对您信息的处理</li>
                <li><strong>可携权：</strong>请求以结构化格式获得您的数据</li>
                <li><strong>反对权：</strong>反对基于合法利益的数据处理</li>
                <li><strong>撤回同意：</strong>撤回您之前给予的同意</li>
              </ul>
              <Paragraph>
                要行使这些权利，请通过下方联系方式与我们联系。我们将在法律规定的时间内回应您的请求。
              </Paragraph>
            </div>

            <Divider />

            {/* 儿童隐私 */}
            <div id="children">
              <Title level={2}>8. 儿童隐私保护</Title>
              <Paragraph>
                我们的服务不面向13岁以下的儿童。我们不会故意收集13岁以下儿童的个人信息。
                如果我们发现收集了儿童的个人信息，我们将立即删除这些信息。
              </Paragraph>
              <Paragraph>
                如果您是父母或监护人，发现您的孩子向我们提供了个人信息，请联系我们，
                我们将采取措施删除这些信息。
              </Paragraph>
            </div>

            <Divider />

            {/* 政策变更 */}
            <div id="changes">
              <Title level={2}>9. 隐私政策变更</Title>
              <Paragraph>
                我们可能会不时更新本隐私政策。重大变更时，我们会通过以下方式通知您：
              </Paragraph>
              <ul>
                <li>在网站上发布显著通知</li>
                <li>向您的注册邮箱发送通知</li>
                <li>通过应用内通知告知您</li>
              </ul>
              <Paragraph>
                我们建议您定期查看本政策以了解最新信息。继续使用我们的服务即表示您接受更新后的政策。
              </Paragraph>
            </div>

            <Divider />

            {/* 联系我们 */}
            <div id="contact">
              <Title level={2}>10. 联系我们</Title>
              <Paragraph>
                如果您对本隐私政策有任何疑问或需要行使您的权利，请通过以下方式联系我们：
              </Paragraph>
              
              <Card style={{ background: '#f9f9f9', marginTop: '16px' }}>
                <Space direction="vertical" size="middle">
                  <Space>
                    <MailOutlined style={{ color: '#1890ff' }} />
                    <Text><strong>邮箱：</strong>privacy@pluginmarket.com</Text>
                  </Space>
                  <Space>
                    <PhoneOutlined style={{ color: '#1890ff' }} />
                    <Text><strong>电话：</strong>+86 400-123-4567</Text>
                  </Space>
                  <Text>
                    <strong>邮寄地址：</strong>中国上海市浦东新区张江高科技园区
                    插件市场平台 隐私保护部门
                  </Text>
                </Space>
              </Card>

              <Paragraph style={{ marginTop: '16px' }}>
                我们承诺在收到您的请求后30天内回复。对于复杂的请求，我们可能需要额外的时间，
                但会及时告知您处理进度。
              </Paragraph>
            </div>

            {/* 底部操作 */}
            <div style={{ textAlign: 'center', marginTop: '48px', paddingTop: '24px', borderTop: '1px solid #f0f0f0' }}>
              <Space size="large">
                <Button type="primary" onClick={() => navigate('/')}>
                  返回首页
                </Button>
                <Button onClick={() => navigate('/terms')}>
                  查看服务条款
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

export default Privacy;