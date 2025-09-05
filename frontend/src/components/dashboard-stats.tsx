import React from 'react';
import { Activity, FolderOpen, Clock, TrendingUp } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface DashboardStatsProps {
  stats?: {
    totalProjects: number;
    activeUrls: number;
    recentActivity: number;
    successRate: number;
  };
}

export function DashboardStats({ stats = {
  totalProjects: 0,
  activeUrls: 0,
  recentActivity: 0,
  successRate: 100
} }: DashboardStatsProps) {
  const statCards = [
    {
      title: "总项目数",
      value: stats.totalProjects,
      description: "项目总数量",
      icon: FolderOpen,
      trend: stats.totalProjects > 0 ? `当前有 ${stats.totalProjects} 个项目` : "还没有项目"
    },
    {
      title: "活跃URL",
      value: stats.activeUrls,
      description: "当前活跃的URL数量",
      icon: Activity,
      trend: stats.activeUrls > 0 ? `正在运行 ${stats.activeUrls} 个URL` : "暂无活跃URL"
    },
    {
      title: "最近活动",
      value: stats.recentActivity,
      description: "过去24小时的活动",
      icon: Clock,
      trend: stats.recentActivity > 0 ? `${stats.recentActivity} 个项目有更新` : "暂无最近活动"
    },
    {
      title: "成功率",
      value: `${stats.successRate}%`,
      description: "URL访问成功率",
      icon: TrendingUp,
      trend: stats.successRate >= 95 ? "系统运行稳定" : "需要关注系统状态"
    }
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {statCards.map((card, index) => {
        const Icon = card.icon;
        return (
          <Card key={index}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {card.title}
              </CardTitle>
              <Icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
              <p className="text-xs text-muted-foreground">
                {card.trend}
              </p>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}