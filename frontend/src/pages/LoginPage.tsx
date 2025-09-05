import React, { useState } from 'react';
import { Button, Form, Input, Card, Typography, message, Spin } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate, useLocation } from 'react-router-dom';

const { Title } = Typography;

interface LoginFormData {
  username: string;
  password: string;
}

export default function LoginPage() {
  const [form] = Form.useForm();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // 获取重定向路径，默认回到首页
  const from = (location.state as any)?.from?.pathname || '/';

  const handleSubmit = async (values: LoginFormData) => {
    try {
      setIsSubmitting(true);
      await login(values.username, values.password);
      
      // 登录成功后跳转
      navigate(from, { replace: true });
    } catch (error) {
      // 错误已经在AuthContext中处理
      console.error('Login failed:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div style={{ 
        height: '100vh', 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center' 
      }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      padding: '20px'
    }}>
      <Card 
        style={{ 
          width: '100%', 
          maxWidth: 400, 
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
          borderRadius: '12px'
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: '32px' }}>
          <Title level={2} style={{ color: '#1890ff', marginBottom: '8px' }}>
            URL管理系统
          </Title>
          <p style={{ color: '#666', margin: 0 }}>
            欢迎回来，请登录您的账户
          </p>
        </div>

        <Form
          form={form}
          name="login"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名!' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#1890ff' }} />}
              placeholder="用户名"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码!' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#1890ff' }} />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: '16px' }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={isSubmitting}
              block
              style={{ 
                height: '44px',
                borderRadius: '6px',
                fontSize: '16px'
              }}
            >
              登录
            </Button>
          </Form.Item>
        </Form>

        <div style={{ 
          textAlign: 'center', 
          color: '#666', 
          fontSize: '14px',
          borderTop: '1px solid #f0f0f0',
          paddingTop: '16px',
          marginTop: '16px'
        }}>
          <p style={{ margin: 0 }}>
            默认管理员账户：admin / admin123
          </p>
        </div>
      </Card>
    </div>
  );
}