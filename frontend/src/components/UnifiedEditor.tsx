import React, { useState, useEffect } from 'react';
import type { AppTemplate, TemplateSpec, EphemeralURL, UpdateURLRequest } from '../types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Plus, Minus, Settings, Network, Database, Shield, Layers, Code } from 'lucide-react';

// 编辑模式类型
type EditMode = 'yaml' | 'structured';

// 编辑器类型
type EditorType = 'template' | 'url';

// 统一的编辑器属性
interface UnifiedEditorProps {
  type: EditorType;
  data: AppTemplate | EphemeralURL;
  onUpdate: (data: AppTemplate | EphemeralURL) => void;
  mode?: EditMode;
  onModeChange?: (mode: EditMode) => void;
  yamlContent?: string;
  onYamlChange?: (yaml: string) => void;
}

const UnifiedEditor: React.FC<UnifiedEditorProps> = ({
  type,
  data,
  onUpdate,
  mode = 'structured',
  onModeChange,
  yamlContent,
  onYamlChange
}) => {
  const [currentMode, setCurrentMode] = useState<EditMode>(mode);

  const handleModeChange = (newMode: EditMode) => {
    setCurrentMode(newMode);
    onModeChange?.(newMode);
  };

  // 获取规格数据
  const getSpec = (): TemplateSpec => {
    if (type === 'template') {
      return (data as AppTemplate).parsed_spec;
    } else {
      // 将URL数据转换为TemplateSpec格式
      const url = data as EphemeralURL;
      const spec = {
        image: url.image || '',
        env: url.env || [],
        replicas: url.replicas || 1,
        resources: url.resources,
        container_name: url.container_config?.container_name,
        command: url.container_config?.command,
        args: url.container_config?.args,
        working_dir: url.container_config?.working_dir,
      };
      
      return spec;
    }
  };

  // 更新规格数据
  const updateSpec = (updates: Partial<TemplateSpec>) => {
    
    if (type === 'template') {
      const template = data as AppTemplate;
      onUpdate({
        ...template,
        parsed_spec: { ...template.parsed_spec, ...updates }
      });
    } else {
      const url = data as EphemeralURL;
      const updatedURL: EphemeralURL = { ...url };
      
      // 直接更新URL对象的字段
      if (updates.image !== undefined) updatedURL.image = updates.image;
      if (updates.env !== undefined) updatedURL.env = updates.env;
      if (updates.replicas !== undefined) updatedURL.replicas = updates.replicas;
      if (updates.resources !== undefined) updatedURL.resources = updates.resources;
      
      // 处理容器配置 - 总是更新完整的container_config
      if (updates.container_name !== undefined || updates.command !== undefined || 
          updates.args !== undefined || updates.working_dir !== undefined) {
        
        // 确保container_config存在
        if (!updatedURL.container_config) {
          updatedURL.container_config = {};
        }
        
        // 更新指定的字段，保留其他字段
        if (updates.container_name !== undefined) {
          updatedURL.container_config.container_name = updates.container_name;
        }
        if (updates.command !== undefined) {
          updatedURL.container_config.command = updates.command;
        }
        if (updates.args !== undefined) {
          updatedURL.container_config.args = updates.args;
        }
        if (updates.working_dir !== undefined) {
          updatedURL.container_config.working_dir = updates.working_dir;
        }
      }
      onUpdate(updatedURL);
    }
  };

  const spec = getSpec();

  return (
    <div className="space-y-4">
      {/* 编辑模式切换 */}
      {onModeChange && (
        <div className="flex items-center space-x-2">
          <Label>编辑模式：</Label>
          <div className="flex space-x-2">
            <Button
              type="button"
              variant={currentMode === 'structured' ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleModeChange('structured')}
            >
              <Settings className="h-4 w-4 mr-2" />
              结构化编辑
            </Button>
            <Button
              type="button"
              variant={currentMode === 'yaml' ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleModeChange('yaml')}
            >
              <Code className="h-4 w-4 mr-2" />
              YAML编辑
            </Button>
          </div>
        </div>
      )}

      {currentMode === 'yaml' && onYamlChange ? (
        <YamlEditor content={yamlContent || ''} onChange={onYamlChange} />
      ) : (
        <StructuredEditor
          type={type}
          spec={spec}
          onUpdate={updateSpec}
        />
      )}
    </div>
  );
};

// YAML编辑器组件
interface YamlEditorProps {
  content: string;
  onChange: (content: string) => void;
}

