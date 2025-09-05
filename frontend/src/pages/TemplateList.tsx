import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  message,
  Space,
  Popconfirm,
  Typography,
  Drawer,
  Divider,
  Tag,
  Spin,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  UserOutlined,
} from '@ant-design/icons';
import type { AppTemplate, CreateTemplateRequest, UpdateTemplateRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useAuth } from '../contexts/AuthContext';
import { useRoleCheck } from '../components/RoleGuard';

const { Title, Text } = Typography;
const { TextArea } = Input;

interface TemplateListProps {
  inProject?: boolean;
  onSelect?: (template: AppTemplate) => void;
}

const TemplateList: React.FC<TemplateListProps> = ({ inProject = false, onSelect }) => {
  const [templates, setTemplates] = useState<AppTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const { user, isLoading: authLoading } = useAuth();
  const { isAdmin } = useRoleCheck();

  // 如果认证状态还在加载中且不是在项目内使用，显示加载指示器
  if (authLoading && !inProject) {
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

  // Modal states
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [currentTemplate, setCurrentTemplate] = useState<AppTemplate | null>(null);

  // Forms
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const loadTemplates = useCallback(async (page = 1) => {
    // 只有在认证状态初始化完成后才加载模版列表
    if (authLoading) {
      return;
    }
    
    setLoading(true);
    try {
      const response = await ApiService.getTemplates({
        limit: pageSize,
        offset: (page - 1) * pageSize,
      });
      setTemplates(response.templates);
      setTotal(response.total);
      setCurrentPage(page);
    } catch {
      message.error('获取模版列表失败');
    } finally {
      setLoading(false);
    }
  }, [pageSize, authLoading]);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  const handleCreate = async (values: CreateTemplateRequest) => {
    try {
      await ApiService.createTemplate(values);
      message.success('模版创建成功');
      setCreateModalVisible(false);
      createForm.resetFields();
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版创建失败';
      message.error(errorMsg);
    }
  };

  const handleEdit = (template: AppTemplate) => {
    setCurrentTemplate(template);
    editForm.setFieldsValue(template);
    setEditModalVisible(true);
  };

  const handleUpdate = async (values: UpdateTemplateRequest) => {
    if (!currentTemplate) return;

    try {
      await ApiService.updateTemplate(currentTemplate.id, values);
      message.success('模版更新成功');
      setEditModalVisible(false);
      editForm.resetFields();
      setCurrentTemplate(null);
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版更新失败';
      message.error(errorMsg);
    }
  };

  const handleDelete = async (templateId: string) => {
    try {
      await ApiService.deleteTemplate(templateId);
      message.success('模版删除成功');
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版删除失败';
      message.error(errorMsg);
    }
  };

  const handlePreview = (template: AppTemplate) => {
    setCurrentTemplate(template);
    setPreviewVisible(true);
  };

  const handleSelect = (template: AppTemplate) => {
    if (onSelect) {
      onSelect(template);
    }
  };

  const columns = [
    {
      title: '模版名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: AppTemplate) => (
        <Space>
          <Text strong>{text}</Text>
          {record.description && (
            <Text type="secondary" style={{ fontSize: '12px' }}>({record.description})</Text>
          )}
        </Space>
      ),
    },
    // 管理员可以看到模版所有者
    ...(isAdmin ? [{
      title: '所有者',
      dataIndex: 'user_id',
      key: 'owner',
      width: 120,
      render: (userId: string, record: AppTemplate) => {
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
      width: 180,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (text: string) => formatDate(text),
      width: 180,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: AppTemplate) => {
        // 管理员始终有权限，或者是模版所有者
        const canModify = isAdmin || (record.user_id && record.user_id === user?.id);
        
        return (
          <Space>
            {inProject ? (
              <Button 
                type="primary" 
                size="small" 
                onClick={() => handleSelect(record)}
              >
                选择
              </Button>
            ) : null}
            <Button
              icon={<EyeOutlined />}
              size="small"
              onClick={() => handlePreview(record)}
              title="预览"
            />
            {canModify && (
              <>
                <Button
                  icon={<EditOutlined />}
                  size="small"
                  onClick={() => handleEdit(record)}
                  title="编辑"
                />
                <Popconfirm
                  title="确定要删除这个模版吗？"
                  description="删除后无法恢复，请确认。"
                  onConfirm={() => handleDelete(record.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    icon={<DeleteOutlined />}
                    size="small"
                    danger
                    title="删除"
                  />
                </Popconfirm>
              </>
            )}
          </Space>
        );
      },
      width: 200,
    },
  ];

  return (
    <div>
      {!inProject && (
        <>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
            <div>
              <Title level={2} style={{ margin: 0 }}>模版管理</Title>
              {isAdmin && (
                <Typography.Text type="secondary" style={{ marginTop: 4, display: 'block' }}>
                  作为管理员，您可以查看和管理所有用户的模版
                </Typography.Text>
              )}
            </div>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              创建模版
            </Button>
          </div>
        </>
      )}
      
      {inProject && (
        <div style={{ marginBottom: 16 }}>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建模版
          </Button>
        </div>
      )}

      <Table
        columns={columns}
        dataSource={templates}
        rowKey="id"
        loading={loading}
        pagination={{
          current: currentPage,
          pageSize: pageSize,
          total: total,
          onChange: loadTemplates,
          showSizeChanger: false,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
        }}
      />

      {/* 创建模版模态框 */}
      <Modal
        title="创建应用模版"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          createForm.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreate}
        >
          <Form.Item
            label="模版名称"
            name="name"
            rules={[
              { required: true, message: '请输入模版名称' },
              { max: 100, message: '模版名称不能超过100个字符' },
            ]}
          >
            <Input placeholder="请输入模版名称" />
          </Form.Item>

          <Form.Item
            label="描述"
            name="description"
          >
            <Input placeholder="请输入模版描述（可选）" />
          </Form.Item>

          <Form.Item
            label="YAML规范"
            name="yaml_spec"
            rules={[
              { required: true, message: '请输入YAML规范' },
            ]}
          >
            <TextArea
              rows={15}
              placeholder="请输入Kubernetes YAML规范，支持 ${PLACEHOLDER} 占位符"
              style={{ fontFamily: 'Monaco, Consolas, monospace' }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
              <Button 
                onClick={() => {
                  setCreateModalVisible(false);
                  createForm.resetFields();
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑模版模态框 */}
      <Modal
        title="编辑应用模版"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          editForm.resetFields();
          setCurrentTemplate(null);
        }}
        footer={null}
        width={800}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdate}
        >
          <Form.Item
            label="模版名称"
            name="name"
            rules={[
              { required: true, message: '请输入模版名称' },
              { max: 100, message: '模版名称不能超过100个字符' },
            ]}
          >
            <Input placeholder="请输入模版名称" />
          </Form.Item>

          <Form.Item
            label="描述"
            name="description"
          >
            <Input placeholder="请输入模版描述（可选）" />
          </Form.Item>

          <Form.Item
            label="YAML规范"
            name="yaml_spec"
            rules={[
              { required: true, message: '请输入YAML规范' },
            ]}
          >
            <TextArea
              rows={15}
              placeholder="请输入Kubernetes YAML规范，支持 ${PLACEHOLDER} 占位符"
              style={{ fontFamily: 'Monaco, Consolas, monospace' }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                更新
              </Button>
              <Button 
                onClick={() => {
                  setEditModalVisible(false);
                  editForm.resetFields();
                  setCurrentTemplate(null);
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 预览模版抽屉 */}
      <Drawer
        title="模版预览"
        open={previewVisible}
        onClose={() => {
          setPreviewVisible(false);
          setCurrentTemplate(null);
        }}
        width={800}
        placement="right"
      >
        {currentTemplate && (
          <div>
            <Card size="small" style={{ marginBottom: 16 }}>
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <div>
                  <Text strong>模版名称：</Text>
                  <Text>{currentTemplate.name}</Text>
                </div>
                <div>
                  <Text strong>描述：</Text>
                  <Text>{currentTemplate.description || '无'}</Text>
                </div>
                <div>
                  <Text strong>创建时间：</Text>
                  <Text>{formatDate(currentTemplate.created_at)}</Text>
                </div>
                <div>
                  <Text strong>更新时间：</Text>
                  <Text>{formatDate(currentTemplate.updated_at)}</Text>
                </div>
              </Space>
            </Card>

            <Divider orientation="left">YAML规范</Divider>
            <pre
              style={{
                backgroundColor: '#f5f5f5',
                padding: '16px',
                borderRadius: '6px',
                overflow: 'auto',
                fontSize: '12px',
                fontFamily: 'Monaco, Consolas, monospace',
              }}
            >
              {currentTemplate.yaml_spec}
            </pre>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default TemplateList;