import { useCallback } from 'react';
import { useUploadQueueContext } from '../contexts/UploadQueueContext';
import { UploadTask } from '../types';

/**
 * Hook for managing upload queue operations
 */
export const useUploadQueue = () => {
  const {
    queueState,
    addTask,
    addTasks,
    updateTask,
    removeTask,
    clearQueue,
    cancelTask,
    getTask,
    getTasksByStatus,
  } = useUploadQueueContext();

  const createTask = useCallback((
    file: File,
    storageProvider: string,
    id?: string
  ): UploadTask => {
    return {
      id: id || `task-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      file,
      storageProvider,
      status: 'pending',
      progress: 0,
      createdAt: Date.now(),
    };
  }, []);

  const createTasks = useCallback((
    files: File[],
    storageProvider: string
  ): UploadTask[] => {
    return files.map((file, index) =>
      createTask(file, storageProvider, `task-${Date.now()}-${index}`)
    );
  }, [createTask]);

  const getActiveTask = useCallback((): UploadTask | undefined => {
    if (!queueState.activeTaskId) {
      return undefined;
    }
    return getTask(queueState.activeTaskId);
  }, [queueState.activeTaskId, getTask]);

  const getPendingTasks = useCallback((): UploadTask[] => {
    return getTasksByStatus('pending');
  }, [getTasksByStatus]);

  const getUploadingTasks = useCallback((): UploadTask[] => {
    return getTasksByStatus('uploading');
  }, [getTasksByStatus]);

  const getCompletedTasks = useCallback((): UploadTask[] => {
    return queueState.tasks.filter(
      task => task.status === 'success' || task.status === 'error' || task.status === 'cancelled'
    );
  }, [queueState.tasks]);

  const getSuccessfulTasks = useCallback((): UploadTask[] => {
    return getTasksByStatus('success');
  }, [getTasksByStatus]);

  const getFailedTasks = useCallback((): UploadTask[] => {
    return getTasksByStatus('error');
  }, [getTasksByStatus]);

  const getCancelledTasks = useCallback((): UploadTask[] => {
    return getTasksByStatus('cancelled');
  }, [getTasksByStatus]);

  const getQueueStats = useCallback(() => {
    const total = queueState.tasks.length;
    const pending = getPendingTasks().length;
    const uploading = getUploadingTasks().length;
    const success = getSuccessfulTasks().length;
    const error = getFailedTasks().length;
    const cancelled = getCancelledTasks().length;

    return {
      total,
      pending,
      uploading,
      success,
      error,
      cancelled,
      completed: success + error + cancelled,
    };
  }, [
    queueState.tasks.length,
    getPendingTasks,
    getUploadingTasks,
    getSuccessfulTasks,
    getFailedTasks,
    getCancelledTasks,
  ]);

  return {
    // State
    queueState,
    tasks: queueState.tasks,
    activeTaskId: queueState.activeTaskId,
    maxConcurrent: queueState.maxConcurrent,

    // Task creation
    createTask,
    createTasks,

    // Task management
    addTask,
    addTasks,
    updateTask,
    removeTask,
    cancelTask,
    clearQueue,

    // Task queries
    getTask,
    getActiveTask,
    getPendingTasks,
    getUploadingTasks,
    getCompletedTasks,
    getSuccessfulTasks,
    getFailedTasks,
    getCancelledTasks,
    getTasksByStatus,

    // Statistics
    getQueueStats,
  };
};
