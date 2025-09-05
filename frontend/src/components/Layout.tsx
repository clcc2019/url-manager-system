import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import { SidebarProvider } from '@/components/ui/sidebar-provider';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import {
  Menu,
  FileText,
  Home,
  FolderOpen,
  Building2,
} from 'lucide-react';
import UserMenu from './UserMenu';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  const navigationGroups = [
    {
      label: '主要功能',
      items: [
        {
          title: '项目管理',
          url: '/projects',
          icon: FolderOpen,
          isActive: location.pathname.startsWith('/projects'),
        },
        {
          title: '模版管理', 
          url: '/templates',
          icon: FileText,
          isActive: location.pathname.startsWith('/templates'),
        },
      ],
    },
  ];

  const getBreadcrumbItems = () => {
    const pathSegments = location.pathname.split('/').filter(Boolean);
    const items = [];

    // 只有当不在projects页面时才显示首页链接
    if (pathSegments[0] !== 'projects') {
      items.push({
        title: '首页',
        href: '/projects',
        icon: Home
      });
    }

    if (pathSegments[0] === 'projects') {
      items.push({
        title: '项目管理',
        href: '/projects',
        icon: FolderOpen
      });

      if (pathSegments[1] === 'new') {
        items.push({
          title: '创建项目',
          href: `/projects/new`,
          icon: FolderOpen
        });
      } else if (pathSegments[1]) {
        items.push({
          title: '项目详情',
          href: `/projects/${pathSegments[1]}`,
          icon: FolderOpen
        });
      }
    } else if (pathSegments[0] === 'templates') {
      items.push({
        title: '模版管理',
        href: '/templates',
        icon: FileText
      });
    }

    return items;
  };

  return (
    <SidebarProvider>
      {/* Application Sidebar */}
      <Sidebar variant="inset">
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>应用</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    onClick={() => navigate('/projects')}
                    className="flex items-center gap-2"
                  >
                    <Building2 className="h-4 w-4" />
                    <span>URL管理系统</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
          
          {navigationGroups.map((group) => (
            <SidebarGroup key={group.label}>
              <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {group.items.map((item) => {
                    const Icon = item.icon;
                    return (
                      <SidebarMenuItem key={item.title}>
                        <SidebarMenuButton
                          isActive={item.isActive}
                          onClick={() => navigate(item.url)}
                          className="flex items-center gap-2"
                        >
                          <Icon className="h-4 w-4" />
                          <span>{item.title}</span>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          ))}
        </SidebarContent>
      </Sidebar>

      {/* Main Content Area */}
      <SidebarInset>
        {/* Header */}
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4 flex-1 min-w-0">
            <SidebarTrigger className="-ml-1">
              <Menu className="h-4 w-4" />
            </SidebarTrigger>
            <Separator orientation="vertical" className="mr-2 h-4" />
            <Breadcrumb className="flex-1 min-w-0">
              <BreadcrumbList className="flex-wrap">
                {getBreadcrumbItems().map((item, index) => {
                  const Icon = item.icon;
                  const isLast = index === getBreadcrumbItems().length - 1;
                  return (
                    <React.Fragment key={item.href}>
                      <BreadcrumbItem className="hidden md:block">
                        {isLast ? (
                          <BreadcrumbPage className="flex items-center gap-1 truncate">
                            {Icon && <Icon className="h-4 w-4 flex-shrink-0" />}
                            <span className="truncate">{item.title}</span>
                          </BreadcrumbPage>
                        ) : (
                          <BreadcrumbLink
                            onClick={() => navigate(item.href)}
                            className="flex items-center gap-1 cursor-pointer truncate"
                          >
                            {Icon && <Icon className="h-4 w-4 flex-shrink-0" />}
                            <span className="truncate">{item.title}</span>
                          </BreadcrumbLink>
                        )}
                      </BreadcrumbItem>
                      {!isLast && <BreadcrumbSeparator className="hidden md:block" />}
                    </React.Fragment>
                  );
                })}
              </BreadcrumbList>
            </Breadcrumb>
          </div>
          <div className="px-3 flex-shrink-0">
            <UserMenu />
          </div>
        </header>
        
        {/* Page Content */}
        <div className="flex-1 space-y-4 p-4 md:p-6 lg:p-8 overflow-auto">
          <div className="mx-auto w-full max-w-none">
            {children}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
};

export default Layout;