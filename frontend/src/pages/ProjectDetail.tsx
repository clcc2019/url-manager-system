import React, { useState, useEffect, useCallback } from 'react';
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
  ClockCircleOutlined,
  RocketOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import type { Project, EphemeralURL, CreateURLRequest, EnvironmentVar } from '../types/api.js';
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

  const fetchProject = useCallback(async () => {
    if (!id) return;

    setLoading(true);
    try {
      const projectData = await ApiService.getProject(id);
      setProject(projectData);
    } catch {
      message.error('获取项目信息失败');
      navigate('/projects');
    } finally {
      setLoading(false);
    }
  }, [id, navigate]);

  const fetchURLs = useCallback(async () => {
    if (!id) return;

    setUrlsLoading(true);
    try {
      const response = await ApiService.getProjectURLs(id);
      setUrls(response.urls);
    } catch {
      message.error('获取URL列表失败');
    } finally {
      setUrlsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchProject();
    fetchURLs();
  }, [id, fetchProject, fetchURLs]);

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
        container_config: values.container_config ? {
          ...values.container_config,
          devices: values.container_config.devices?.filter((device: any) =>
            device.host_path && device.container_path
          ) || undefined,
          command: values.container_config.command?.filter((cmd: string) => cmd.trim()) || undefined,
          args: values.container_config.args?.filter((arg: string) => arg.trim()) || undefined,
        } : undefined,
      };

      const response = await ApiService.createURL(id, request);
      message.success(`URL创建成功: ${response.url}`);
      setCreateModalVisible(false);
      form.resetFields();
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL创建失败';
      message.error(errorMsg);
    }
  };

  const handleDeleteURL = async (urlId: string) => {
    try {
      await ApiService.deleteURL(urlId);
      message.success('URL删除成功');
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL删除失败';
      message.error(errorMsg);
    }
  };

  const handleDeployURL = async (urlId: string) => {
    try {
      await ApiService.deployURL(urlId);
      message.success('URL部署成功');
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL部署失败';
      message.error(errorMsg);
    }
  };


  const getStatusTag = (status: string) => {
    const statusConfig = {
      draft: { color: 'default', text: '草稿' },
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
      width: window.innerWidth < 768 ? 150 : undefined,
      render: (path: string, record: EphemeralURL) => {
        if (record.status === 'active') {
          return (
            <Button
              type="link"
              icon={<LinkOutlined />}
              onClick={() => window.open(`https://example.com${path}`, '_blank')}
              style={{ padding: 0, fontSize: window.innerWidth < 768 ? '12px' : undefined }}
            >
              {window.innerWidth < 768 ? path.substring(0, 20) + '...' : path}
            </Button>
          );
        }
        return <Text code style={{ fontSize: window.innerWidth < 768 ? '12px' : undefined }}>{path}</Text>;
      },
    },
    ...(window.innerWidth < 768 ? [] : [{
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
      ellipsis: true,
    }]),
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: window.innerWidth < 768 ? 80 : undefined,
      render: (status: string) => getStatusTag(status),
    },
    ...(window.innerWidth < 768 ? [] : [{
      title: '副本数',
      dataIndex: 'replicas',
      key: 'replicas',
    }]),
    {
      title: window.innerWidth < 768 ? '过期' : '过期时间',
      dataIndex: 'expire_at',
      key: 'expire_at',
      width: window.innerWidth < 768 ? 100 : undefined,
      render: (expireAt: string) => (
        <Space direction={window.innerWidth < 768 ? 'horizontal' : 'vertical'} size="small">
          <Text style={{ fontSize: window.innerWidth < 768 ? '12px' : undefined }}>
            {window.innerWidth < 768 ? expireAt.split('T')[0] : formatDate(expireAt)}
          </Text>
          {window.innerWidth >= 768 && (
            <Text type="secondary" style={{ fontSize: '12px' }}>
              <ClockCircleOutlined /> {getTimeUntilExpiry(expireAt)}
            </Text>
          )}
        </Space>
      ),
    },
    ...(window.innerWidth < 768 ? [] : [{
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => formatDate(text),
    }]),
    {
      title: '操作',
      key: 'actions',
      width: window.innerWidth < 768 ? 120 : 120,
      render: (_: any, record: EphemeralURL) => (
        <Space size={window.innerWidth < 768 ? 'small' : 'middle'}>
          {(record.status === 'draft' || record.status === 'failed') && (
            <Button
              type="primary"
              icon={<RocketOutlined />}
              size={window.innerWidth < 768 ? 'small' : 'small'}
              onClick={() => handleDeployURL(record.id)}
            >
              {window.innerWidth < 768 ? '' : (record.status === 'failed' ? '重新部署' : '部署')}
            </Button>
          )}
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
                size={window.innerWidth < 768 ? 'small' : 'small'}
              >
                {window.innerWidth < 768 ? '' : '删除'}
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
        title={
          <Title
            level={window.innerWidth < 768 ? 3 : 2}
            style={{
              margin: 0,
              fontSize: window.innerWidth < 768 ? '18px' : undefined
            }}
          >
            {project.name}
          </Title>
        }
        extra={
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              fetchProject();
              fetchURLs();
            }}
            size={window.innerWidth < 768 ? 'small' : 'middle'}
          >
            {window.innerWidth < 768 ? '' : '刷新'}
          </Button>
        }
      >
        <Descriptions column={window.innerWidth < 768 ? 1 : 2}>
          <Descriptions.Item label="项目ID">{project.id}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDate(project.created_at)}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{formatDate(project.updated_at)}</Descriptions.Item>
          <Descriptions.Item label="描述" span={window.innerWidth < 768 ? 1 : 2}>
            <Paragraph>{project.description || '无描述'}</Paragraph>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Divider />

      <Card
        title={<span style={{ fontSize: window.innerWidth < 768 ? '16px' : '18px' }}>URL 列表</span>}
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
            size={window.innerWidth < 768 ? 'small' : 'middle'}
          >
            {window.innerWidth < 768 ? '' : '创建URL'}
          </Button>
        }
      >
        <Table
          columns={urlColumns}
          dataSource={urls}
          rowKey="id"
          loading={urlsLoading}
          scroll={{ x: window.innerWidth < 768 ? 800 : undefined }}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            size: window.innerWidth < 768 ? 'small' : 'default',
          }}
          size={window.innerWidth < 768 ? 'small' : 'middle'}
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
        width={window.innerWidth < 768 ? '95%' : 800}
        style={{ maxWidth: '95vw' }}
        bodyStyle={{
          maxHeight: window.innerWidth < 768 ? '70vh' : 'none',
          overflowY: window.innerWidth < 768 ? 'auto' : 'visible'
        }}
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

          <Divider orientation="left">容器配置</Divider>

          <Form.Item
            label="容器名称"
            name={['container_config', 'container_name']}
            rules={[
              {
                pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
                message: '容器名称只能包含小写字母、数字和连字符，且必须以字母数字开头和结尾'
              }
            ]}
          >
            <Input placeholder="可选，自定义容器名称" />
          </Form.Item>

          <Form.Item
            label="工作目录"
            name={['container_config', 'working_dir']}
            rules={[
              {
                pattern: /^\/.*/,
                message: '工作目录必须是绝对路径'
              }
            ]}
          >
            <Input placeholder="例如: /app" />
          </Form.Item>

          <Form.Item label="TTY">
            <Form.Item name={['container_config', 'tty']} noStyle valuePropName="checked">
              <input type="checkbox" />
            </Form.Item>
            <span style={{ marginLeft: 8 }}>分配TTY</span>
          </Form.Item>

          <Form.Item label="Stdin">
            <Form.Item name={['container_config', 'stdin']} noStyle valuePropName="checked">
              <input type="checkbox" />
            </Form.Item>
            <span style={{ marginLeft: 8 }}>打开Stdin</span>
          </Form.Item>

          <Divider orientation="left">设备映射</Divider>

          <Form.List name={['container_config', 'devices']}>
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8, flexWrap: 'wrap' }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'host_path']}
                      rules={[
                        { required: true, message: '请输入主机路径' },
                        { pattern: /^\/.*/, message: '主机路径必须是绝对路径' }
                      ]}
                      style={{ minWidth: 200 }}
                    >
                      <Input placeholder="主机路径 (例如: /dev/kvm)" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'container_path']}
                      rules={[
                        { required: true, message: '请输入容器路径' },
                        { pattern: /^\/.*/, message: '容器路径必须是绝对路径' }
                      ]}
                      style={{ minWidth: 200 }}
                    >
                      <Input placeholder="容器路径 (例如: /dev/kvm)" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'permissions']}
                      rules={[
                        { pattern: /^[rwm]*$/, message: '权限只能包含 r、w、m 字符' }
                      ]}
                      style={{ minWidth: 100 }}
                    >
                      <Input placeholder="权限 (rwm)" />
                    </Form.Item>
                    <Button onClick={() => remove(name)} icon={<DeleteOutlined />} />
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加设备映射
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Divider orientation="left">启动配置</Divider>

          <Form.Item
            label="启动命令"
            name={['container_config', 'command']}
          >
            <Select mode="tags" placeholder="可选，覆盖默认启动命令" />
          </Form.Item>

          <Form.Item
            label="启动参数"
            name={['container_config', 'args']}
          >
            <Select mode="tags" placeholder="可选，传递给启动命令的参数" />
          </Form.Item>

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