const YamlEditor: React.FC<YamlEditorProps> = ({ content, onChange }) => (
  <div className="space-y-2">
    <Label htmlFor="yaml-content">YAML配置</Label>
    <textarea
      id="yaml-content"
      value={content}
      onChange={(e) => onChange(e.target.value)}
      placeholder="请输入YAML配置"
      className="min-h-96 w-full rounded-md border px-3 py-2 text-sm font-mono bg-background"
    />
  </div>
);

// 结构化编辑器组件
interface StructuredEditorProps {
  type: EditorType;
  spec: TemplateSpec;
  onUpdate: (updates: Partial<TemplateSpec>) => void;
}

const StructuredEditor: React.FC<StructuredEditorProps> = ({ type, spec, onUpdate }) => {
  const isTemplate = type === 'template';
  const [commandText, setCommandText] = useState<string>(() => (spec.command || []).join('\n'));
  const [argsText, setArgsText] = useState<string>(() => (spec.args || []).join('\n'));

  useEffect(() => {
    // 仅当外部spec与本地解析结果不一致时，才同步到文本框
    const localParsed = parseCommandInput(commandText);
    const external = JSON.stringify(spec.command || []) !== JSON.stringify(localParsed);
    if (external) {
      setCommandText((spec.command || []).join('\n'));
    }
  }, [spec.command]);

  useEffect(() => {
    const localParsed = parseCommandInput(argsText);
    const external = JSON.stringify(spec.args || []) !== JSON.stringify(localParsed);
    if (external) {
      setArgsText((spec.args || []).join('\n'));
    }
  }, [spec.args]);

  const parseShellLike = (value: string): string[] => {
    if (!value) return [];
    const parts: string[] = [];
    let current = '';
    let inQuotes = false;
    let quoteChar = '';

    for (let i = 0; i < value.length; i++) {
      const char = value[i];

      // 切换引号状态
      if ((char === '"' || char === "'") && !inQuotes) {
        inQuotes = true;
        quoteChar = char;
        current += char;
        continue;
      }
      if (char === quoteChar && inQuotes) {
        inQuotes = false;
        current += char;
        quoteChar = '';
        continue;
      }

      // 在非引号内，空白字符(空格/制表符/换行)作为分隔符
      if (!inQuotes && (char === ' ' || char === '\n' || char === '\r' || char === '\t')) {
        if (current.trim()) {
          parts.push(current.trim());
          current = '';
        }
        // 连续空白直接跳过，保持输入框中的原样由本地状态呈现
        continue;
      }

      // 其它字符直接追加
      current += char;
    }

    if (current.trim()) {
      parts.push(current.trim());
    }

    return parts;
  };

  const parseYamlLines = (value: string): string[] => {
    const lines = value.split(/\r?\n/);
    const items: string[] = [];
    for (const rawLine of lines) {
      let line = rawLine.trim();
      if (!line) continue;
      if (line.startsWith('- ')) {
        line = line.slice(2).trimStart();
      }
      // 该模式下每一行就是一个完整项（不再按空格拆分）
      if (line !== '') items.push(line);
    }
    return items;
  };

  const parseCommandInput = (value: string): string[] => {
    if (!value) return [];
    const hasYamlMarker = /(^|\n)\s*-\s+/.test(value);
    if (hasYamlMarker) {
      return parseYamlLines(value);
    }
    // 没有显式 YAML 标记时：
    // - 若包含换行，则按“每行一个项”的规则解析（不再用空格拆分）
    if (value.includes('\n')) {
      return parseYamlLines(value);
    }
    // 单行时，使用 shell-like 解析，支持引号保护空格
    return parseShellLike(value);
  };

  return (
    <Tabs defaultValue="basic" className="w-full">
      <TabsList className="grid w-full grid-cols-4">
        <TabsTrigger value="basic" className="flex items-center gap-2">
          <Layers className="h-4 w-4" />
          基础配置
        </TabsTrigger>
        <TabsTrigger value="network" className="flex items-center gap-2">
          <Network className="h-4 w-4" />
          网络端口
        </TabsTrigger>
        <TabsTrigger value="storage" className="flex items-center gap-2">
          <Database className="h-4 w-4" />
          存储卷
        </TabsTrigger>
        {isTemplate && (
          <TabsTrigger value="advanced" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            高级配置
          </TabsTrigger>
        )}
      </TabsList>

      {/* 基础配置 */}
      <TabsContent value="basic" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>容器配置</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="image">镜像地址 *</Label>
                <Input
                  id="image"
                  value={spec.image || ''}
                  onChange={(e) => onUpdate({ image: e.target.value })}
                  placeholder="nginx:latest"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="container-name">容器名称</Label>
                <Input
                  id="container-name"
                  value={spec.container_name || ''}
                  onChange={(e) => onUpdate({ container_name: e.target.value })}
                  placeholder="app"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="replicas">副本数</Label>
                <Input
                  id="replicas"
                  type="number"
                  min="1"
                  max="100"
                  value={spec.replicas || 1}
                  onChange={(e) => onUpdate({ replicas: parseInt(e.target.value) || 1 })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="working-dir">工作目录</Label>
                <Input
                  id="working-dir"
                  value={spec.working_dir || ''}
                  onChange={(e) => onUpdate({ working_dir: e.target.value })}
                  placeholder="/app"
                />
              </div>
            </div>

            <div className="space-y-4">
              {/* 启动命令 - 智能解析版本（支持空格与换行） */}
              <div className="space-y-2">
                <Label htmlFor="command">启动命令</Label>
                <textarea
                  id="command"
                  value={commandText}
                  onChange={(e) => {
                    const value = e.target.value;
                    setCommandText(value);
                    if (value === '') {
                      onUpdate({ command: [] });
                    } else {
                      onUpdate({ command: parseCommandInput(value) });
                    }
                  }}
                  placeholder={'支持两种格式:\n1) 单行: /bin/sh -c "sleep 1000"\n2) YAML列表: \n- /bin/sh\n- -c\n- "sleep 1000"'}
                  className="w-full min-h-24 px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono"
                />
              </div>

              {/* 启动参数 - 智能解析版本（支持空格与换行） */}
              <div className="space-y-2">
                <Label htmlFor="args">启动参数 (可选)</Label>
                <textarea
                  id="args"
                  value={argsText}
                  onChange={(e) => {
                    const value = e.target.value;
                    setArgsText(value);
                    if (value === '') {
                      onUpdate({ args: [] });
                    } else {
                      onUpdate({ args: parseCommandInput(value) });
                    }
                  }}
                  placeholder={'支持两种格式:\n1) 单行: --config "/app/config file.json"\n2) YAML列表: \n- --config\n- "/app/config file.json"'}
                  className="w-full min-h-20 px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 环境变量 */}
        <Card>
          <CardHeader>
            <CardTitle>环境变量</CardTitle>
          </CardHeader>
          <CardContent>
            <ArrayEditor
              data={spec.env || []}
              onChange={(env) => onUpdate({ env })}
              renderItem={(item, index, onChange, onRemove) => (
                <div key={index} className="flex space-x-2">
                  <Input
                    placeholder="变量名"
                    value={item.name}
                    onChange={(e) => onChange({ ...item, name: e.target.value })}
                  />
                  <Input
                    placeholder="变量值"
                    value={item.value}
                    onChange={(e) => onChange({ ...item, value: e.target.value })}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onRemove}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                </div>
              )}
              createNew={() => ({ name: '', value: '' })}
            />
          </CardContent>
        </Card>

        {/* 资源配置 */}
        <Card>
          <CardHeader>
            <CardTitle>资源配置</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-4">
                <h4 className="font-medium">资源请求</h4>
                <div className="space-y-2">
                  <Label htmlFor="cpu-request">CPU请求</Label>
                  <Input
                    id="cpu-request"
                    value={spec.resources?.requests?.cpu || ''}
                    onChange={(e) => onUpdate({
                      resources: {
                        ...spec.resources,
                        requests: {
                          ...spec.resources?.requests,
                          cpu: e.target.value
                        }
                      }
                    })}
                    placeholder="100m"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="memory-request">内存请求</Label>
                  <Input
                    id="memory-request"
                    value={spec.resources?.requests?.memory || ''}
                    onChange={(e) => onUpdate({
                      resources: {
                        ...spec.resources,
                        requests: {
                          ...spec.resources?.requests,
                          memory: e.target.value
                        }
                      }
                    })}
                    placeholder="128Mi"
                  />
                </div>
              </div>
              <div className="space-y-4">
                <h4 className="font-medium">资源限制</h4>
                <div className="space-y-2">
                  <Label htmlFor="cpu-limit">CPU限制</Label>
                  <Input
                    id="cpu-limit"
                    value={spec.resources?.limits?.cpu || ''}
                    onChange={(e) => onUpdate({
                      resources: {
                        ...spec.resources,
                        requests: spec.resources?.requests || { cpu: '', memory: '' },
                        limits: {
                          ...spec.resources?.limits,
                          cpu: e.target.value
                        }
                      }
                    })}
                    placeholder="500m"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="memory-limit">内存限制</Label>
                  <Input
                    id="memory-limit"
                    value={spec.resources?.limits?.memory || ''}
                    onChange={(e) => onUpdate({
                      resources: {
                        ...spec.resources,
                        requests: spec.resources?.requests || { cpu: '', memory: '' },
                        limits: {
                          ...spec.resources?.limits,
                          memory: e.target.value
                        }
                      }
                    })}
                    placeholder="256Mi"
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* 网络端口配置 */}
      <TabsContent value="network" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>端口配置</CardTitle>
          </CardHeader>
          <CardContent>
            <ArrayEditor
              data={spec.ports || []}
              onChange={(ports) => onUpdate({ ports })}
              renderItem={(item, index, onChange, onRemove) => (
                <div key={index} className="grid grid-cols-4 gap-2">
                  <Input
                    placeholder="端口名称"
                    value={item.name || ''}
                    onChange={(e) => onChange({ ...item, name: e.target.value })}
                  />
                  <Input
                    type="number"
                    placeholder="端口号"
                    value={item.container_port}
                    onChange={(e) => onChange({ ...item, container_port: parseInt(e.target.value) || 0 })}
                  />
                  <select
                    value={item.protocol || 'TCP'}
                    onChange={(e) => onChange({ ...item, protocol: e.target.value })}
                    className="rounded-md border px-3 py-2 text-sm"
                  >
                    <option value="TCP">TCP</option>
                    <option value="UDP">UDP</option>
                    <option value="SCTP">SCTP</option>
                  </select>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onRemove}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                </div>
              )}
              createNew={() => ({ container_port: 8080, protocol: 'TCP' })}
            />
          </CardContent>
        </Card>
      </TabsContent>

      {/* 存储卷配置 */}
      <TabsContent value="storage" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>卷挂载</CardTitle>
          </CardHeader>
          <CardContent>
            <ArrayEditor
              data={spec.volume_mounts || []}
              onChange={(volume_mounts) => onUpdate({ volume_mounts })}
              renderItem={(item, index, onChange, onRemove) => (
                <div key={index} className="grid grid-cols-5 gap-2">
                  <Input
                    placeholder="卷名称"
                    value={item.name}
                    onChange={(e) => onChange({ ...item, name: e.target.value })}
                  />
                  <Input
                    placeholder="挂载路径"
                    value={item.mount_path}
                    onChange={(e) => onChange({ ...item, mount_path: e.target.value })}
                  />
                  <Input
                    placeholder="子路径"
                    value={item.sub_path || ''}
                    onChange={(e) => onChange({ ...item, sub_path: e.target.value })}
                  />
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={item.read_only || false}
                      onChange={(e) => onChange({ ...item, read_only: e.target.checked })}
                    />
                    <Label className="text-sm">只读</Label>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onRemove}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                </div>
              )}
              createNew={() => ({ name: '', mount_path: '', sub_path: '', read_only: false })}
            />
          </CardContent>
        </Card>
      </TabsContent>

      {/* 高级配置（仅模板） */}
      {isTemplate && (
        <TabsContent value="advanced" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Deployment配置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="deployment-name">Deployment名称</Label>
                  <Input
                    id="deployment-name"
                    value={spec.deployment_name || ''}
                    onChange={(e) => onUpdate({ deployment_name: e.target.value })}
                    placeholder="my-app"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="namespace">命名空间</Label>
                  <Input
                    id="namespace"
                    value={spec.namespace || ''}
                    onChange={(e) => onUpdate({ namespace: e.target.value })}
                    placeholder="default"
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      )}
    </Tabs>
  );
};

// 数组编辑器组件
interface ArrayEditorProps<T> {
  data: T[];
  onChange: (data: T[]) => void;
  renderItem: (item: T, index: number, onChange: (item: T) => void, onRemove: () => void) => React.ReactNode;
  createNew: () => T;
}

function ArrayEditor<T>({ data, onChange, renderItem, createNew }: ArrayEditorProps<T>) {
  const updateItem = (index: number, item: T) => {
    const newData = [...data];
    newData[index] = item;
    onChange(newData);
  };

  const removeItem = (index: number) => {
    const newData = data.filter((_, i) => i !== index);
    onChange(newData);
  };

  const addItem = () => {
    onChange([...data, createNew()]);
  };

  return (
    <div className="space-y-2">
      {data.map((item, index) => 
        renderItem(item, index, (newItem) => updateItem(index, newItem), () => removeItem(index))
      )}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={addItem}
        className="w-full"
      >
        <Plus className="h-4 w-4 mr-2" />
        添加
      </Button>
    </div>
  );
}

export default UnifiedEditor;
