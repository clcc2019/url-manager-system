import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { Project, EphemeralURL, CreateURLRequest, CreateURLFromTemplateRequest, EnvironmentVar, AppTemplate } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate, getTimeUntilExpiry } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
import { useEditor } from '@/hooks/useEditor';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Popconfirm } from '@/components/ui/popconfirm';
import { Spinner } from '@/components/ui/spinner';
import UnifiedEditor from '@/components/UnifiedEditor';
import { Separator } from '@/components/ui/separator';
import { 
  Plus, 
  Trash2, 
  RefreshCw,
  Clock,
  Rocket,
  Loader2,
  Link,
  Server,
  Copy,
  Eye,
  Settings,
  Activity,
  ArrowLeft,
  MoreHorizontal,
  ExternalLink,
  Globe,
  Calendar,
  FileText,
  AlertCircle,
  CheckCircle,
  Info
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  
  const [project, setProject] = useState<Project | null>(null);
  const [urls, setUrls] = useState<EphemeralURL[]>([]);
  const [templates, setTemplates] = useState<AppTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [urlsLoading, setUrlsLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createMode, setCreateMode] = useState<'manual' | 'template'>('manual');
  const [selectedTemplate, setSelectedTemplate] = useState<AppTemplate | null>(null);
  const [editingURL, setEditingURL] = useState<EphemeralURL | null>(null);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [formData, setFormData] = useState<any>({});
  const [formErrors, setFormErrors] = useState<any>({});
  const urlEditor = useEditor<EphemeralURL>();

  const fetchProject = useCallback(async () => {
    if (!id) return;

    setLoading(true);
    try {
      const projectData = await ApiService.getProject(id);
      setProject(projectData);
    } catch (error: any) {
      console.error('Failed to fetch project:', error);
      const errorMessage = error?.response?.data?.error || '获取项目信息失败';
      toast({
        title: '获取失败',
        description: errorMessage,
        variant: 'destructive',
      });
      navigate('/projects');
    } finally {
      setLoading(false);
    }
  }, [id, navigate, toast]);

  const fetchURLs = useCallback(async () => {
    if (!id) return;

    setUrlsLoading(true);
    try {
      const response = await ApiService.getProjectURLs(id);
      setUrls(response.urls || []);
    } catch {
      toast({
        title: '获取失败',
        description: '获取URL列表失败',
        variant: 'destructive',
      });
      setUrls([]); // 出错时设置为空数组
    } finally {
      setUrlsLoading(false);
    }
  }, [id, toast]);

  const fetchTemplates = useCallback(async () => {
    try {
      const response = await ApiService.getTemplates();
      setTemplates(response.templates);
    } catch {
      toast({
        title: '获取失败',
        description: '获取模版列表失败',
        variant: 'destructive',
      });
    }
  }, [toast]);

  useEffect(() => {
    fetchProject();
    fetchURLs();
    fetchTemplates();
  }, [fetchProject, fetchURLs, fetchTemplates]);

  const handleCreateURL = async (values: any) => {
    if (!id) return;

    try {
      if (createMode === 'template' && selectedTemplate) {
        const request: CreateURLFromTemplateRequest = {
          template_id: selectedTemplate.id,
          ttl_seconds: values.ttl_seconds,
          variables: values.variables || {},
        };
        
        const response = await ApiService.createURLFromTemplate(id, request);
        toast({
          title: '创建成功',
          description: `URL创建成功: ${response.url}`,
        });
      } else {
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
        toast({
          title: '创建成功',
          description: `URL创建成功: ${response.url}`,
        });
      }
      
      setCreateModalVisible(false);
      setCreateMode('manual');
      setSelectedTemplate(null);
      setFormData({});
      setFormErrors({});
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL创建失败';
      toast({
        title: '创建失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDeleteURL = async (urlId: string) => {
    try {
      await ApiService.deleteURL(urlId);
      toast({
        title: '删除成功',
        description: 'URL已成功删除',
      });
      // 重新获取URL列表，因为URL已被真正删除
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL删除失败';
      toast({
        title: '删除失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleEditURL = async (url: EphemeralURL) => {
    try {
      // 获取完整URL详情，包含 container_config 等字段
      const full = await ApiService.getURL(url.id);
      setEditingURL(full);
      urlEditor.startEditing(full);
      setEditModalVisible(true);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '获取URL详情失败，无法编辑';
      toast({
        title: '获取失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleUpdateURL = async () => {
    if (!editingURL || !urlEditor.editedData) return;

    try {
      const updateData = {
        image: urlEditor.editedData.image,
        env: urlEditor.editedData.env,
        ttl_seconds: urlEditor.editedData.ttl_seconds,
        replicas: urlEditor.editedData.replicas,
        resources: urlEditor.editedData.resources,
        container_config: urlEditor.editedData.container_config,
        ingress_host: urlEditor.editedData.ingress_host
      };

      await ApiService.updateURL(editingURL.id, updateData);
      // 更新后立即重新部署，避免等待失败
      await ApiService.deployURL(editingURL.id);

      toast({
        title: '更新成功',
        description: 'URL配置已更新并重新部署',
      });

      setEditModalVisible(false);
      setEditingURL(null);
      urlEditor.cancelEditing();

      // 刷新URL列表
      fetchURLs();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || 'URL更新失败';
      toast({
        title: '更新失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDeployURL = async (urlId: string) => {
    try {
      await ApiService.deployURL(urlId);
      toast({
        title: '部署成功',
        description: 'URL已成功部署',
      });
      fetchURLs();
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

  const ttlOptions = [
    { label: '30分钟', value: 1800 },
    { label: '1小时', value: 3600 },
    { label: '6小时', value: 21600 },
    { label: '1天', value: 86400 },
    { label: '3天', value: 259200 },
    { label: '7天', value: 604800 },
  ];

  // 计算统计数据
  const stats = {
    totalUrls: urls?.length || 0,
    activeUrls: urls?.filter(url => url.status === 'active').length || 0,
    failedUrls: urls?.filter(url => url.status === 'failed').length || 0,
    pendingUrls: urls?.filter(url => ['creating', 'waiting'].includes(url.status)).length || 0,
  };

  if (loading || !project) {
    return (
      <div className="flex flex-col justify-center items-center min-h-96 space-y-4">
        <Spinner size="lg" />
        <div className="text-center">
          <p className="text-lg font-medium">正在加载项目详情...</p>
          <p className="text-sm text-muted-foreground">项目ID: {id}</p>
        </div>
      </div>
    );
  }

  const statusConfig = {
    draft: { variant: 'secondary' as const, text: '草稿', icon: FileText },
    creating: { variant: 'default' as const, text: '创建中', icon: RefreshCw },
    waiting: { variant: 'outline' as const, text: '等待中', icon: Clock },
    active: { variant: 'default' as const, text: '运行中', icon: CheckCircle },
    deleting: { variant: 'destructive' as const, text: '删除中', icon: Trash2 },
    deleted: { variant: 'secondary' as const, text: '已删除', icon: AlertCircle },
    failed: { variant: 'destructive' as const, text: '失败', icon: AlertCircle },
  };

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center space-x-2 text-sm text-muted-foreground">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate('/projects')}
          className="h-auto p-0 text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="mr-1 h-3 w-3" />
          <span className="hidden sm:inline">项目列表</span>
          <span className="sm:hidden">返回</span>
        </Button>
        <span>/</span>
        <span className="hidden sm:inline">项目详情</span>
        <span className="sm:hidden">详情</span>
      </div>

      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-2 flex-1 min-w-0">
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl truncate">{project.name}</h1>
          <p className="text-sm text-muted-foreground sm:text-base">
            {project.description || '暂无描述'}
          </p>
          <div className="flex items-center gap-2 text-xs text-muted-foreground sm:text-sm">
            <Calendar className="h-3 w-3 sm:h-4 sm:w-4" />
            <span>创建于 {formatDate(project.created_at)}</span>
          </div>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              fetchProject();
              fetchURLs();
            }}
            className="flex items-center gap-2 w-full sm:w-auto"
          >
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          <Dialog open={createModalVisible} onOpenChange={setCreateModalVisible}>
            <DialogTrigger asChild>
              <Button size="sm" className="flex items-center gap-2 w-full sm:w-auto">
                <Plus className="h-4 w-4" />
                创建URL
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle>创建临时URL</DialogTitle>
                <DialogDescription>
                  选择创建方式：手动配置或使用模板
                </DialogDescription>
              </DialogHeader>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  const errors: any = {};
                  
                  if (createMode === 'template') {
                    if (!selectedTemplate) {
                      errors.template = '请选择一个模板';
                    }
                  } else {
                    if (!formData.image) {
                      errors.image = '请输入容器镜像';
                    }
                  }
                  
                  if (!formData.ttl_seconds) {
                    errors.ttl_seconds = '请选择过期时间';
                  }
                  
                  if (Object.keys(errors).length > 0) {
                    setFormErrors(errors);
                    return;
                  }
                  
                  handleCreateURL(formData);
                }}
                className="space-y-4"
              >
                {/* 创建模式选择 */}
                <div className="space-y-2">
                  <Label>创建方式</Label>
                  <Select value={createMode} onValueChange={(value: 'manual' | 'template') => {
                    setCreateMode(value);
                    setSelectedTemplate(null);
                    setFormData({});
                    setFormErrors({});
                  }}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="manual">手动配置</SelectItem>
                      <SelectItem value="template">使用模板</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* 模板选择 */}
                {createMode === 'template' && (
                  <div className="space-y-2">
                    <Label>选择模板 *</Label>
                    <Select 
                      value={selectedTemplate?.id || ''}
                      onValueChange={(templateId) => {
                        const template = templates.find(t => t.id === templateId);
                        if (template) {
                          setSelectedTemplate(template);
                        }
                      }}
                    >
                      <SelectTrigger className={formErrors.template ? 'border-destructive' : ''}>
                        <SelectValue placeholder="请选择模板" />
                      </SelectTrigger>
                      <SelectContent>
                        {templates.map(template => (
                          <SelectItem key={template.id} value={template.id}>
                            <div>
                              <div className="font-bold">{template.name}</div>
                              <div className="text-xs text-muted-foreground">{template.description}</div>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {formErrors.template && (
                      <p className="text-sm text-destructive">{formErrors.template}</p>
                    )}
                  </div>
                )}

                {/* 手动配置模式 */}
                {createMode === 'manual' && (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label>容器镜像 *</Label>
                      <Input
                        value={formData.image || ''}
                        onChange={(e) => {
                          setFormData(prev => ({ ...prev, image: e.target.value }));
                          if (formErrors.image) setFormErrors(prev => ({ ...prev, image: undefined }));
                        }}
                        placeholder="例如: nginx:latest"
                        className={formErrors.image ? 'border-destructive' : ''}
                      />
                      {formErrors.image && (
                        <p className="text-sm text-destructive">{formErrors.image}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label>副本数</Label>
                      <Input
                        type="number"
                        min="1"
                        max="10"
                        value={formData.replicas || 1}
                        onChange={(e) => setFormData(prev => ({ ...prev, replicas: parseInt(e.target.value) || 1 }))}
                      />
                    </div>
                  </div>
                )}

                {/* TTL配置（公共） */}
                <div className="space-y-2">
                  <Label>过期时间 *</Label>
                  <Select 
                    value={formData.ttl_seconds?.toString() || ''}
                    onValueChange={(value) => {
                      setFormData(prev => ({ ...prev, ttl_seconds: parseInt(value) }));
                      if (formErrors.ttl_seconds) setFormErrors(prev => ({ ...prev, ttl_seconds: undefined }));
                    }}
                  >
                    <SelectTrigger className={formErrors.ttl_seconds ? 'border-destructive' : ''}>
                      <SelectValue placeholder="选择过期时间" />
                    </SelectTrigger>
                    <SelectContent>
                      {ttlOptions.map(option => (
                        <SelectItem key={option.value} value={option.value.toString()}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {formErrors.ttl_seconds && (
                    <p className="text-sm text-destructive">{formErrors.ttl_seconds}</p>
                  )}
                </div>

                {/* Ingress Host配置（可选） */}
                <div className="space-y-2">
                  <Label>Ingress Host</Label>
                  <Input
                    value={formData.ingress_host || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, ingress_host: e.target.value }))}
                    placeholder="例如: myapp.example.com（可选）"
                  />
                  <p className="text-sm text-muted-foreground">
                    留空将使用系统默认的ingress host
                  </p>
                </div>

                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setCreateModalVisible(false);
                      setCreateMode('manual');
                      setSelectedTemplate(null);
                      setFormData({});
                      setFormErrors({});
                    }}
                  >
                    取消
                  </Button>
                  <Button type="submit">
                    创建URL
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Edit URL Modal */}
      <Dialog open={editModalVisible} onOpenChange={setEditModalVisible}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>编辑URL</DialogTitle>
            <DialogDescription>
              修改URL的配置信息
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {urlEditor.editedData && (
              <UnifiedEditor
                type="url"
                data={urlEditor.editedData}
                onUpdate={urlEditor.updateData}
              />
            )}
            
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditModalVisible(false);
                  setEditingURL(null);
                  urlEditor.cancelEditing();
                }}
              >
                取消
              </Button>
              <Button onClick={handleUpdateURL}>
                更新URL
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总URL数</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalUrls}</div>
            <p className="text-xs text-muted-foreground">
              当前项目的所有URL
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">活跃URL</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{stats.activeUrls}</div>
            <p className="text-xs text-muted-foreground">
              正在运行中的URL
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">等待中</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">{stats.pendingUrls}</div>
            <p className="text-xs text-muted-foreground">
              创建或等待部署中
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">失败</CardTitle>
            <Settings className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{stats.failedUrls}</div>
            <p className="text-xs text-muted-foreground">
              部署失败需要处理
            </p>
          </CardContent>
        </Card>
      </div>

      {/* URL List */}
      <Card>
        <CardHeader>
          <CardTitle>URL 列表</CardTitle>
          <CardDescription>
            管理项目中的所有临时URL
          </CardDescription>
        </CardHeader>
        <CardContent>
          {urlsLoading ? (
            <div className="flex justify-center items-center py-12">
              <Spinner size="lg" tip="加载中..." />
            </div>
          ) : urls.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Server className="mx-auto h-12 w-12 opacity-50" />
              <h3 className="mt-4 text-lg font-semibold">暂无URL</h3>
              <p className="text-sm">点击上方"创建URL"按钮开始创建您的第一个临时URL</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="min-w-[120px]">URL路径</TableHead>
                    <TableHead className="hidden md:table-cell min-w-[150px]">镜像/模板</TableHead>
                    <TableHead className="min-w-[80px]">状态</TableHead>
                    <TableHead className="hidden lg:table-cell min-w-[80px]">副本数</TableHead>
                    <TableHead className="hidden xl:table-cell min-w-[120px]">过期时间</TableHead>
                    <TableHead className="hidden xl:table-cell min-w-[120px]">创建时间</TableHead>
                    <TableHead className="min-w-[100px]">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {urls.map((url) => {
                    const config = statusConfig[url.status as keyof typeof statusConfig] || { 
                      variant: 'secondary' as const, 
                      text: url.status,
                      icon: Info
                    };
                    const StatusIcon = config.icon;
                    
                    return (
                      <TableRow key={url.id}>
                        <TableCell>
                          <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-2">
                            <code className="text-xs bg-muted px-2 py-1 rounded font-mono break-all sm:text-sm">
                              {url.path}
                            </code>
                            {url.status === 'active' && (
                              <div className="flex items-center gap-1">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => window.open(`https://example.com${url.path}`, '_blank')}
                                  className="h-6 w-6 p-0 sm:h-8 sm:w-8"
                                >
                                  <ExternalLink className="h-3 w-3" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => copyToClipboard(`https://example.com${url.path}`)}
                                  className="h-6 w-6 p-0 sm:h-8 sm:w-8"
                                >
                                  <Copy className="h-3 w-3" />
                                </Button>
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="hidden md:table-cell">
                          {url.template ? (
                            <Badge variant="secondary" className="bg-blue-100 text-blue-800">
                              <Server className="mr-1 h-3 w-3" />
                              {url.template.name}
                            </Badge>
                          ) : (
                            <span className="text-sm font-mono">{url.image}</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant={config.variant} className="flex items-center gap-1 w-fit">
                            <StatusIcon className="h-3 w-3" />
                            {config.text}
                          </Badge>
                        </TableCell>
                        <TableCell className="hidden lg:table-cell text-center">{url.replicas}</TableCell>
                        <TableCell className="hidden xl:table-cell">
                          <div className="text-xs space-y-1 sm:text-sm">
                            <div className="truncate">{formatDate(url.expire_at)}</div>
                            <div className="text-xs text-muted-foreground flex items-center gap-1">
                              <Clock className="h-3 w-3" /> 
                              <span className="truncate">{getTimeUntilExpiry(url.expire_at)}</span>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="hidden xl:table-cell">
                          <span className="text-xs text-muted-foreground sm:text-sm">
                            {formatDate(url.created_at)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 sm:gap-2">
                            {(url.status === 'draft' || url.status === 'failed') && (
                              <Button
                                size="sm"
                                onClick={() => handleDeployURL(url.id)}
                                className="flex items-center gap-1 px-2 sm:px-3"
                              >
                                <Rocket className="h-3 w-3" />
                                <span className="hidden sm:inline text-xs">
                                  {url.status === 'failed' ? '重新部署' : '部署'}
                                </span>
                              </Button>
                            )}
                            
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="outline" size="sm" className="h-8 w-8 p-0 sm:h-9 sm:w-auto sm:px-3">
                                  <MoreHorizontal className="h-4 w-4" />
                                  <span className="sr-only">更多操作</span>
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => navigate(`/urls/${url.id}`)}>
                                  <Eye className="mr-2 h-4 w-4" />
                                  查看详情
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleEditURL(url)}>
                                  <Settings className="mr-2 h-4 w-4" />
                                  编辑配置
                                </DropdownMenuItem>
                                {url.status !== 'deleted' && (
                                  <Popconfirm
                                    title="确定要删除这个URL吗？"
                                    description="删除后将无法访问"
                                    onConfirm={() => handleDeleteURL(url.id)}
                                    okText="确定"
                                    cancelText="取消"
                                  >
                                    <DropdownMenuItem className="text-destructive focus:text-destructive">
                                      <Trash2 className="mr-2 h-4 w-4" />
                                      删除URL
                                    </DropdownMenuItem>
                                  </Popconfirm>
                                )}
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default ProjectDetail;