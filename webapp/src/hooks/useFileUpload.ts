import React, { useState, useEffect, useRef } from 'react';
import { uploadFileService } from '../services/grpcService';
import { FileUploadStatus } from '../types';
import { useUploadQueue } from './useUploadQueue';
import { getUploadWorkerManager } from '../utils/uploadWorkerManager';
import { WorkerResponseMessage } from '../workers/upload-worker.types';

interface UseFileUploadOptions {
  useWorker?: boolean; // Whether to use Web Worker (default: true)
}

export const useFileUpload = (
  storageProvider: string,
  options: UseFileUploadOptions = { useWorker: true }
) => {
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [uploadStatus, setUploadStatus] = useState<string>('');
  const [isUploading, setIsUploading] = useState(false);
  
  const {
    createTasks,
    addTasks,
    updateTask,
    queueState,
    getActiveTask,
    getQueueStats,
    cancelTask: cancelQueueTask,
  } = useUploadQueue();

  const workerManagerRef = useRef(getUploadWorkerManager());
  const unsubscribeHandlersRef = useRef<Array<() => void>>([]);

  // Initialize Worker and set up message handlers
  useEffect(() => {
    if (!options.useWorker) {
      return;
    }

    const workerManager = workerManagerRef.current;
    const unsubscribeHandlers: Array<() => void> = [];

    // Set up message handlers
    const handleProgressUpdate = (message: WorkerResponseMessage) => {
      if (message.type === 'PROGRESS_UPDATE') {
        updateTask(message.payload.taskId, {
          progress: message.payload.progress,
        });
      }
    };

    const handleTaskComplete = (message: WorkerResponseMessage) => {
      if (message.type === 'TASK_COMPLETE') {
        const { taskId, success, result, error } = message.payload;
        updateTask(taskId, {
          status: success ? 'success' : 'error',
          progress: success ? 100 : undefined,
          result: result,
          error: error,
          completedAt: Date.now(),
        });
      }
    };

    const handleTaskError = (message: WorkerResponseMessage) => {
      if (message.type === 'TASK_ERROR') {
        const { taskId, error } = message.payload;
        updateTask(taskId, {
          status: 'error',
          error: error,
          completedAt: Date.now(),
        });
      }
    };

    unsubscribeHandlers.push(
      workerManager.onMessage('PROGRESS_UPDATE', handleProgressUpdate),
      workerManager.onMessage('TASK_COMPLETE', handleTaskComplete),
      workerManager.onMessage('TASK_ERROR', handleTaskError)
    );

    unsubscribeHandlersRef.current = unsubscribeHandlers;

    // Initialize Worker when component mounts
    console.log('[useFileUpload] Initializing Worker on mount...');
    workerManager.initialize()
      .then(() => {
        console.log('[useFileUpload] Worker initialized successfully');
        console.log('[useFileUpload] Worker isAvailable:', workerManager.isAvailable());
      })
      .catch((error) => {
        console.error('[useFileUpload] Worker initialization error:', error);
      });

    return () => {
      // Cleanup: unsubscribe all handlers
      unsubscribeHandlers.forEach(unsubscribe => unsubscribe());
    };
  }, [options.useWorker, updateTask]);

  // Get API config - ????????
  const getApiConfig = () => {
    // ???: http://localhost:8080
    const API_BASE_URL = 'http://localhost:8080';
    const AUTH_TOKEN = 'my-secret-token';
    
    return { API_BASE_URL, AUTH_TOKEN };
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const filesArray = Array.from(e.target.files);
      setSelectedFiles(filesArray);
    }
  };

  const handleFileUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedFiles.length === 0) {
      setUploadStatus('Please select at least one file first.');
      return;
    }

    // Validate file names
    const invalidFiles = selectedFiles.filter(f => !f.name || f.name.trim() === '');
    if (invalidFiles.length > 0) {
      setUploadStatus('Some files have invalid filenames.');
      return;
    }

    setIsUploading(true);
    setUploadStatus(`Uploading ${selectedFiles.length} file(s)...`);

    // Create upload tasks using queue manager
    const tasks = createTasks(selectedFiles, storageProvider);
    addTasks(tasks);

    const { API_BASE_URL, AUTH_TOKEN } = getApiConfig();
    
    // ??URL???????????????
    console.log('[useFileUpload] About to upload with API_BASE_URL:', API_BASE_URL);
    if (!API_BASE_URL || (!API_BASE_URL.startsWith('http://') && !API_BASE_URL.startsWith('https://'))) {
      console.error('[useFileUpload] ERROR: API_BASE_URL is not absolute:', API_BASE_URL);
      throw new Error(`API_BASE_URL must be absolute URL. Got: "${API_BASE_URL}"`);
    }
    
    const workerManager = workerManagerRef.current;
    
    // Try to use Worker if enabled, fall back to sync if Worker unavailable
    const useWorker = options.useWorker;
    
    if (useWorker) {
      try {
        // Try to initialize Worker first
        await workerManager.initialize();
        
        if (workerManager.isAvailable()) {
          // Use Web Worker for background processing
          console.log('[useFileUpload] Passing to Worker - API_BASE_URL:', API_BASE_URL);
          await uploadWithWorker(tasks, workerManager, API_BASE_URL, AUTH_TOKEN);
        } else {
          // Worker not available, fall back to sync
          console.warn('Worker is not available, using synchronous upload');
          await uploadSynchronously(tasks);
        }
      } catch (error: any) {
        // Worker initialization failed, fall back to sync
        console.warn('Worker initialization failed, falling back to sync:', error);
        await uploadSynchronously(tasks);
      }
    } else {
      // Use synchronous upload (original implementation)
      await uploadSynchronously(tasks);
    }

    setIsUploading(false);
  };

  // Upload using Web Worker
  const uploadWithWorker = async (
    tasks: ReturnType<typeof createTasks>,
    workerManager: ReturnType<typeof getUploadWorkerManager>,
    apiBaseUrl: string,
    authToken: string
  ) => {
    // Ensure Worker is initialized
    try {
      await workerManager.initialize();
    } catch (error: any) {
      console.error('Failed to initialize Worker, falling back to sync:', error);
      await uploadSynchronously(tasks);
      return;
    }

    if (!workerManager.isAvailable()) {
      console.warn('Worker is not available, falling back to sync');
      await uploadSynchronously(tasks);
      return;
    }

    // Start all uploads in parallel (Worker handles them)
    const uploadPromises = tasks.map(async (task) => {
      // Check if task was cancelled before starting
      const currentTask = queueState.tasks.find(t => t.id === task.id);
      if (currentTask?.status === 'cancelled') {
        return;
      }

      updateTask(task.id, {
        status: 'uploading',
        startedAt: Date.now(),
        progress: 0,
      });

      try {
        // uploadTask() now returns a Promise that resolves when upload completes
        await workerManager.uploadTask(
          task.id,
          task.file,
          task.storageProvider,
          apiBaseUrl,
          authToken
        );
        
        // Check task status after upload completes (handled by message handlers)
        // The message handlers will update the task status via TASK_COMPLETE/TASK_ERROR
      } catch (error: any) {
        // Handle upload errors
        const errorMsg = error?.message || error?.toString() || 'Unknown error';
        console.error(`Worker upload failed for task ${task.id}:`, errorMsg);
        
        // Update task status to error
        updateTask(task.id, {
          status: 'error',
          error: errorMsg,
          completedAt: Date.now(),
        });
        
        // If it's a Worker initialization error, fall back to sync for this task
        if (errorMsg.includes('Worker is not available') || errorMsg.includes('Failed to initialize')) {
          console.warn(`Falling back to sync upload for task ${task.id}`);
          try {
            await uploadTaskSynchronously(task);
          } catch (syncError: any) {
            // Sync upload also failed, error already set above
            console.error(`Sync fallback also failed for task ${task.id}:`, syncError);
          }
        }
      }
    });

    // Wait for all uploads to complete (or fail)
    // Use Promise.allSettled to handle partial failures gracefully
    const results = await Promise.allSettled(uploadPromises);

    // Check for any errors
    const errors = results.filter(r => r.status === 'rejected');
    if (errors.length > 0) {
      console.warn(`${errors.length} upload task(s) had errors`);
    }

    // Update final status after all tasks complete
    // Give a small delay to allow message handlers to update task statuses
    setTimeout(() => {
      updateFinalStatus();
    }, 100);
  };

  // Upload synchronously (original implementation)
  const uploadSynchronously = async (
    tasks: ReturnType<typeof createTasks>
  ) => {
    const results: { success: number; failed: number } = { success: 0, failed: 0 };

    for (let i = 0; i < tasks.length; i++) {
      const task = tasks[i];
      
      // Check if task was cancelled
      const currentTask = queueState.tasks.find(t => t.id === task.id);
      if (currentTask?.status === 'cancelled') {
        continue;
      }

      await uploadTaskSynchronously(task, results);
    }

    // Set final status
    if (results.failed === 0) {
      setUploadStatus(`All ${results.success} file(s) uploaded successfully.`);
      setSelectedFiles([]);
    } else if (results.success === 0) {
      setUploadStatus(`All ${results.failed} file(s) failed to upload.`);
    } else {
      setUploadStatus(`${results.success} file(s) succeeded, ${results.failed} file(s) failed.`);
    }
  };

  // Upload single task synchronously
  const uploadTaskSynchronously = async (
    task: ReturnType<typeof createTasks>[0],
    results?: { success: number; failed: number }
  ) => {
    // Update task status to uploading
    updateTask(task.id, {
      status: 'uploading',
      startedAt: Date.now(),
      progress: 10,
    });

    try {
      const data: FileUploadStatus = await uploadFileService(task.file, storageProvider);
      
      // Determine success
      const isSuccess = data.success === true || 
                       (data.success !== false && 
                        data.message && 
                        (data.message.toLowerCase().includes('uploaded') || 
                         data.message.toLowerCase().includes('success')));
      
      if (results) {
        if (isSuccess) {
          results.success++;
        } else {
          results.failed++;
        }
      }

      if (isSuccess) {
        updateTask(task.id, {
          status: 'success',
          progress: 100,
          result: data,
          completedAt: Date.now(),
        });
      } else {
        const failureMessage = data.message || data.error || 'Upload failed';
        updateTask(task.id, {
          status: 'error',
          error: failureMessage,
          completedAt: Date.now(),
        });
      }
    } catch (error: any) {
      if (results) {
        results.failed++;
      }
      const errorMsg = error?.message || error?.toString() || 'Unknown error';
      updateTask(task.id, {
        status: 'error',
        error: errorMsg,
        completedAt: Date.now(),
      });
    }
  };

  // Update final status based on queue state
  const updateFinalStatus = () => {
    const stats = getQueueStats();
    if (stats.error === 0 && stats.pending === 0) {
      setUploadStatus(`All ${stats.success} file(s) uploaded successfully.`);
      setSelectedFiles([]);
    } else if (stats.success === 0 && stats.pending === 0) {
      setUploadStatus(`All ${stats.error} file(s) failed to upload.`);
    } else if (stats.pending === 0) {
      setUploadStatus(`${stats.success} file(s) succeeded, ${stats.error} file(s) failed.`);
    }
  };

  const clearSelectedFiles = () => {
    setSelectedFiles([]);
    setUploadStatus('');
  };

  const removeFile = (index: number) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
  };

  return { 
    selectedFiles, 
    uploadStatus, 
    uploadTasks: queueState.tasks,
    isUploading,
    handleFileChange, 
    handleFileUpload,
    clearSelectedFiles,
    removeFile,
  };
};
