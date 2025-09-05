import React, { useState, useEffect, useCallback, useMemo } from 'react';
import type { AppTemplate, CreateTemplateRequest, UpdateTemplateRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useAuth } from '../contexts/AuthContext';
import { useRoleCheck } from '../components/RoleGuard';
import { useToast } from '@/hooks/use-toast';
import { useBreakpoint } from '@/hooks/useBreakpoint';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
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
import { Popconfirm } from '@/components/ui/popconfirm';
import { Spinner } from '@/components/ui/spinner';
import { Separator } from '@/components/ui/separator';
import {
  Plus,
  Edit,
  Trash2,
  Eye,
  User,
  Loader2
} from 'lucide-react';

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
  const { toast } = useToast();
  const { isMobile } = useBreakpoint();

  // Modal states
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [currentTemplate, setCurrentTemplate] = useState<AppTemplate | null>(null);
  const [formData, setFormData] = useState({ name: '', description: '', yaml_spec: '' });
  const [formErrors, setFormErrors] = useState<{ name?: string; yaml_spec?: string }>({});

  const loadTemplates = useCallback(async (page = 1) => {
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
      toast({
        title: '获取失败',
        description: '获取模版列表失败',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [pageSize]);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  const handleCreate = async (values: CreateTemplateRequest) => {
    try {
      await ApiService.createTemplate(values);
      toast({
        title: '创建成功',
        description: '模版已成功创建',
      });
      setCreateModalVisible(false);
      setFormData({ name: '', description: '', yaml_spec: '' });
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版创建失败';
      toast({
        title: '创建失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleEdit = (template: AppTemplate) => {
    setCurrentTemplate(template);
    setFormData({
      name: template.name,
      description: template.description || '',
      yaml_spec: template.yaml_spec
    });
    setEditModalVisible(true);
  };

  const handleUpdate = async (values: UpdateTemplateRequest) => {
    if (!currentTemplate) return;

    try {
      await ApiService.updateTemplate(currentTemplate.id, values);
      toast({
        title: '更新成功',
        description: '模版已成功更新',
      });
      setEditModalVisible(false);
      setFormData({ name: '', description: '', yaml_spec: '' });
      setCurrentTemplate(null);
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版更新失败';
      toast({
        title: '更新失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDelete = async (templateId: string) => {
    try {
      await ApiService.deleteTemplate(templateId);
      toast({
        title: '删除成功',
        description: '模版已成功删除',
      });
      loadTemplates(currentPage);
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '模版删除失败';
      toast({
        title: '删除失败',
        description: errorMsg,
        variant: 'destructive',
      });
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

  const columns = useMemo(() => [
    {
      title: '模版名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: AppTemplate) => (
        <div className="space-y-1">
          <span className="font-medium">{text}</span>
          {record.description && (
            <p className="text-sm text-muted-foreground">({record.description})</p>
          )}
        </div>
      ),
    },
    // 所有者列（所有用户都可以看到）
    {
      title: '所有者',
      dataIndex: 'user_id',
      key: 'owner',
      render: (userId: string, record: AppTemplate) => {
        const isOwner = userId && userId === user?.id;
        return (
          <Badge 
            variant={isOwner ? 'default' : 'secondary'}
            className="inline-flex items-center gap-1"
          >
            <User className="h-3 w-3" />
            {isOwner ? '你' : userId ? `用户${userId.slice(-8)}` : '未知用户'}
          </Badge>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(text)}
        </span>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (text: string) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(text)}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: AppTemplate) => {
        // 所有用户都有所有权限
        
        return (
          <div className="flex items-center gap-2">
            {inProject ? (
              <Button
                size="sm"
                onClick={() => handleSelect(record)}
              >
                选择
              </Button>
            ) : (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePreview(record)}
                  className="inline-flex items-center gap-1"
                >
                  <Eye className="h-4 w-4" />
                  预览
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleEdit(record)}
                  className="inline-flex items-center gap-1"
                >
                  <Edit className="h-4 w-4" />
                  编辑
                </Button>
                <Popconfirm
                  title="确定要删除这个模版吗？"
                  description="删除后将无法恢复"
                  onConfirm={() => handleDelete(record.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    variant="destructive"
                    size="sm"
                    className="inline-flex items-center gap-1"
                  >
                    <Trash2 className="h-4 w-4" />
                    删除
                  </Button>
                </Popconfirm>
              </>
            )}
          </div>
        );
      },
    },
  ], [user, inProject, onSelect, handlePreview, handleEdit, handleDelete]);

  // 如果认证状态还在加载中且不是在项目内使用，显示加载指示器
  if (authLoading && !inProject) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <Spinner size="lg" tip="加载用户信息..." />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {!inProject && (
        <div className="flex justify-between items-center">
          <div>
            <h2 className={`${isMobile ? 'text-xl' : 'text-2xl'} font-bold`}>模版管理</h2>
            <p className="text-muted-foreground mt-1">
              您可以查看和管理所有用户的模版
            </p>
          </div>
          <Button
            onClick={() => setCreateModalVisible(true)}
            className="inline-flex items-center gap-2"
            size={isMobile ? 'sm' : 'default'}
          >
            <Plus className="h-4 w-4" />
            {isMobile ? '创建' : '创建模版'}
          </Button>
        </div>
      )}
      
      {inProject && (
        <div className="mb-4">
          <Button
            onClick={() => setCreateModalVisible(true)}
            className="inline-flex items-center gap-2"
            size={isMobile ? 'sm' : 'default'}
          >
            <Plus className="h-4 w-4" />
            {isMobile ? '创建' : '创建模版'}
          </Button>
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>模版名称</TableHead>
              <TableHead>所有者</TableHead>
              {!isMobile && <TableHead>创建时间</TableHead>}
              {!isMobile && <TableHead>更新时间</TableHead>}
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={isMobile ? 3 : 5} className="text-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                  <p className="mt-2 text-muted-foreground">加载中...</p>
                </TableCell>
              </TableRow>
            ) : templates.length === 0 ? (
              <TableRow>
                <TableCell colSpan={isMobile ? 3 : 5} className="text-center py-8 text-muted-foreground">
                  暂无模版数据
                </TableCell>
              </TableRow>
            ) : (
              templates.map((template) => (
                <TableRow key={template.id}>
                  <TableCell>
                    <div className="space-y-1">
                      <span className="font-medium">{template.name}</span>
                      {template.description && (
                        <p className="text-sm text-muted-foreground">({template.description})</p>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge 
                      variant={template.user_id === user?.id ? 'default' : 'secondary'}
                      className="inline-flex items-center gap-1"
                    >
                      <User className="h-3 w-3" />
                      {template.user_id === user?.id ? '你' : template.user_id ? `用户${template.user_id.slice(-8)}` : '未知用户'}
                    </Badge>
                  </TableCell>
                  {!isMobile && (
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {formatDate(template.created_at)}
                      </span>
                    </TableCell>
                  )}
                  {!isMobile && (
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {formatDate(template.updated_at)}
                      </span>
                    </TableCell>
                  )}
                  <TableCell>
                    <div className={`flex items-center ${isMobile ? 'gap-1' : 'gap-2'}`}>
                      {inProject ? (
                        <Button
                          size="sm"
                          onClick={() => handleSelect(template)}
                        >
                          选择
                        </Button>
                      ) : (
                        <>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handlePreview(template)}
                            className="inline-flex items-center gap-1"
                          >
                            <Eye className="h-4 w-4" />
                            {isMobile ? '' : '预览'}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEdit(template)}
                            className="inline-flex items-center gap-1"
                          >
                            <Edit className="h-4 w-4" />
                            {isMobile ? '' : '编辑'}
                          </Button>
                          <Popconfirm
                            title="确定要删除这个模版吗？"
                            description="删除后将无法恢复"
                            onConfirm={() => handleDelete(template.id)}
                            okText="确定"
                            cancelText="取消"
                          >
                            <Button
                              variant="destructive"
                              size="sm"
                              className="inline-flex items-center gap-1"
                            >
                              <Trash2 className="h-4 w-4" />
                              {isMobile ? '' : '删除'}
                            </Button>
                          </Popconfirm>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* 分页 */}
      {total > pageSize && (
        <div className="flex justify-center">
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadTemplates(currentPage - 1)}
              disabled={currentPage === 1}
            >
              上一页
            </Button>
            <span className="text-sm text-muted-foreground">
              第 {currentPage} 页，共 {Math.ceil(total / pageSize)} 页
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadTemplates(currentPage + 1)}
              disabled={currentPage >= Math.ceil(total / pageSize)}
            >
              下一页
            </Button>
          </div>
        </div>
      )}

      {/* 创建模版对话框 */}
      <Dialog open={createModalVisible} onOpenChange={setCreateModalVisible}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>创建应用模版</DialogTitle>
            <DialogDescription>
              创建一个新的应用模版，支持使用 ${'${PLACEHOLDER}'} 格式的占位符
            </DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!formData.name.trim()) {
                setFormErrors({ name: '请输入模版名称' });
                return;
              }
              if (!formData.yaml_spec.trim()) {
                setFormErrors({ yaml_spec: '请输入YAML规范' });
                return;
              }
              handleCreate({
                name: formData.name.trim(),
                description: formData.description.trim() || undefined,
                yaml_spec: formData.yaml_spec.trim()
              });
            }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label htmlFor="name">模版名称 *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => {
                  setFormData(prev => ({ ...prev, name: e.target.value }));
                  if (formErrors.name) setFormErrors(prev => ({ ...prev, name: undefined }));
                }}
                placeholder="请输入模版名称"
                className={formErrors.name ? 'border-destructive' : ''}
              />
              {formErrors.name && (
                <p className="text-sm text-destructive">{formErrors.name}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">描述</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="请输入模版描述（可选）"
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
                placeholder="请输入Kubernetes YAML规范，支持 ${'${PLACEHOLDER}'} 占位符"
                className={`min-h-96 w-full rounded-md border px-3 py-2 text-sm font-mono ${formErrors.yaml_spec ? 'border-destructive' : 'border-input'} bg-background`}
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
                  setFormData({ name: '', description: '', yaml_spec: '' });
                  setFormErrors({});
                }}
              >
                取消
              </Button>
              <Button type="submit">
                创建
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* 编辑模版对话框 */}
      <Dialog open={editModalVisible} onOpenChange={setEditModalVisible}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>编辑应用模版</DialogTitle>
            <DialogDescription>
              修改模版信息，支持使用 ${'${PLACEHOLDER}'} 格式的占位符
            </DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!formData.name.trim()) {
                setFormErrors({ name: '请输入模版名称' });
                return;
              }
              if (!formData.yaml_spec.trim()) {
                setFormErrors({ yaml_spec: '请输入YAML规范' });
                return;
              }
              handleUpdate({
                name: formData.name.trim(),
                description: formData.description.trim() || undefined,
                yaml_spec: formData.yaml_spec.trim()
              });
            }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label htmlFor="edit-name">模版名称 *</Label>
              <Input
                id="edit-name"
                value={formData.name}
                onChange={(e) => {
                  setFormData(prev => ({ ...prev, name: e.target.value }));
                  if (formErrors.name) setFormErrors(prev => ({ ...prev, name: undefined }));
                }}
                placeholder="请输入模版名称"
                className={formErrors.name ? 'border-destructive' : ''}
              />
              {formErrors.name && (
                <p className="text-sm text-destructive">{formErrors.name}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-description">描述</Label>
              <Input
                id="edit-description"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="请输入模版描述（可选）"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-yaml_spec">YAML规范 *</Label>
              <textarea
                id="edit-yaml_spec"
                value={formData.yaml_spec}
                onChange={(e) => {
                  setFormData(prev => ({ ...prev, yaml_spec: e.target.value }));
                  if (formErrors.yaml_spec) setFormErrors(prev => ({ ...prev, yaml_spec: undefined }));
                }}
                placeholder="请输入Kubernetes YAML规范，支持 ${'${PLACEHOLDER}'} 占位符"
                className={`min-h-96 w-full rounded-md border px-3 py-2 text-sm font-mono ${formErrors.yaml_spec ? 'border-destructive' : 'border-input'} bg-background`}
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
                  setEditModalVisible(false);
                  setFormData({ name: '', description: '', yaml_spec: '' });
                  setFormErrors({});
                  setCurrentTemplate(null);
                }}
              >
                取消
              </Button>
              <Button type="submit">
                更新
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* 预览模版对话框 */}
      <Dialog open={previewVisible} onOpenChange={setPreviewVisible}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>模版预览</DialogTitle>
            <DialogDescription>
              查看模版的详细信息和YAML规范
            </DialogDescription>
          </DialogHeader>
          {currentTemplate && (
            <div className="space-y-4">
              <Card>
                <CardContent className="pt-6">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-muted-foreground">模版名称</Label>
                      <p className="font-medium">{currentTemplate.name}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">描述</Label>
                      <p className="font-medium">{currentTemplate.description || '无'}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">创建时间</Label>
                      <p className="font-medium">{formatDate(currentTemplate.created_at)}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">更新时间</Label>
                      <p className="font-medium">{formatDate(currentTemplate.updated_at)}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <div>
                <Label className="text-muted-foreground">YAML规范</Label>
                <pre className="mt-2 w-full rounded-md bg-muted p-4 text-sm font-mono overflow-auto max-h-96">
                  {currentTemplate.yaml_spec}
                </pre>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button
              onClick={() => {
                setPreviewVisible(false);
                setCurrentTemplate(null);
              }}
            >
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default TemplateList;