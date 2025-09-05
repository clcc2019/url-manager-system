import React, { useState, useEffect } from 'react';
import { Table, Button, message, Popconfirm, Typography, Space, Modal, Form, Input, Tag, Spin } from 'antd';
import { PlusOutlined, DeleteOutlined, EyeOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { Project, CreateProjectRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useAuth } from '../contexts/AuthContext';
import { useRoleCheck } from '../components/RoleGuard';

const { Title } = Typography;

const ProjectList: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const { user, isLoading: authLoading } = useAuth();
  const { isAdmin } = useRoleCheck();

  // 如果认证状态还在加载中，显示加载指示器
  if (authLoading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        minHeight: '400px' 
      }}>
        <Spin size="large" tip="加载用户信息..." />
      </div>
    );
  }

  const fetchProjects = async (page = 1, size = 10) => {
    setLoading(true);
    try {
      const response = await ApiService.getProjects({
        limit: size,
        offset: (page - 1) * size,
      });
      setProjects(response.projects);
      setTotal(response.total);
    } catch {
      message.error('获取项目列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // 只有在认证状态初始化完成后才加载项目列表
    if (!authLoading) {
      fetchProjects(current, pageSize);
    }
  }, [current, pageSize, authLoading]);

  const handleDelete = async (id: string) => {
    try {
      await ApiService.deleteProject(id);
      message.success('删除成功');
      fetchProjects(current, pageSize);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '删除失败';
      message.error(errorMsg);
    }
  };

  const handleCreateProject = async (values: CreateProjectRequest) => {
    setCreateLoading(true);
    try {
      await ApiService.createProject(values);
      message.success('项目创建成功');
      setCreateModalVisible(false);
      form.resetFields();
      fetchProjects(current, pageSize);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '创建失败';
      message.error(errorMsg);
    } finally {
      setCreateLoading(false);
    }
  };

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Project) => (
        <Button 
          type="link" 
          onClick={() => navigate(`/projects/${record.id}`)}
          style={{ padding: 0, height: 'auto' }}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text: string) => text || '-',
    },
    // 管理员可以看到项目所有者
    ...(isAdmin ? [{
      title: '所有者',
      dataIndex: 'user_id',
      key: 'owner',
      width: 120,
      render: (userId: string, record: Project) => {
        const isOwner = userId && userId === user?.id;
        return (
          <Tag 
            color={isOwner ? 'blue' : 'default'}
            icon={<UserOutlined />}
          >
            {isOwner ? '你' : userId ? `用户${userId.slice(-8)}` : '未知用户'}
          </Tag>
        );
      },
    }] : []),
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => formatDate(text),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (text: string) => formatDate(text),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_: any, record: Project) => {
        // 管理员始终有权限，或者是项目所有者
        const canModify = isAdmin || (record.user_id && record.user_id === user?.id);
        
        return (
          <Space>
            <Button
              type="primary"
              icon={<EyeOutlined />}
              size="small"
              onClick={() => navigate(`/projects/${record.id}`)}
            >
              查看
            </Button>
            {canModify && (
              <Popconfirm
                title="确定要删除这个项目吗？"
                description="删除后无法恢复，且项目下不能有活跃的URL"
                onConfirm={() => handleDelete(record.id)}
                okText="确定"
                cancelText="取消"
              >
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  size="small"
                >
                  删除
                </Button>
              </Popconfirm>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>项目管理</Title>
          {isAdmin && (
            <Typography.Text type="secondary" style={{ marginTop: 4, display: 'block' }}>
              作为管理员，您可以查看和管理所有用户的项目
            </Typography.Text>
          )}
        </div>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建项目
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={projects}
        rowKey="id"
        loading={loading}
        pagination={{
          current,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          onChange: (page, size) => {
            setCurrent(page);
            setPageSize(size || 10);
          },
        }}
      />

      <Modal
        title="创建项目"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateProject}
        >
          <Form.Item
            label="项目名称"
            name="name"
            rules={[
              { required: true, message: '请输入项目名称' },
              { min: 2, message: '项目名称至少2个字符' },
              { max: 50, message: '项目名称不能超过50个字符' }
            ]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>

          <Form.Item
            label="项目描述"
            name="description"
            rules={[
              { max: 200, message: '描述不能超过200个字符' }
            ]}
          >
            <Input.TextArea 
              placeholder="请输入项目描述（可选）" 
              rows={4}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={createLoading}
              >
                创建项目
              </Button>
              <Button 
                onClick={() => {
                  setCreateModalVisible(false);
                  form.resetFields();
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ProjectList;