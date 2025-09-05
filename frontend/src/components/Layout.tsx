import React from 'react';
import { Layout as AntLayout, Menu, Typography, Breadcrumb } from 'antd';
import { ProjectOutlined, LinkOutlined } from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';

const { Header, Content, Sider } = AntLayout;
const { Title } = Typography;

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/projects',
      icon: <ProjectOutlined />,
      label: '项目管理',
      onClick: () => navigate('/projects'),
    },
  ];

  const getBreadcrumbItems = () => {
    const pathSegments = location.pathname.split('/').filter(Boolean);
    const items = [{ title: '首页' }];

    if (pathSegments[0] === 'projects') {
      items.push({ title: '项目管理' });
      
      if (pathSegments[1] === 'new') {
        items.push({ title: '创建项目' });
      } else if (pathSegments[1]) {
        items.push({ title: '项目详情' });
      }
    }

    return items;
  };

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header style={{ padding: '0 24px', background: '#fff', boxShadow: '0 1px 4px rgba(0,21,41,.08)' }}>
        <div style={{ display: 'flex', alignItems: 'center', height: '100%' }}>
          <LinkOutlined style={{ fontSize: '24px', marginRight: '12px', color: '#1890ff' }} />
          <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
            URL管理系统
          </Title>
        </div>
      </Header>
      
      <AntLayout>
        <Sider 
          width={256} 
          style={{ background: '#fff' }}
          breakpoint="lg"
          collapsedWidth="0"
          theme="light"
        >
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            style={{ height: '100%', borderRight: 0 }}
            items={menuItems}
          />
        </Sider>
        
        <AntLayout style={{ padding: '24px' }}>
          <Breadcrumb style={{ marginBottom: '24px' }} items={getBreadcrumbItems()} />
          
          <Content
            style={{
              padding: '24px',
              margin: 0,
              minHeight: 280,
              background: '#fff',
              borderRadius: '6px',
            }}
          >
            {children}
          </Content>
        </AntLayout>
      </AntLayout>
    </AntLayout>
  );
};

export default Layout;