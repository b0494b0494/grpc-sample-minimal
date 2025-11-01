import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { UploadTask, UploadQueueState } from '../types';

interface UploadQueueContextType {
  queueState: UploadQueueState;
  addTask: (task: UploadTask) => void;
  addTasks: (tasks: UploadTask[]) => void;
  updateTask: (taskId: string, updates: Partial<UploadTask>) => void;
  removeTask: (taskId: string) => void;
  clearQueue: () => void;
  cancelTask: (taskId: string) => void;
  getTask: (taskId: string) => UploadTask | undefined;
  getTasksByStatus: (status: UploadTask['status']) => UploadTask[];
}

const UploadQueueContext = createContext<UploadQueueContextType | undefined>(undefined);

const initialState: UploadQueueState = {
  tasks: [],
  activeTaskId: null,
  maxConcurrent: 1, // Default to sequential uploads
};

export const UploadQueueProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [queueState, setQueueState] = useState<UploadQueueState>(initialState);

  const addTask = useCallback((task: UploadTask) => {
    setQueueState(prev => ({
      ...prev,
      tasks: [...prev.tasks, task],
    }));
  }, []);

  const addTasks = useCallback((tasks: UploadTask[]) => {
    setQueueState(prev => ({
      ...prev,
      tasks: [...prev.tasks, ...tasks],
    }));
  }, []);

  const updateTask = useCallback((taskId: string, updates: Partial<UploadTask>) => {
    setQueueState(prev => {
      const updatedTasks = prev.tasks.map(task =>
        task.id === taskId ? { ...task, ...updates } : task
      );

      // Update activeTaskId if task status changed to uploading
      let activeTaskId = prev.activeTaskId;
      if (updates.status === 'uploading' && prev.activeTaskId !== taskId) {
        activeTaskId = taskId;
      } else if (
        (updates.status === 'success' || updates.status === 'error' || updates.status === 'cancelled') &&
        prev.activeTaskId === taskId
      ) {
        // Find next pending task
        const nextPendingTask = updatedTasks.find(t => t.status === 'pending');
        activeTaskId = nextPendingTask?.id || null;
      }

      return {
        ...prev,
        tasks: updatedTasks,
        activeTaskId,
      };
    });
  }, []);

  const removeTask = useCallback((taskId: string) => {
    setQueueState(prev => ({
      ...prev,
      tasks: prev.tasks.filter(task => task.id !== taskId),
      activeTaskId: prev.activeTaskId === taskId ? null : prev.activeTaskId,
    }));
  }, []);

  const clearQueue = useCallback(() => {
    setQueueState(initialState);
  }, []);

  const cancelTask = useCallback((taskId: string) => {
    setQueueState(prev => {
      const updatedTasks = prev.tasks.map(task =>
        task.id === taskId && (task.status === 'pending' || task.status === 'uploading')
          ? { ...task, status: 'cancelled' as const }
          : task
      );

      return {
        ...prev,
        tasks: updatedTasks,
        activeTaskId: prev.activeTaskId === taskId ? null : prev.activeTaskId,
      };
    });
  }, []);

  const getTask = useCallback((taskId: string) => {
    return queueState.tasks.find(task => task.id === taskId);
  }, [queueState.tasks]);

  const getTasksByStatus = useCallback((status: UploadTask['status']) => {
    return queueState.tasks.filter(task => task.status === status);
  }, [queueState.tasks]);

  const value: UploadQueueContextType = {
    queueState,
    addTask,
    addTasks,
    updateTask,
    removeTask,
    clearQueue,
    cancelTask,
    getTask,
    getTasksByStatus,
  };

  return (
    <UploadQueueContext.Provider value={value}>
      {children}
    </UploadQueueContext.Provider>
  );
};

export const useUploadQueueContext = () => {
  const context = useContext(UploadQueueContext);
  if (context === undefined) {
    throw new Error('useUploadQueueContext must be used within an UploadQueueProvider');
  }
  return context;
};
