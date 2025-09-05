import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { AuthProvider } from './contexts/AuthContext';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import ProjectList from './pages/ProjectList';
import ProjectDetail from './pages/ProjectDetail';
import TemplateList from './pages/TemplateList';
import LoginPage from './pages/LoginPage';
import './App.css';

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <AuthProvider>
        <Router>
          <Routes>
            {/* 公开路由 - 登录页 */}
            <Route path="/login" element={<LoginPage />} />
            
            {/* 受保护的路由 */}
            <Route path="/" element={
              <ProtectedRoute>
                <Layout>
                  <Navigate to="/projects" replace />
                </Layout>
              </ProtectedRoute>
            } />
            
            <Route path="/projects" element={
              <ProtectedRoute>
                <Layout>
                  <ProjectList />
                </Layout>
              </ProtectedRoute>
            } />
            
            <Route path="/projects/:id" element={
              <ProtectedRoute>
                <Layout>
                  <ProjectDetail />
                </Layout>
              </ProtectedRoute>
            } />
            
            <Route path="/templates" element={
              <ProtectedRoute>
                <Layout>
                  <TemplateList />
                </Layout>
              </ProtectedRoute>
            } />
            
            {/* 未匹配的路由重定向到首页 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </AuthProvider>
    </ConfigProvider>
  );
}

export default App;