import React from 'react';
import { Avatar, Dropdown, Space, Button } from 'antd';
import { UserOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { useAuth } from '../contexts/AuthContext';

export default function UserMenu() {
  const { user, logout } = useAuth();

  if (!user) return null;

  const handleLogout = () => {
    logout();
  };

  const items: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
      disabled: true, // 暂时禁用，后续可以添加个人信息页面
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      disabled: true, // 暂时禁用，后续可以添加设置页面
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  return (
    <Dropdown
      menu={{ items }}
      placement="bottomRight"
      arrow
    >
      <Button type="text" style={{ height: 'auto', padding: '8px 12px' }}>
        <Space>
          <Avatar
            size="small"
            icon={<UserOutlined />}
            style={{ backgroundColor: '#1890ff' }}
          />
          <span style={{ color: '#fff' }}>
            {user.username}
            {user.role === 'admin' && (
              <span style={{ 
                marginLeft: '4px', 
                fontSize: '12px', 
                opacity: 0.8,
                background: 'rgba(255,255,255,0.2)',
                padding: '1px 4px',
                borderRadius: '2px'
              }}>
                管理员
              </span>
            )}
          </span>
        </Space>
      </Button>
    </Dropdown>
  );
}