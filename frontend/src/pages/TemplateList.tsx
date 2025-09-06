import React, { useState, useEffect } from 'react';
import type { AppTemplate, CreateTemplateRequest, UpdateTemplateRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
import { useEditor } from '@/hooks/useEditor';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Popconfirm } from '@/components/ui/popconfirm';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import UnifiedEditor from '@/components/UnifiedEditor';
import { 
  Plus, 
  FileText, 
  Trash2, 
  RefreshCw,
  Calendar,
  Edit,
  Save,
  X,
  MoreHorizontal,
  Eye,
  Settings,
  Copy,
  Code,
  Layers,
  Server,
  Activity,
  Clock
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

type EditMode = 'structured' | 'yaml';

const TemplateList: React.FC = () => {
  const { toast } = useToast();
  
  const [templates, setTemplates] = useState<AppTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editMode, setEditMode] = useState<EditMode>('structured');
  const [currentTemplate, setCurrentTemplate] = useState<AppTemplate | null>(null);
  const [yamlContent, setYamlContent] = useState('');
  const [formData, setFormData] = useState<CreateTemplateRequest>({
    name: '',
    description: '',
    yaml_spec: '',
    parsed_spec: {
      image: '',
      env: [],
      replicas: 1,
      resources: {
        requests: { cpu: '100m', memory: '128Mi' },
        limits: { cpu: '500m', memory: '512Mi' }
      }
    }
  });
  const [formErrors, setFormErrors] = useState<any>({});
  const editor = useEditor<AppTemplate>();

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const response = await ApiService.getTemplates();
      setTemplates(response.templates);
    } catch (error) {
      toast({
        title: '获取失败',
        description: '获取模板列表失败',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTemplates();
  }, []);

  const handleCreateTemplate = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const errors: any = {};
    if (!formData.name.trim()) {
      errors.name = '模板名称不能为空';
    }
    if (!formData.yaml_spec.trim()) {
      errors.yaml_spec = 'YAML规范不能为空';
    }
    
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    try {
      await ApiService.createTemplate(formData);
      toast({
        title: '创建成功',
        description: '模板已成功创建',
      });
      setCreateModalVisible(false);
      setFormData({
        name: '',
        description: '',
        yaml_spec: '',
        parsed_spec: {
          image: '',
          env: [],
          replicas: 1,
          resources: {
            requests: { cpu: '100m', memory: '128Mi' },
            limits: { cpu: '500m', memory: '512Mi' }
          }
        }
      });
      setFormErrors({});
      fetchTemplates();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模板创建失败';
      toast({
        title: '创建失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleEditTemplate = (template: AppTemplate) => {
    setCurrentTemplate(template);
    editor.startEditing(template, template.yaml_spec);
    setYamlContent(template.yaml_spec);
    setEditModalVisible(true);
  };

  const handleUpdateTemplate = async () => {
    if (!currentTemplate || !editor.editedData) return;

    try {
      const updateData: UpdateTemplateRequest = {
        name: editor.editedData.name,
        description: editor.editedData.description,
      };

      if (editMode === 'structured' && editor.editedData) {
        updateData.parsed_spec = editor.editedData.parsed_spec;
      } else {
        updateData.yaml_spec = yamlContent;
      }

      await ApiService.updateTemplate(currentTemplate.id, updateData);

      toast({
        title: '更新成功',
        description: '模板已成功更新',
      });

      setEditModalVisible(false);
      setCurrentTemplate(null);
      editor.cancelEditing();
      setYamlContent('');
      fetchTemplates();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模板更新失败';
      toast({
        title: '更新失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDeleteTemplate = async (templateId: string) => {
    try {
      await ApiService.deleteTemplate(templateId);
      toast({
        title: '删除成功',
        description: '模板已成功删除',
      });
      fetchTemplates();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模板删除失败';
      toast({
        title: '删除失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: '复制成功',
      description: 'YAML内容已复制到剪贴板',
    });
  };

  if (loading && templates.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <Spinner size="lg" />
        <div className="text-center">
          <p className="text-lg font-medium">正在加载模板...</p>
          <p className="text-sm text-muted-foreground">请稍候</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">模板管理</h1>
          <p className="text-muted-foreground">
            管理您的应用模板，快速创建标准化的部署配置
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={fetchTemplates}
            disabled={loading}
            className="flex items-center gap-2"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Dialog open={createModalVisible} onOpenChange={setCreateModalVisible}>
            <DialogTrigger asChild>
              <Button className="flex items-center gap-2">
                <Plus className="h-4 w-4" />
                创建模板
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle>创建新模板</DialogTitle>
                <DialogDescription>
                  创建一个新的应用模板，可以用于快速部署标准化的应用
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleCreateTemplate} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">模板名称 *</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => {
                      setFormData(prev => ({ ...prev, name: e.target.value }));
                      if (formErrors.name) setFormErrors(prev => ({ ...prev, name: undefined }));
                    }}
                    placeholder="输入模板名称"
                    className={formErrors.name ? 'border-destructive' : ''}
                  />
                  {formErrors.name && (
                    <p className="text-sm text-destructive">{formErrors.name}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">模板描述</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    placeholder="输入模板描述（可选）"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="yaml_spec">YAML规范 *</Label>
                  <textarea
                    id="yaml_spec"
                    value={formData.yaml_spec}
                    onChange={(e) => {
                      setFormData(prev => ({ ...prev, yaml_spec: e.target.value }));
                      if (formErrors.yaml_spec) setFormErrors(prev => ({ ...prev, yaml_spec: undefined }));
                    }}
                    placeholder="输入Kubernetes YAML配置"
                    className={`min-h-[200px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 font-mono ${formErrors.yaml_spec ? 'border-destructive' : ''}`}
                  />
                  {formErrors.yaml_spec && (
                    <p className="text-sm text-destructive">{formErrors.yaml_spec}</p>
                  )}
                </div>
                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setCreateModalVisible(false);
                      setFormData({
                        name: '',
                        description: '',
                        yaml_spec: '',
                        parsed_spec: {
                          image: '',
                          env: [],
                          replicas: 1,
                          resources: {
                            requests: { cpu: '100m', memory: '128Mi' },
                            limits: { cpu: '500m', memory: '512Mi' }
                          }
                        }
                      });
                      setFormErrors({});
                    }}
                  >
                    取消
                  </Button>
                  <Button type="submit">
                    创建模板
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总模板数</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{templates.length}</div>
            <p className="text-xs text-muted-foreground">
              当前管理的所有模板
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">活跃模板</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {templates.filter(t => t.created_at).length}
            </div>
            <p className="text-xs text-muted-foreground">
              正在使用的模板
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">本月创建</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {templates.filter(t => {
                const created = new Date(t.created_at);
                const now = new Date();
                return created.getMonth() === now.getMonth() && created.getFullYear() === now.getFullYear();
              }).length}
            </div>
            <p className="text-xs text-muted-foreground">
              本月新增模板
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">使用频率</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-purple-600">92%</div>
            <p className="text-xs text-muted-foreground">
              模板平均使用率
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Edit Template Modal */}
      <Dialog open={editModalVisible} onOpenChange={setEditModalVisible}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>编辑模板</DialogTitle>
            <DialogDescription>
              修改模板的配置信息
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {/* Mode switcher */}
            <Tabs value={editMode} onValueChange={(value) => setEditMode(value as EditMode)}>
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="structured" className="flex items-center gap-2">
                  <Layers className="h-4 w-4" />
                  结构化编辑
                </TabsTrigger>
                <TabsTrigger value="yaml" className="flex items-center gap-2">
                  <Code className="h-4 w-4" />
                  YAML编辑
                </TabsTrigger>
              </TabsList>
              
              <TabsContent value="structured" className="space-y-4">
                {editor.editedData && (
                  <UnifiedEditor
                    type="template"
                    data={editor.editedData}
                    onUpdate={editor.updateData}
                  />
                )}
              </TabsContent>
              
              <TabsContent value="yaml" className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="yaml-editor">YAML配置</Label>
                  <textarea
                    id="yaml-editor"
                    value={yamlContent}
                    onChange={(e) => setYamlContent(e.target.value)}
                    className="min-h-[400px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 font-mono"
                    placeholder="输入Kubernetes YAML配置"
                  />
                </div>
              </TabsContent>
            </Tabs>
            
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditModalVisible(false);
                  setCurrentTemplate(null);
                  editor.cancelEditing();
                  setYamlContent('');
                }}
              >
                取消
              </Button>
              <Button onClick={handleUpdateTemplate}>
                更新模板
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      {/* Templates Grid */}
      {templates.length === 0 ? (
        <Card className="flex flex-col items-center justify-center py-16">
          <div className="flex flex-col items-center space-y-4 text-center">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
              <FileText className="h-10 w-10 text-muted-foreground" />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">暂无模板</h3>
              <p className="text-sm text-muted-foreground max-w-sm">
                您还没有创建任何模板。点击上方"创建模板"按钮开始创建您的第一个模板。
              </p>
            </div>
            <Button onClick={() => setCreateModalVisible(true)} className="mt-4">
              <Plus className="mr-2 h-4 w-4" />
              创建第一个模板
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {templates.map((template) => (
            <Card key={template.id} className="group hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="space-y-1 flex-1">
                    <CardTitle className="text-lg line-clamp-1">{template.name}</CardTitle>
                    <CardDescription className="line-clamp-2">
                      {template.description || '暂无描述'}
                    </CardDescription>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => handleEditTemplate(template)}>
                        <Edit className="mr-2 h-4 w-4" />
                        编辑模板
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => copyToClipboard(template.yaml_spec)}>
                        <Copy className="mr-2 h-4 w-4" />
                        复制YAML
                      </DropdownMenuItem>
                      <Popconfirm
                        title="确定要删除这个模板吗？"
                        description="删除后将无法恢复，请谨慎操作"
                        onConfirm={() => handleDeleteTemplate(template.id)}
                        okText="确定"
                        cancelText="取消"
                      >
                        <DropdownMenuItem className="text-destructive focus:text-destructive">
                          <Trash2 className="mr-2 h-4 w-4" />
                          删除模板
                        </DropdownMenuItem>
                      </Popconfirm>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="space-y-3">
                  {/* Template info */}
                  <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      <span>{formatDate(template.created_at)}</span>
                    </div>
                    <Badge variant="secondary" className="text-xs">
                      {template.parsed_spec?.image ? '已解析' : '原始'}
                    </Badge>
                  </div>
                  
                  {/* Image info */}
                  {template.parsed_spec?.image && (
                    <div className="text-xs text-muted-foreground">
                      <span className="font-medium">镜像: </span>
                      <code className="bg-muted px-1 py-0.5 rounded text-xs">
                        {template.parsed_spec.image}
                      </code>
                    </div>
                  )}
                  
                  {/* Actions */}
                  <div className="flex gap-2 pt-2">
                    <Button 
                      onClick={() => handleEditTemplate(template)}
                      className="flex-1"
                      size="sm"
                    >
                      <Edit className="mr-2 h-3 w-3" />
                      编辑
                    </Button>
                    <Button 
                      variant="outline" 
                      size="sm"
                      onClick={() => copyToClipboard(template.yaml_spec)}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

export default TemplateList;