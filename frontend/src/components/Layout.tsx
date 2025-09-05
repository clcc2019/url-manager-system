import React, { useState, useEffect } from 'react';
import { Layout as AntLayout, Menu, Typography, Breadcrumb, Button } from 'antd';
import { ProjectOutlined, LinkOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';

const { Header, Content, Sider } = AntLayout;
const { Title } = Typography;

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);

  // 响应式：窗口宽度小于1024px时自动折叠
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < 1024 && !collapsed) {
        setCollapsed(true);
      }
    };

    // 初始化检查
    handleResize();

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [collapsed]);

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
        <div style={{ display: 'flex', alignItems: 'center', height: '100%', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{
                fontSize: '16px',
                width: 32,
                height: 32,
                marginRight: '12px'
              }}
            />
            <LinkOutlined style={{ fontSize: '24px', marginRight: '12px', color: '#1890ff' }} />
            <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
              URL管理系统
            </Title>
          </div>
        </div>
      </Header>
      
      <AntLayout>
        <Sider
          width={256}
          style={{ background: '#fff' }}
          collapsed={collapsed}
          collapsedWidth={0}
          theme="light"
          trigger={null}
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