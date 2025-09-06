import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { EphemeralURL, UpdateURLRequest, ContainerStatus, PodEvent, ContainerLog } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate, getTimeUntilExpiry } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
import { useEditor } from '@/hooks/useEditor';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import UnifiedEditor from '@/components/UnifiedEditor';
import {
  ArrowLeft,
  Link,
  Copy,
  Eye,
  Trash2,
  Rocket,
  Clock,
  Server,
  Settings,
  Activity,
  RefreshCw,
  FileText,
  AlertCircle,
  CheckCircle,
  Info,
  Edit,
  Save,
  X,
  Container,
  Calendar,
  Terminal,
  Play,
  Pause,
  RotateCcw,
  Download,
  ExternalLink,
  Globe,
  Cpu,
  MemoryStick,
  HardDrive
} from 'lucide-react';
import { Popconfirm } from '@/components/ui/popconfirm';
import { cn } from '@/lib/utils';

const URLDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  
  const [url, setUrl] = useState<EphemeralURL | null>(null);
  const [loading, setLoading] = useState(false);
  const [containerStatuses, setContainerStatuses] = useState<ContainerStatus[]>([]);
  const [podEvents, setPodEvents] = useState<PodEvent[]>([]);
  const [containerLogs, setContainerLogs] = useState<ContainerLog[]>([]);
  const [statusLoading, setStatusLoading] = useState(false);
  const [eventsLoading, setEventsLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [selectedContainer, setSelectedContainer] = useState<string>('');
  const [logLines, setLogLines] = useState<number>(100);
  const editor = useEditor<EphemeralURL>();

  const fetchURL = async () => {
    if (!id) return;

    setLoading(true);
    try {
      const urlData = await ApiService.getURL(id);
      setUrl(urlData);
    } catch {
      toast({
        title: '获取失败',
        description: '获取URL信息失败',
        variant: 'destructive',
      });
      navigate(-1);
    } finally {
      setLoading(false);
    }
  };

  const fetchContainerStatus = async () => {
    if (!id) return;

    setStatusLoading(true);
    try {
      const statuses = await ApiService.getURLContainerStatus(id);
      setContainerStatuses(statuses);
      if (statuses.length > 0 && !selectedContainer) {
        setSelectedContainer(statuses[0].name);
      }
    } catch (error) {
      toast({
        title: '获取失败',
        description: '获取容器状态失败',
        variant: 'destructive',
      });
    } finally {
      setStatusLoading(false);
    }
  };

  const fetchPodEvents = async () => {
    if (!id) return;

    setEventsLoading(true);
    try {
      const events = await ApiService.getURLPodEvents(id);
      setPodEvents(events);
    } catch (error) {
      toast({
        title: '获取失败',
        description: '获取Pod事件失败',
        variant: 'destructive',
      });
    } finally {
      setEventsLoading(false);
    }
  };

  const fetchContainerLogs = async () => {
    if (!id) return;

    setLogsLoading(true);
    try {
      const logs = await ApiService.getURLContainerLogs(id, selectedContainer, logLines);
      setContainerLogs(logs);
    } catch (error) {
      toast({
        title: '获取失败',
        description: '获取容器日志失败',
        variant: 'destructive',
      });
    } finally {
      setLogsLoading(false);
    }
  };

  useEffect(() => {
    fetchURL();
  }, [id]);

  const handleDelete = async () => {
    if (!url) return;
    
    try {
      await ApiService.deleteURL(url.id);
      toast({
        title: '删除成功',
        description: 'URL已成功删除',
      });
      navigate(-1);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL删除失败';
      toast({
        title: '删除失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDeploy = async () => {
    if (!url) return;
    
    try {
      await ApiService.deployURL(url.id);
      toast({
        title: '部署成功',
        description: 'URL已成功重新部署',
      });
      
      // 立即刷新所有数据
      await fetchURL();
      fetchContainerStatus();
      fetchPodEvents();
      if (selectedContainer) {
        fetchContainerLogs();
      }
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL部署失败';
      toast({
        title: '部署失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: '复制成功',
      description: '链接已复制到剪贴板',
    });
  };

  const startEditing = () => {
    if (!url) return;
    editor.startEditing(url);
  };

  const cancelEditing = () => {
    editor.cancelEditing();
  };

  const saveEditing = async () => {
    if (!url || !editor.editedData || !id) return;

    try {
      const updateData: UpdateURLRequest = {
        image: editor.editedData.image,
        env: editor.editedData.env,
        ttl_seconds: editor.editedData.ttl_seconds,
        replicas: editor.editedData.replicas,
        resources: editor.editedData.resources,
        container_config: editor.editedData.container_config,
        ingress_host: editor.editedData.ingress_host
      };

      // 更新URL配置
      await ApiService.updateURL(id, updateData);
      
      // 立即重新部署以应用更改到集群
      await ApiService.deployURL(id);
      
      toast({
        title: '更新成功',
        description: 'URL配置已更新并重新部署',
      });
      
      editor.cancelEditing();
      
      // 立即刷新所有数据
      await fetchURL();
      fetchContainerStatus();
      fetchPodEvents();
      if (selectedContainer) {
        fetchContainerLogs();
      }
      
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL更新失败';
      toast({
        title: '更新失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  if (loading || !url) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <Spinner size="lg" tip="加载中..." />
      </div>
    );
  }

  const statusConfig = {
    draft: { variant: 'secondary' as const, text: '草稿', color: 'text-gray-600', icon: FileText },
    creating: { variant: 'default' as const, text: '创建中', color: 'text-blue-600', icon: RefreshCw },
    waiting: { variant: 'outline' as const, text: '等待中', color: 'text-yellow-600', icon: Clock },
    active: { variant: 'default' as const, text: '运行中', color: 'text-green-600', icon: CheckCircle },
    deleting: { variant: 'destructive' as const, text: '删除中', color: 'text-red-600', icon: Trash2 },
    deleted: { variant: 'secondary' as const, text: '已删除', color: 'text-gray-600', icon: X },
    failed: { variant: 'destructive' as const, text: '失败', color: 'text-red-600', icon: AlertCircle },
  };
  
  const config = statusConfig[url.status as keyof typeof statusConfig] || { 
    variant: 'secondary' as const, 
    text: url.status,
    color: 'text-gray-600',
    icon: Info
  };

  const StatusIcon = config.icon;

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center space-x-2 text-sm text-muted-foreground">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate(-1)}
          className="h-auto p-0 text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="mr-1 h-3 w-3" />
          返回
        </Button>
        <span>/</span>
        <span>URL详情</span>
      </div>

      {/* Header */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-2">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold tracking-tight">URL详情</h1>
            <Badge variant={config.variant} className="flex items-center gap-1">
              <StatusIcon className="h-3 w-3" />
              {config.text}
            </Badge>
          </div>
          <div className="flex items-center gap-2 text-muted-foreground">
            <Globe className="h-4 w-4" />
            <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
              {url.path}
            </code>
            {url.status === 'active' && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => window.open(`https://example.com${url.path}`, '_blank')}
                  className="h-auto p-1"
                >
                  <ExternalLink className="h-3 w-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(`https://example.com${url.path}`)}
                  className="h-auto p-1"
                >
                  <Copy className="h-3 w-3" />
                </Button>
              </>
            )}
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchURL}
            className="flex items-center gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          {!editor.editing ? (
            <>
              {(url.status === 'draft' || url.status === 'failed' || url.status === 'active') && (
                <Button
                  size="sm"
                  onClick={handleDeploy}
                  className="flex items-center gap-2"
                  variant={url.status === 'active' ? 'outline' : 'default'}
                >
                  <Rocket className="h-4 w-4" />
                  {url.status === 'failed' ? '重新部署' : url.status === 'active' ? '重新部署' : '部署'}
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={startEditing}
                className="flex items-center gap-2"
              >
                <Edit className="h-4 w-4" />
                编辑
              </Button>
              {url.status === 'active' && (
                <Button
                  size="sm"
                  onClick={() => window.open(`https://example.com${url.path}`, '_blank')}
                  className="flex items-center gap-2"
                >
                  <Eye className="h-4 w-4" />
                  访问
                </Button>
              )}
              {url.status !== 'deleted' && (
                <Popconfirm
                  title="确定要删除这个URL吗？"
                  description="删除后将无法访问"
                  onConfirm={handleDelete}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    variant="destructive"
                    size="sm"
                    className="flex items-center gap-2"
                  >
                    <Trash2 className="h-4 w-4" />
                    删除
                  </Button>
                </Popconfirm>
              )}
            </>
          ) : (
            <>
              <Button
                size="sm"
                onClick={saveEditing}
                className="flex items-center gap-2"
              >
                <Save className="h-4 w-4" />
                保存
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={cancelEditing}
                className="flex items-center gap-2"
              >
                <X className="h-4 w-4" />
                取消
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Main content */}
      <Tabs 
        defaultValue="overview" 
        className="space-y-6"
        onValueChange={(value) => {
          if (value === 'containers' && containerStatuses.length === 0) {
            fetchContainerStatus();
          } else if (value === 'events' && podEvents.length === 0) {
            fetchPodEvents();
          } else if (value === 'logs' && containerLogs.length === 0) {
            fetchContainerLogs();
          }
        }}
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <Info className="h-4 w-4" />
            概览
          </TabsTrigger>
          <TabsTrigger value="containers" className="flex items-center gap-2">
            <Container className="h-4 w-4" />
            容器状态
          </TabsTrigger>
          <TabsTrigger value="events" className="flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            事件
          </TabsTrigger>
          <TabsTrigger value="logs" className="flex items-center gap-2">
            <Terminal className="h-4 w-4" />
            日志
          </TabsTrigger>
        </TabsList>

        {/* 概览标签页 */}
        <TabsContent value="overview" className="space-y-6">
          {/* Quick Stats */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">副本数</CardTitle>
                <Server className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{url.replicas}</div>
                <p className="text-xs text-muted-foreground">
                  运行中的实例数量
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">CPU请求</CardTitle>
                <Cpu className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{url.resources?.requests?.cpu || 'N/A'}</div>
                <p className="text-xs text-muted-foreground">
                  CPU资源请求
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">内存请求</CardTitle>
                <MemoryStick className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{url.resources?.requests?.memory || 'N/A'}</div>
                <p className="text-xs text-muted-foreground">
                  内存资源请求
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">过期时间</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">
                  {getTimeUntilExpiry(url.expire_at).split(' ')[0]}
                </div>
                <p className="text-xs text-muted-foreground">
                  {getTimeUntilExpiry(url.expire_at)}
                </p>
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-6 lg:grid-cols-2">
            {/* 基本信息 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Info className="h-5 w-5" />
                  基本信息
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-muted-foreground">状态</span>
                    <Badge variant={config.variant} className="flex items-center gap-1">
                      <StatusIcon className="h-3 w-3" />
                      {config.text}
                    </Badge>
                  </div>
                  <Separator />
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-muted-foreground">创建时间</span>
                    <span className="text-sm">{formatDate(url.created_at)}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-muted-foreground">更新时间</span>
                    <span className="text-sm">{formatDate(url.updated_at)}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-muted-foreground">过期时间</span>
                    <span className="text-sm">{formatDate(url.expire_at)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* 镜像信息 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Server className="h-5 w-5" />
                  {url.template ? '模板信息' : '镜像信息'}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {url.template ? (
                  <div className="space-y-3">
                    <div>
                      <div className="font-medium">{url.template.name}</div>
                      <div className="text-sm text-muted-foreground">{url.template.description}</div>
                    </div>
                    <Separator />
                    <div>
                      <span className="text-sm font-medium text-muted-foreground">实际镜像:</span>
                      <code className="ml-2 text-sm bg-muted px-2 py-1 rounded">{url.image}</code>
                    </div>
                  </div>
                ) : (
                  <div>
                    <span className="text-sm font-medium text-muted-foreground">镜像:</span>
                    <code className="ml-2 text-sm bg-muted px-2 py-1 rounded">{url.image}</code>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 环境变量 */}
          {url.env && url.env.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Settings className="h-5 w-5" />
                  环境变量
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {url.env.map((env, index) => (
                    <div key={index} className="flex items-center gap-2 text-sm">
                      <code className="bg-muted px-2 py-1 rounded font-medium">{env.name}</code>
                      <span className="text-muted-foreground">=</span>
                      <code className="bg-muted px-2 py-1 rounded flex-1">{env.value}</code>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* 编辑表单 */}
          {editor.editing && editor.editedData && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Edit className="h-5 w-5" />
                  编辑配置
                </CardTitle>
                <CardDescription>
                  修改URL的配置信息，保存后将自动重新部署
                </CardDescription>
              </CardHeader>
              <CardContent>
                <UnifiedEditor
                  type="url"
                  data={editor.editedData}
                  onUpdate={editor.updateData}
                />
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* 容器状态标签页 */}
        <TabsContent value="containers" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Container className="h-5 w-5" />
                  容器状态
                </CardTitle>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={fetchContainerStatus}
                  disabled={statusLoading}
                  className="flex items-center gap-2"
                >
                  <RefreshCw className={`h-4 w-4 ${statusLoading ? 'animate-spin' : ''}`} />
                  刷新
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {statusLoading ? (
                <div className="flex justify-center items-center py-8">
                  <Spinner size="lg" tip="加载容器状态..." />
                </div>
              ) : containerStatuses.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Container className="mx-auto h-12 w-12 opacity-50" />
                  <h3 className="mt-4 text-lg font-semibold">暂无容器状态</h3>
                  <p className="text-sm">容器可能尚未启动或URL状态不是运行中</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {containerStatuses.map((container, index) => (
                    <div key={index} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-3">
                        <h4 className="font-semibold flex items-center gap-2">
                          <Container className="h-4 w-4" />
                          {container.name}
                        </h4>
                        <div className="flex items-center gap-2">
                          <Badge variant={container.ready ? 'default' : 'destructive'}>
                            {container.ready ? '就绪' : '未就绪'}
                          </Badge>
                          <Badge variant={container.started ? 'default' : 'secondary'}>
                            {container.started ? '已启动' : '未启动'}
                          </Badge>
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="font-medium">镜像:</span>
                          <p className="font-mono text-xs mt-1">{container.image}</p>
                        </div>
                        <div>
                          <span className="font-medium">重启次数:</span>
                          <p className="mt-1">{container.restart_count}</p>
                        </div>
                        {container.container_id && (
                          <div className="col-span-2">
                            <span className="font-medium">容器ID:</span>
                            <p className="font-mono text-xs mt-1">{container.container_id}</p>
                          </div>
                        )}
                      </div>

                      {/* 容器状态详情 */}
                      <div className="mt-4 space-y-2">
                        {container.state.running && (
                          <div className="flex items-center gap-2 text-green-600">
                            <Play className="h-4 w-4" />
                            <span className="text-sm">运行中</span>
                            {container.state.running.started_at && (
                              <span className="text-xs text-muted-foreground">
                                启动于 {formatDate(container.state.running.started_at)}
                              </span>
                            )}
                          </div>
                        )}
                        
                        {container.state.waiting && (
                          <div className="flex items-center gap-2 text-yellow-600">
                            <Pause className="h-4 w-4" />
                            <span className="text-sm">等待中</span>
                            {container.state.waiting.reason && (
                              <span className="text-xs">({container.state.waiting.reason})</span>
                            )}
                          </div>
                        )}
                        
                        {container.state.terminated && (
                          <div className="flex items-center gap-2 text-red-600">
                            <X className="h-4 w-4" />
                            <span className="text-sm">已终止</span>
                            <span className="text-xs">退出码: {container.state.terminated.exit_code}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 事件标签页 */}
        <TabsContent value="events" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Calendar className="h-5 w-5" />
                  Pod事件
                </CardTitle>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={fetchPodEvents}
                  disabled={eventsLoading}
                  className="flex items-center gap-2"
                >
                  <RefreshCw className={`h-4 w-4 ${eventsLoading ? 'animate-spin' : ''}`} />
                  刷新
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {eventsLoading ? (
                <div className="flex justify-center items-center py-8">
                  <Spinner size="lg" tip="加载事件..." />
                </div>
              ) : podEvents.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Calendar className="mx-auto h-12 w-12 opacity-50" />
                  <h3 className="mt-4 text-lg font-semibold">暂无事件</h3>
                  <p className="text-sm">没有找到相关的Pod事件</p>
                </div>
              ) : (
                <div className="space-y-3 max-h-96 overflow-y-auto">
                  {podEvents.map((event, index) => {
                    const getEventIcon = (type: string) => {
                      switch (type.toLowerCase()) {
                        case 'warning':
                          return <AlertCircle className="h-4 w-4 text-yellow-500" />;
                        case 'error':
                          return <AlertCircle className="h-4 w-4 text-red-500" />;
                        default:
                          return <Info className="h-4 w-4 text-blue-500" />;
                      }
                    };

                    const getEventColor = (type: string) => {
                      switch (type.toLowerCase()) {
                        case 'warning':
                          return 'border-yellow-200 bg-yellow-50';
                        case 'error':
                          return 'border-red-200 bg-red-50';
                        default:
                          return 'border-blue-200 bg-blue-50';
                      }
                    };

                    return (
                      <div
                        key={index}
                        className={cn(
                          "flex items-start gap-3 p-3 rounded-lg border",
                          getEventColor(event.type)
                        )}
                      >
                        {getEventIcon(event.type)}
                        <div className="flex-1 space-y-1">
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-sm">{event.reason}</span>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground">
                              {event.count > 1 && (
                                <Badge variant="secondary" className="text-xs">
                                  {event.count}次
                                </Badge>
                              )}
                              <span>{formatDate(event.last_timestamp)}</span>
                            </div>
                          </div>
                          <p className="text-sm text-muted-foreground">{event.message}</p>
                          {event.source_component && (
                            <p className="text-xs text-muted-foreground">
                              来源: {event.source_component}
                            </p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 日志标签页 */}
        <TabsContent value="logs" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Terminal className="h-5 w-5" />
                  容器日志
                </CardTitle>
                <div className="flex items-center gap-2">
                  {containerStatuses.length > 0 && (
                    <select
                      value={selectedContainer}
                      onChange={(e) => setSelectedContainer(e.target.value)}
                      className="text-sm border rounded px-2 py-1"
                    >
                      {containerStatuses.map((container) => (
                        <option key={container.name} value={container.name}>
                          {container.name}
                        </option>
                      ))}
                    </select>
                  )}
                  <select
                    value={logLines}
                    onChange={(e) => setLogLines(parseInt(e.target.value))}
                    className="text-sm border rounded px-2 py-1"
                  >
                    <option value={50}>50行</option>
                    <option value={100}>100行</option>
                    <option value={200}>200行</option>
                    <option value={500}>500行</option>
                  </select>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={fetchContainerLogs}
                    disabled={logsLoading}
                    className="flex items-center gap-2"
                  >
                    <RefreshCw className={`h-4 w-4 ${logsLoading ? 'animate-spin' : ''}`} />
                    刷新
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {logsLoading ? (
                <div className="flex justify-center items-center py-8">
                  <Spinner size="lg" tip="加载日志..." />
                </div>
              ) : containerLogs.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Terminal className="mx-auto h-12 w-12 opacity-50" />
                  <h3 className="mt-4 text-lg font-semibold">暂无日志</h3>
                  <p className="text-sm">没有找到容器日志或容器尚未启动</p>
                </div>
              ) : (
                <div className="bg-slate-950 text-green-400 p-4 rounded-lg font-mono text-sm max-h-96 overflow-y-auto">
                  {containerLogs.map((log, index) => (
                    <div key={index} className="flex gap-2">
                      <span className="text-slate-500 text-xs whitespace-nowrap">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </span>
                      <span className="flex-1">{log.log}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default URLDetail;