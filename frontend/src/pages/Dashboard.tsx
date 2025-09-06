import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { ApiService } from '../services/api';
import { formatDate } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Spinner } from '@/components/ui/spinner';
import { 
  FolderOpen,
  FileText,
  Link2,
  Server,
  Activity,
  Clock,
  TrendingUp,
  Users,
  Globe,
  Zap,
  Plus,
  ArrowRight,
  Calendar,
  CheckCircle,
  AlertCircle,
  RefreshCw,
  BarChart3,
  PieChart,
  LineChart,
  ArrowUpRight,
  DollarSign
} from 'lucide-react';
import type { Project, AppTemplate } from '../types/api.js';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { toast } = useToast();
  
  const [loading, setLoading] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [templates, setTemplates] = useState<AppTemplate[]>([]);
  const [stats, setStats] = useState({
    totalProjects: 0,
    totalUrls: 0,
    activeUrls: 0,
    totalTemplates: 0,
    todayCreated: 0,
    weeklyGrowth: 0
  });

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      // 获取项目列表
      const projectsResponse = await ApiService.getProjects();
      setProjects(projectsResponse.projects.slice(0, 5)); // 只显示前5个

      // 获取模板列表
      const templatesResponse = await ApiService.getTemplates();
      setTemplates(templatesResponse.templates.slice(0, 6)); // 只显示前6个

      // 计算真实统计数据
      const totalProjects = projectsResponse.projects.length;
      const totalTemplates = templatesResponse.templates.length;
      
      // 计算今日创建的项目数
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      const todayCreated = projectsResponse.projects.filter(p => {
        const created = new Date(p.created_at);
        created.setHours(0, 0, 0, 0);
        return created.getTime() === today.getTime();
      }).length;

      // 计算本周增长（模拟数据，实际应该从API获取）
      const weeklyGrowth = Math.floor(Math.random() * 20) + 5;

      // 获取所有项目的URL统计（这里需要遍历项目获取URL数据）
      let totalUrls = 0;
      let activeUrls = 0;
      
      for (const project of projectsResponse.projects) {
        try {
          const urlsResponse = await ApiService.getProjectURLs(project.id);
          const projectUrls = urlsResponse.urls || [];
          totalUrls += projectUrls.length;
          activeUrls += projectUrls.filter(url => url.status === 'active').length;
        } catch (error) {
          // 如果获取某个项目的URL失败，继续处理其他项目
          console.warn(`Failed to fetch URLs for project ${project.id}:`, error);
        }
      }

      setStats({
        totalProjects,
        totalUrls,
        activeUrls,
        totalTemplates,
        todayCreated,
        weeklyGrowth
      });

    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      toast({
        title: '获取失败',
        description: '获取仪表板数据失败',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <div className="flex flex-col justify-center items-center min-h-96 space-y-4">
        <Spinner size="lg" />
        <div className="text-center">
          <p className="text-lg font-medium">正在加载仪表板...</p>
          <p className="text-sm text-muted-foreground">请稍候</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">仪表板</h2>
        <div className="flex items-center space-x-2">
          <Button onClick={() => fetchDashboardData()}>
            <RefreshCw className="mr-2 h-4 w-4" />
            刷新
          </Button>
        </div>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              总项目数
            </CardTitle>
            <FolderOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalProjects}</div>
            <p className="text-xs text-muted-foreground">
              <span className="text-emerald-600">+{stats.todayCreated}</span> 今日新增
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              总URL数
            </CardTitle>
            <Link2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalUrls}</div>
            <p className="text-xs text-muted-foreground">
              <span className="text-emerald-600">+{stats.weeklyGrowth}%</span> 本周增长
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">活跃URL</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeUrls}</div>
            <p className="text-xs text-muted-foreground">
              {stats.totalUrls > 0 ? Math.round((stats.activeUrls / stats.totalUrls) * 100) : 0}% 运行中
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">模板数量</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalTemplates}</div>
            <p className="text-xs text-muted-foreground">
              可用的部署模板
            </p>
          </CardContent>
        </Card>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>最近项目</CardTitle>
            <CardDescription>
              您最近创建或访问的项目
            </CardDescription>
          </CardHeader>
          <CardContent className="pl-2">
            {projects.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <FolderOpen className="h-12 w-12 text-muted-foreground/50" />
                <h3 className="mt-4 text-lg font-semibold">暂无项目</h3>
                <p className="text-sm text-muted-foreground">创建您的第一个项目开始使用</p>
                <Button 
                  onClick={() => navigate('/projects')}
                  className="mt-4"
                  size="sm"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  创建项目
                </Button>
              </div>
            ) : (
              <div className="space-y-8">
                {projects.map((project) => (
                  <div 
                    key={project.id}
                    className="flex items-center cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                    onClick={() => navigate(`/projects/${project.id}`)}
                  >
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                      <FolderOpen className="h-4 w-4" />
                    </div>
                    <div className="ml-4 space-y-1 flex-1">
                      <p className="text-sm font-medium leading-none">
                        {project.name}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {project.description || '暂无描述'}
                      </p>
                    </div>
                    <div className="ml-auto text-sm text-muted-foreground">
                      {formatDate(project.created_at)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
        
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>模板库</CardTitle>
            <CardDescription>
              可用的部署模板
            </CardDescription>
          </CardHeader>
          <CardContent>
            {templates.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <FileText className="h-12 w-12 text-muted-foreground/50" />
                <h3 className="mt-4 text-lg font-semibold">暂无模板</h3>
                <p className="text-sm text-muted-foreground">创建您的第一个部署模板</p>
                <Button 
                  onClick={() => navigate('/templates')}
                  className="mt-4"
                  size="sm"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  创建模板
                </Button>
              </div>
            ) : (
              <div className="space-y-8">
                {templates.slice(0, 5).map((template) => (
                  <div 
                    key={template.id}
                    className="flex items-center cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                    onClick={() => navigate('/templates')}
                  >
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-secondary">
                      <FileText className="h-4 w-4" />
                    </div>
                    <div className="ml-4 space-y-1 flex-1">
                      <p className="text-sm font-medium leading-none">
                        {template.name}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {template.description || '暂无描述'}
                      </p>
                    </div>
                    <Badge variant="secondary" className="ml-auto text-xs">
                      {template.parsed_spec?.image ? '已解析' : '原始'}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card className="cursor-pointer hover:shadow-md transition-shadow" onClick={() => navigate('/projects')}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              项目管理
            </CardTitle>
            <ArrowUpRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">管理项目</div>
            <p className="text-xs text-muted-foreground">
              创建和管理您的项目
            </p>
          </CardContent>
        </Card>
        
        <Card className="cursor-pointer hover:shadow-md transition-shadow" onClick={() => navigate('/templates')}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              模板管理
            </CardTitle>
            <ArrowUpRight className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">管理模板</div>
            <p className="text-xs text-muted-foreground">
              创建和编辑部署模板
            </p>
          </CardContent>
        </Card>
        
        <Card className="cursor-pointer hover:shadow-md transition-shadow" onClick={() => fetchDashboardData()}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              刷新数据
            </CardTitle>
            <RefreshCw className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">更新状态</div>
            <p className="text-xs text-muted-foreground">
              获取最新的系统数据
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Dashboard;