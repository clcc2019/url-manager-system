import { useState } from 'react';

export type EditMode = 'yaml' | 'structured';

export interface UseEditorOptions {
  defaultMode?: EditMode;
}

export function useEditor<T>(options: UseEditorOptions = {}) {
  const [editing, setEditing] = useState(false);
  const [editMode, setEditMode] = useState<EditMode>(options.defaultMode || 'structured');
  const [originalData, setOriginalData] = useState<T | null>(null);
  const [editedData, setEditedData] = useState<T | null>(null);
  const [yamlContent, setYamlContent] = useState('');

  const startEditing = (data: T, yaml?: string) => {
    setOriginalData(data);
    setEditedData(JSON.parse(JSON.stringify(data))); // 深拷贝
    setYamlContent(yaml || '');
    setEditing(true);
  };

  const cancelEditing = () => {
    setEditing(false);
    setOriginalData(null);
    setEditedData(null);
    setYamlContent('');
  };

  const updateData = (data: T) => {
    setEditedData(data);
  };

  const updateYaml = (yaml: string) => {
    setYamlContent(yaml);
  };

  const changeMode = (mode: EditMode) => {
    setEditMode(mode);
  };

  return {
    // 状态
    editing,
    editMode,
    originalData,
    editedData,
    yamlContent,
    
    // 操作
    startEditing,
    cancelEditing,
    updateData,
    updateYaml,
    changeMode,
  };
}
