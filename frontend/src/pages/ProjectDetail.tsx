import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Descriptions, 
  Button, 
  Table, 
  message, 
  Modal, 
  Form, 
  Input, 
  InputNumber, 
  Select, 
  Tag, 
  Space,
  Popconfirm,
  Typography,
  Divider
} from 'antd';
import { 
  PlusOutlined, 
  DeleteOutlined, 
  ReloadOutlined,
  LinkOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { Project, EphemeralURL, CreateURLRequest, EnvironmentVar } from '../types/api';
import { ApiService } from '../services/api';
import { formatDate, getTimeUntilExpiry } from '../utils/date';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  const [project, setProject] = useState<Project | null>(null);
  const [urls, setUrls] = useState<EphemeralURL[]>([]);
  const [loading, setLoading] = useState(false);
  const [urlsLoading, setUrlsLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchProject = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const projectData = await ApiService.getProject(id);
      setProject(projectData);
    } catch (error) {
      message.error('获取项目信息失败');
      navigate('/projects');
    } finally {
      setLoading(false);
    }
  };

  const fetchURLs = async () => {
    if (!id) return;
    
    setUrlsLoading(true);
    try {
      const response = await ApiService.getProjectURLs(id);
      setUrls(response.urls);
    } catch (error) {
      message.error('获取URL列表失败');
    } finally {
      setUrlsLoading(false);
    }
  };

  useEffect(() => {
    fetchProject();
    fetchURLs();
  }, [id]);

  const handleCreateURL = async (values: any) => {
    if (!id) return;

    try {
      const request: CreateURLRequest = {
        image: values.image,
        ttl_seconds: values.ttl_seconds,
        replicas: values.replicas || 1,
        env: values.env?.filter((env: EnvironmentVar) => env.name && env.value) || [],
        resources: {
          requests: {
            cpu: values.requests_cpu || '100m',
            memory: values.requests_memory || '128Mi',
          },
          limits: {
            cpu: values.limits_cpu || '500m',
            memory: values.limits_memory || '512Mi',
          },
        },
      };

      const response = await ApiService.createURL(id, request);
      message.success(`URL创建成功: ${response.url}`);
      setCreateModalVisible(false);
      form.resetFields();
      fetchURLs();
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || 'URL创建失败';
      message.error(errorMsg);
    }
  };

  const handleDeleteURL = async (urlId: string) => {
    try {
      await ApiService.deleteURL(urlId);
      message.success('URL删除成功');
      fetchURLs();
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || 'URL删除失败';
      message.error(errorMsg);
    }
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      creating: { color: 'processing', text: '创建中' },
      active: { color: 'success', text: '运行中' },
      deleting: { color: 'warning', text: '删除中' },
      deleted: { color: 'default', text: '已删除' },
      failed: { color: 'error', text: '失败' },
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const urlColumns = [
    {
      title: 'URL路径',
      dataIndex: 'path',
      key: 'path',
      render: (path: string, record: EphemeralURL) => {
        if (record.status === 'active') {
          return (
            <Button 
              type="link" 
              icon={<LinkOutlined />}
              onClick={() => window.open(`https://example.com${path}`, '_blank')}
              style={{ padding: 0 }}
            >
              {path}
            </Button>
          );
        }
        return <Text code>{path}</Text>;
      },
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '副本数',
      dataIndex: 'replicas',
      key: 'replicas',
    },
    {
      title: '过期时间',
      dataIndex: 'expire_at',
      key: 'expire_at',
      render: (expireAt: string) => (
        <Space direction="vertical" size="small">
          <Text>{formatDate(expireAt)}</Text>
          <Text type="secondary" style={{ fontSize: '12px' }}>
            <ClockCircleOutlined /> {getTimeUntilExpiry(expireAt)}
          </Text>
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => formatDate(text),
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_: any, record: EphemeralURL) => (
        <Space>
          {record.status !== 'deleted' && (
            <Popconfirm
              title="确定要删除这个URL吗？"
              description="删除后将无法访问"
              onConfirm={() => handleDeleteURL(record.id)}
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
      ),
    },
  ];

  const ttlOptions = [
    { label: '30分钟', value: 1800 },
    { label: '1小时', value: 3600 },
    { label: '6小时', value: 21600 },
    { label: '1天', value: 86400 },
    { label: '3天', value: 259200 },
    { label: '7天', value: 604800 },
  ];

  if (loading || !project) {
    return <div>加载中...</div>;
  }

  return (
    <div>
      <Card 
        title={<Title level={2} style={{ margin: 0 }}>{project.name}</Title>}
        extra={
          <Button 
            icon={<ReloadOutlined />}
            onClick={() => {
              fetchProject();
              fetchURLs();
            }}
          >
            刷新
          </Button>
        }
      >
        <Descriptions column={2}>
          <Descriptions.Item label="项目ID">{project.id}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDate(project.created_at)}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{formatDate(project.updated_at)}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            <Paragraph>{project.description || '无描述'}</Paragraph>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Divider />

      <Card 
        title="URL 列表"
        extra={
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建URL
          </Button>
        }
      >
        <Table
          columns={urlColumns}
          dataSource={urls}
          rowKey="id"
          loading={urlsLoading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title="创建临时URL"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateURL}
        >
          <Form.Item
            label="容器镜像"
            name="image"
            rules={[{ required: true, message: '请输入容器镜像' }]}
          >
            <Input placeholder="例如: nginx:latest" />
          </Form.Item>

          <Form.Item
            label="过期时间"
            name="ttl_seconds"
            rules={[{ required: true, message: '请选择过期时间' }]}
          >
            <Select placeholder="选择过期时间">
              {ttlOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  {option.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="副本数"
            name="replicas"
            initialValue={1}
          >
            <InputNumber min={1} max={10} />
          </Form.Item>

          <Divider orientation="left">资源配置</Divider>
          
          <Form.Item label="CPU请求">
            <Form.Item name="requests_cpu" noStyle initialValue="100m">
              <Input placeholder="100m" addonAfter="cores" />
            </Form.Item>
          </Form.Item>

          <Form.Item label="内存请求">
            <Form.Item name="requests_memory" noStyle initialValue="128Mi">
              <Input placeholder="128Mi" addonAfter="bytes" />
            </Form.Item>
          </Form.Item>

          <Form.Item label="CPU限制">
            <Form.Item name="limits_cpu" noStyle initialValue="500m">
              <Input placeholder="500m" addonAfter="cores" />
            </Form.Item>
          </Form.Item>

          <Form.Item label="内存限制">
            <Form.Item name="limits_memory" noStyle initialValue="512Mi">
              <Input placeholder="512Mi" addonAfter="bytes" />
            </Form.Item>
          </Form.Item>

          <Divider orientation="left">环境变量</Divider>

          <Form.List name="env">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'name']}
                      rules={[{ required: true, message: '请输入变量名' }]}
                    >
                      <Input placeholder="变量名" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入变量值' }]}
                    >
                      <Input placeholder="变量值" />
                    </Form.Item>
                    <Button onClick={() => remove(name)} icon={<DeleteOutlined />} />
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加环境变量
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建URL
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

export default ProjectDetail;