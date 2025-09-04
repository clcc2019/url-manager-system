import React, { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { CreateProjectRequest } from '../types/api';
import { ApiService } from '../services/api';

const CreateProject: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (values: CreateProjectRequest) => {
    setLoading(true);
    try {
      await ApiService.createProject(values);
      message.success('项目创建成功');
      navigate('/projects');
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || '创建失败';
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title="创建项目" style={{ maxWidth: 600 }}>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
      >
        <Form.Item
          label="项目名称"
          name="name"
          rules={[
            { required: true, message: '请输入项目名称' },
            { min: 1, max: 100, message: '项目名称长度为1-100个字符' },
          ]}
        >
          <Input placeholder="请输入项目名称" />
        </Form.Item>

        <Form.Item
          label="项目描述"
          name="description"
        >
          <Input.TextArea 
            rows={4} 
            placeholder="请输入项目描述（可选）"
          />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            创建项目
          </Button>
          <Button 
            style={{ marginLeft: 8 }} 
            onClick={() => navigate('/projects')}
          >
            取消
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default CreateProject;