import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Project, CreateProjectRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
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
import { 
  Plus, 
  FolderOpen, 
  Trash2, 
  RefreshCw,
  Calendar,
  FileText,
  MoreHorizontal,
  Eye,
  Settings,
  Activity,
  Server,
  Clock,
  ArrowUpRight
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const ProjectList: React.FC = () => {
  const navigate = useNavigate();
  const { toast } = useToast();
  
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    description: '',
  });
  const [formErrors, setFormErrors] = useState<any>({});

  const fetchProjects = async () => {
    setLoading(true);
    try {
      const response = await ApiService.getProjects();
      setProjects(response.projects);
    } catch (error) {
      toast({
        title: '获取失败',
        description: '获取项目列表失败',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const errors: any = {};
    if (!formData.name.trim()) {
      errors.name = '项目名称不能为空';
    }
    
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    try {
      await ApiService.createProject(formData);
      toast({
        title: '创建成功',
        description: '项目已成功创建',
      });
      setCreateModalVisible(false);
      setFormData({ name: '', description: '' });
      setFormErrors({});
      fetchProjects();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '项目创建失败';
      toast({
        title: '创建失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  const handleDeleteProject = async (projectId: string) => {
    try {
      await ApiService.deleteProject(projectId);
      toast({
        title: '删除成功',
        description: '项目已成功删除',
      });
      fetchProjects();
    } catch (error) {
      const errorMsg = (error as any)?.response?.data?.error || '项目删除失败';
      toast({
        title: '删除失败',
        description: errorMsg,
        variant: 'destructive',
      });
    }
  };

  if (loading && projects.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <Spinner size="lg" />
        <div className="text-center">
          <p className="text-lg font-medium">正在加载项目...</p>
          <p className="text-sm text-muted-foreground">请稍候</p>
        </div>
      </div>
    );
  }

  // 计算统计数据
  const stats = {
    totalProjects: projects.length,
    activeProjects: projects.filter(p => p.created_at).length,
    thisMonth: projects.filter(p => {
      const created = new Date(p.created_at);
      const now = new Date();
      return created.getMonth() === now.getMonth() && created.getFullYear() === now.getFullYear();
    }).length,
    avgUsage: 85 // 模拟数据
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">项目管理</h2>
          <p className="text-sm text-muted-foreground sm:text-base">
            管理您的所有项目和临时URL
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchProjects}
            disabled={loading}
            className="w-full sm:w-auto"
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Dialog open={createModalVisible} onOpenChange={setCreateModalVisible}>
            <DialogTrigger asChild>
              <Button size="sm" className="w-full sm:w-auto">
                <Plus className="mr-2 h-4 w-4" />
                创建项目
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>创建新项目</DialogTitle>
                <DialogDescription>
                  创建一个新的项目来管理您的临时URL
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleCreateProject} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">项目名称 *</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => {
                      setFormData(prev => ({ ...prev, name: e.target.value }));
                      if (formErrors.name) setFormErrors(prev => ({ ...prev, name: undefined }));
                    }}
                    placeholder="输入项目名称"
                    className={formErrors.name ? 'border-destructive' : ''}
                  />
                  {formErrors.name && (
                    <p className="text-sm text-destructive">{formErrors.name}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">项目描述</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    placeholder="输入项目描述（可选）"
                  />
                </div>
                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setCreateModalVisible(false);
                      setFormData({ name: '', description: '' });
                      setFormErrors({});
                    }}
                  >
                    取消
                  </Button>
                  <Button type="submit">
                    创建项目
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总项目数</CardTitle>
            <FolderOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalProjects}</div>
            <p className="text-xs text-muted-foreground">
              当前管理的所有项目
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">活跃项目</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeProjects}</div>
            <p className="text-xs text-muted-foreground">
              正在使用的项目
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">本月创建</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.thisMonth}</div>
            <p className="text-xs text-muted-foreground">
              本月新增项目
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均使用</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.avgUsage}%</div>
            <p className="text-xs text-muted-foreground">
              项目平均使用率
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Projects Grid */}
      {projects.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
              <FolderOpen className="h-10 w-10 text-muted-foreground" />
            </div>
            <div className="mt-6 text-center">
              <h3 className="text-lg font-semibold">暂无项目</h3>
              <p className="mt-2 text-sm text-muted-foreground max-w-sm">
                您还没有创建任何项目。点击上方"创建项目"按钮开始创建您的第一个项目。
              </p>
              <Button onClick={() => setCreateModalVisible(true)} className="mt-4">
                <Plus className="mr-2 h-4 w-4" />
                创建第一个项目
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 xl:grid-cols-3">
          {projects.map((project) => (
            <Card key={project.id} className="group relative overflow-hidden hover:shadow-md transition-all duration-200">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="space-y-1 flex-1">
                    <CardTitle className="text-lg line-clamp-1">{project.name}</CardTitle>
                    <CardDescription className="line-clamp-2">
                      {project.description || '暂无描述'}
                    </CardDescription>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => navigate(`/projects/${project.id}`)}>
                        <Eye className="mr-2 h-4 w-4" />
                        查看详情
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => navigate(`/projects/${project.id}/settings`)}>
                        <Settings className="mr-2 h-4 w-4" />
                        项目设置
                      </DropdownMenuItem>
                      <Popconfirm
                        title="确定要删除这个项目吗？"
                        description="删除后将无法恢复，请谨慎操作"
                        onConfirm={() => handleDeleteProject(project.id)}
                        okText="确定"
                        cancelText="取消"
                      >
                        <DropdownMenuItem className="text-destructive focus:text-destructive">
                          <Trash2 className="mr-2 h-4 w-4" />
                          删除项目
                        </DropdownMenuItem>
                      </Popconfirm>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between text-sm text-muted-foreground mb-4">
                  <div className="flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    <span className="text-xs sm:text-sm">{formatDate(project.created_at)}</span>
                  </div>
                  <Badge variant="secondary" className="text-xs w-fit">
                    活跃
                  </Badge>
                </div>
                <div className="flex flex-col gap-2 sm:flex-row">
                  <Button 
                    onClick={() => navigate(`/projects/${project.id}`)}
                    className="flex-1"
                    size="sm"
                  >
                    <span className="hidden sm:inline">查看详情</span>
                    <span className="sm:hidden">详情</span>
                    <ArrowUpRight className="ml-2 h-3 w-3" />
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => navigate(`/projects/${project.id}`)}
                    className="sm:w-auto"
                  >
                    <Plus className="h-3 w-3 sm:mr-2" />
                    <span className="hidden sm:inline">创建URL</span>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProjectList;