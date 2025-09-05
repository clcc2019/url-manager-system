import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { CreateProjectRequest } from '../types/api.js';
import { ApiService } from '../services/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { useToast } from '@/hooks/use-toast';
import { Loader2 } from 'lucide-react';

const CreateProject: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({ name: '', description: '' });
  const [formErrors, setFormErrors] = useState<{ name?: string; description?: string }>({});
  const navigate = useNavigate();
  const { toast } = useToast();

  const validateForm = () => {
    const errors: { name?: string; description?: string } = {};
    
    if (!formData.name.trim()) {
      errors.name = '请输入项目名称';
    } else if (formData.name.trim().length < 1 || formData.name.trim().length > 100) {
      errors.name = '项目名称长度为1-100个字符';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    setLoading(true);
    try {
      await ApiService.createProject({
        name: formData.name.trim(),
        description: formData.description.trim() || undefined
      });
      toast({
        title: '创建成功',
        description: '项目已成功创建',
      });
      navigate('/projects');
    } catch (error) {
      const errorMsg = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '创建失败';
      toast({
        title: '创建失败',
        description: errorMsg,
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="max-w-2xl">
      <CardHeader>
        <CardTitle>创建项目</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">项目名称 *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              placeholder="请输入项目名称"
              className={formErrors.name ? 'border-destructive' : ''}
            />
            {formErrors.name && (
              <p className="text-sm text-destructive">{formErrors.name}</p>
            )}
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="description">项目描述</Label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="请输入项目描述（可选）"
              rows={4}
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
          
          <div className="flex items-center gap-3 pt-4">
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              创建项目
            </Button>
            <Button 
              type="button"
              variant="outline" 
              onClick={() => navigate('/projects')}
            >
              取消
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default CreateProject;