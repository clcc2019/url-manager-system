import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { EphemeralURL } from '../types/api.js';
import { ApiService } from '../services/api';
import { formatDate, getTimeUntilExpiry } from '../utils/date';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Spinner } from '@/components/ui/spinner';
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
  Info
} from 'lucide-react';
import { Popconfirm } from '@/components/ui/popconfirm';

const URLDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  
  const [url, setUrl] = useState<EphemeralURL | null>(null);
  const [loading, setLoading] = useState(false);

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
        description: 'URL已成功部署',
      });
      fetchURL();
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

  if (loading || !url) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <Spinner size="lg" tip="加载中..." />
      </div>
    );
  }

  const statusConfig = {
    draft: { variant: 'secondary' as const, text: '草稿', color: 'text-gray-600' },
    creating: { variant: 'default' as const, text: '创建中', color: 'text-blue-600' },
    waiting: { variant: 'outline' as const, text: '等待中', color: 'text-yellow-600' },
    active: { variant: 'default' as const, text: '运行中', color: 'text-green-600' },
    deleting: { variant: 'destructive' as const, text: '删除中', color: 'text-red-600' },
    deleted: { variant: 'secondary' as const, text: '已删除', color: 'text-gray-600' },
    failed: { variant: 'destructive' as const, text: '失败', color: 'text-red-600' },
  };
  
  const config = statusConfig[url.status as keyof typeof statusConfig] || { 
    variant: 'secondary' as const, 
    text: url.status,
    color: 'text-gray-600'
  };

  return (
    <div className="space-y-6">
      {/* 页面头部 */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            onClick={() => navigate(-1)}
            className="flex items-center gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </Button>
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">URL详情</h1>
            <p className="text-muted-foreground">
              查看和管理临时URL的详细信息
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={fetchURL}
            className="flex items-center gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          {url.status === 'active' && (
            <Button
              onClick={() => window.open(`https://example.com${url.path}`, '_blank')}
              className="flex items-center gap-2"
            >
              <Eye className="h-4 w-4" />
              访问
            </Button>
          )}
        </div>
      </div>

      {/* URL信息卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Link className="h-5 w-5" />
              URL信息
            </CardTitle>
            <Badge variant={config.variant} className={config.color}>
              {config.text}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">URL路径</div>
              <div className="flex items-center gap-2">
                <code className="text-sm bg-muted px-3 py-2 rounded-md font-mono flex-1">
                  {url.path}
                </code>
                {url.status === 'active' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(`https://example.com${url.path}`)}
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">状态</div>
              <Badge variant={config.variant} className={`${config.color} text-base px-3 py-1`}>
                {config.text}
              </Badge>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">副本数</div>
              <div className="text-lg font-semibold">{url.replicas}</div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">过期时间</div>
              <div className="space-y-1">
                <div className="text-lg font-semibold">{formatDate(url.expire_at)}</div>
                <div className="text-sm text-muted-foreground flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {getTimeUntilExpiry(url.expire_at)}
                </div>
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">创建时间</div>
              <div className="text-lg font-semibold">{formatDate(url.created_at)}</div>
            </div>
            
            <div className="space-y-2">
              <div className="text-sm font-medium text-muted-foreground">更新时间</div>
              <div className="text-lg font-semibold">{formatDate(url.updated_at)}</div>
            </div>
          </div>
          
          <Separator />
          
          {/* 镜像或模板信息 */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <Server className="h-5 w-5" />
              {url.template ? '模板信息' : '镜像信息'}
            </h3>
            
            {url.template ? (
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <div className="space-y-2">
                  <div className="font-semibold text-blue-900">{url.template.name}</div>
                  <div className="text-sm text-blue-700">{url.template.description}</div>
                </div>
              </div>
            ) : (
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                <code className="text-sm font-mono">{url.image}</code>
              </div>
            )}
          </div>
          
          {/* 环境变量 */}
          {url.env && url.env.length > 0 && (
            <>
              <Separator />
              <div className="space-y-4">
                <h3 className="text-lg font-semibold flex items-center gap-2">
                  <Settings className="h-5 w-5" />
                  环境变量
                </h3>
                <div className="grid gap-2">
                  {url.env.map((env, index) => (
                    <div key={index} className="flex items-center gap-4 bg-muted p-3 rounded-lg">
                      <code className="font-semibold">{env.name}</code>
                      <span>=</span>
                      <code className="flex-1">{env.value}</code>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* 日志信息 */}
      {url.logs && url.logs.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              创建过程日志
            </CardTitle>
            <CardDescription>
              查看URL的创建和部署过程日志
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {url.logs.map((log, index) => {
                const getLogIcon = (level: string) => {
                  switch (level) {
                    case 'error':
                      return <AlertCircle className="h-4 w-4 text-red-500" />;
                    case 'warn':
                      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
                    case 'info':
                      return <Info className="h-4 w-4 text-blue-500" />;
                    default:
                      return <CheckCircle className="h-4 w-4 text-green-500" />;
                  }
                };

                const getLogColor = (level: string) => {
                  switch (level) {
                    case 'error':
                      return 'border-red-200 bg-red-50';
                    case 'warn':
                      return 'border-yellow-200 bg-yellow-50';
                    default:
                      return 'border-gray-200 bg-gray-50';
                  }
                };

                return (
                  <div
                    key={index}
                    className={`flex items-start gap-3 p-3 rounded-lg border ${getLogColor(log.level)}`}
                  >
                    {getLogIcon(log.level)}
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-sm">{log.message}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(log.timestamp).toLocaleString()}
                        </span>
                      </div>
                      {log.details && (
                        <p className="text-sm text-muted-foreground">{log.details}</p>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 操作按钮 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            操作
          </CardTitle>
          <CardDescription>
            对此URL执行各种操作
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            {(url.status === 'draft' || url.status === 'failed') && (
              <Button
                onClick={handleDeploy}
                className="flex items-center gap-2"
              >
                <Rocket className="h-4 w-4" />
                {url.status === 'failed' ? '重新部署' : '部署'}
              </Button>
            )}
            
            {url.status === 'active' && (
              <Button
                onClick={() => window.open(`https://example.com${url.path}`, '_blank')}
                className="flex items-center gap-2"
              >
                <Eye className="h-4 w-4" />
                访问URL
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
                  className="flex items-center gap-2"
                >
                  <Trash2 className="h-4 w-4" />
                  删除URL
                </Button>
              </Popconfirm>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default URLDetail;