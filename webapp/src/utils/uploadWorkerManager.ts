// Web Worker Management Utility

import {
  WorkerRequestMessage,
  WorkerResponseMessage,
  UploadTaskMessage,
  CancelTaskMessage,
  ProgressUpdateMessage,
  TaskCompleteMessage,
  TaskErrorMessage,
} from '../workers/upload-worker.types';

export type WorkerMessageHandler = (message: WorkerResponseMessage) => void;

/**
 * Web Worker Manager Class
 * Manages Worker lifecycle and messaging
 */
export class UploadWorkerManager {
  private worker: Worker | null = null;
  private messageHandlers: Map<string, WorkerMessageHandler[]> = new Map();
  private initialized = false;
  private workerUrl: string | null = null;
  // Promise resolvers/rejecters for pending upload tasks
  private pendingTasks: Map<string, {
    resolve: (value?: void) => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = new Map();

  /**
   * Initialize Worker
   */
  async initialize(): Promise<void> {
    // 強制的に再初期化（キャッシュ対策）
    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
    }
    if (this.workerUrl) {
      URL.revokeObjectURL(this.workerUrl);
      this.workerUrl = null;
    }
    this.initialized = false;

    try {
      // Create Worker using Blob URL since react-scripts doesn't support direct Worker imports
      const workerCode = this.getWorkerCode();

      // Create Worker using Blob URL
      // タイムスタンプを追加してキャッシュを防ぐ
      const timestamp = Date.now();
      const workerCodeWithTimestamp = `// Worker generated at ${timestamp}\n${workerCode}`;
      const blob = new Blob([workerCodeWithTimestamp], { type: 'application/javascript' });
      this.workerUrl = URL.createObjectURL(blob);
      this.worker = new Worker(this.workerUrl, { type: 'classic' });

      // Set up message handler
      this.worker.onmessage = (event: MessageEvent<WorkerResponseMessage>) => {
        this.handleWorkerMessage(event.data);
      };

      this.worker.onerror = (error) => {
        console.error('[UploadWorkerManager] Worker error:', error);
        this.initialized = false;
        this.worker = null;
      };

      this.initialized = true;
    } catch (error) {
      console.error('Failed to initialize worker:', error);
      this.worker = null;
    }
  }

  /**
   * Get Worker code as string
   */
  private getWorkerCode(): string {
    return `
      const cancelledTasks = new Set();

      self.addEventListener('message', async (event) => {
        const message = event.data;

        switch (message.type) {
          case 'UPLOAD_TASK': {
            const { taskId, fileData, fileName, fileSize, fileType, storageProvider, apiBaseUrl, authToken } = message.payload;
            
            if (cancelledTasks.has(taskId)) {
              cancelledTasks.delete(taskId);
              return;
            }

            try {
              await uploadFile(taskId, fileData, fileName, fileSize, fileType, storageProvider, apiBaseUrl, authToken);
            } catch (error) {
              self.postMessage({
                type: 'TASK_ERROR',
                payload: {
                  taskId,
                  error: error?.message || error?.toString() || 'Unknown error',
                },
              });
            }
            break;
          }
          
          case 'CANCEL_TASK': {
            const { taskId } = message.payload;
            cancelledTasks.add(taskId);
            break;
          }
        }
      });

      async function uploadFile(taskId, fileData, fileName, fileSize, fileType, storageProvider, apiBaseUrl, authToken) {
        // Progress update (start)
        self.postMessage({
          type: 'PROGRESS_UPDATE',
          payload: {
            taskId,
            progress: 10,
          },
        });

        if (cancelledTasks.has(taskId)) {
          cancelledTasks.delete(taskId);
          return;
        }

        // Create FormData
        const formData = new FormData();
        // Create File from Blob
        const blob = new Blob([fileData], { type: fileType });
        const file = new File([blob], fileName, { type: fileType });
        formData.append('uploadFile', file, fileName);
        formData.append('storageProvider', storageProvider);

        try {
          const uploadUrl = 'http://localhost:8080/api/upload-file';
          const response = await fetch(uploadUrl, {
            method: 'POST',
            headers: {
              'Authorization': authToken,
            },
            body: formData,
          });

          if (cancelledTasks.has(taskId)) {
            cancelledTasks.delete(taskId);
            return;
          }

          if (!response.ok) {
            const errorText = await response.text();
            let errorMessage = \`HTTP \${response.status}: \${response.statusText}\`;
            
            try {
              const errorJson = JSON.parse(errorText);
              errorMessage = errorJson.error || errorJson.message || errorMessage;
            } catch {
              if (errorText) {
                errorMessage = errorText;
              }
            }
            
            throw new Error(errorMessage);
          }

          const text = await response.text();
          if (!text) {
            throw new Error('Empty response from server');
          }

          const data = JSON.parse(text);

          const isSuccess = data.success === true ||
            (data.success !== false &&
              data.message &&
              (data.message.toLowerCase().includes('uploaded') ||
                data.message.toLowerCase().includes('success')));

          // Format bytesWritten - handle both string and number from server
          let bytesWrittenStr = '0';
          if (data.bytesWritten !== undefined && data.bytesWritten !== null) {
            if (typeof data.bytesWritten === 'number') {
              bytesWrittenStr = String(data.bytesWritten);
            } else if (typeof data.bytesWritten === 'string' && data.bytesWritten !== '') {
              bytesWrittenStr = data.bytesWritten;
            }
          }

          self.postMessage({
            type: 'PROGRESS_UPDATE',
            payload: {
              taskId,
              progress: 100,
            },
          });

          self.postMessage({
            type: 'TASK_COMPLETE',
            payload: {
              taskId,
              success: isSuccess,
              result: {
                filename: data.filename || fileName,
                bytesWritten: bytesWrittenStr,
                success: isSuccess,
                message: data.message || 'File uploaded successfully',
                error: data.error,
                storageProvider: data.storageProvider || storageProvider,
              },
              error: isSuccess ? undefined : (data.message || data.error || 'Upload failed'),
            },
          });
        } catch (error) {
          if (cancelledTasks.has(taskId)) {
            cancelledTasks.delete(taskId);
            return;
          }
          throw error;
        }
      }
    `;
  }

  /**
   * Handle Worker messages
   */
  private handleWorkerMessage(message: WorkerResponseMessage): void {
    // Handle task completion/error for Promise-based uploadTask()
    if (message.type === 'TASK_COMPLETE' || message.type === 'TASK_ERROR') {
      const taskId = message.type === 'TASK_COMPLETE' 
        ? message.payload.taskId 
        : message.payload.taskId;
      
      const pendingTask = this.pendingTasks.get(taskId);
      if (pendingTask) {
        // Clear timeout
        clearTimeout(pendingTask.timeout);
        this.pendingTasks.delete(taskId);
        
        if (message.type === 'TASK_COMPLETE') {
          // Resolve even if success=false, let caller handle the result
          pendingTask.resolve();
        } else {
          // Reject on error
          const errorMsg = message.payload.error || 'Upload failed';
          pendingTask.reject(new Error(errorMsg));
        }
      }
    }
    
    // Also call registered message handlers for UI updates
    const handlers = this.messageHandlers.get(message.type) || [];
    handlers.forEach(handler => handler(message));
  }

  /**
   * Register message handler
   */
  onMessage(type: string, handler: WorkerMessageHandler): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }
    this.messageHandlers.get(type)!.push(handler);

    // Return function to remove handler
    return () => {
      const handlers = this.messageHandlers.get(type);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
        }
      }
    };
  }

  /**
   * Send upload task and wait for completion
   * Note: File objects cannot be sent directly to Worker, so we convert to ArrayBuffer
   * Returns a Promise that resolves when upload completes or rejects on error
   */
  async uploadTask(
    taskId: string,
    file: File,
    storageProvider: string,
    apiBaseUrl: string,
    authToken: string,
    timeoutMs: number = 30 * 60 * 1000 // Default 30 minutes for large files
  ): Promise<void> {
    if (!this.initialized) {
      await this.initialize();
    }

    if (!this.worker) {
      throw new Error('Worker is not available. Falling back to synchronous upload.');
    }

    // Check if task already pending
    if (this.pendingTasks.has(taskId)) {
      throw new Error(`Task ${taskId} is already in progress`);
    }

    // Convert File to ArrayBuffer
    const arrayBuffer = await file.arrayBuffer();

    if (!apiBaseUrl || (!apiBaseUrl.startsWith('http://') && !apiBaseUrl.startsWith('https://'))) {
      throw new Error(`Invalid API base URL: "${apiBaseUrl}". Must be absolute URL.`);
    }

    // Create Promise for this task
    return new Promise<void>((resolve, reject) => {
      // Set up timeout
      const timeout = setTimeout(() => {
        const pendingTask = this.pendingTasks.get(taskId);
        if (pendingTask) {
          this.pendingTasks.delete(taskId);
          reject(new Error(`Upload timeout after ${timeoutMs}ms for task ${taskId}`));
        }
      }, timeoutMs);

      // Store resolver/rejecter
      this.pendingTasks.set(taskId, { resolve, reject, timeout });

      const message: UploadTaskMessage = {
        type: 'UPLOAD_TASK',
        payload: {
          taskId,
          fileData: arrayBuffer,
          fileName: file.name,
          fileSize: file.size,
          fileType: file.type,
          storageProvider,
          apiBaseUrl,
          authToken,
        },
      };

      // Send message to Worker
      try {
        (this.worker as any).postMessage(message);
      } catch (error) {
        // Clean up on send error
        clearTimeout(timeout);
        this.pendingTasks.delete(taskId);
        reject(new Error(`Failed to send message to Worker: ${error}`));
      }
    });
  }

  /**
   * Cancel task
   */
  cancelTask(taskId: string): void {
    // Cancel pending Promise if exists
    const pendingTask = this.pendingTasks.get(taskId);
    if (pendingTask) {
      clearTimeout(pendingTask.timeout);
      pendingTask.reject(new Error('Upload cancelled'));
      this.pendingTasks.delete(taskId);
    }

    if (!this.worker) {
      return;
    }

    const message: CancelTaskMessage = {
      type: 'CANCEL_TASK',
      payload: {
        taskId,
      },
    };

    this.worker.postMessage(message);
  }

  /**
   * Terminate Worker
   */
  terminate(): void {
    // Reject all pending tasks
    this.pendingTasks.forEach(({ reject, timeout }, taskId) => {
      clearTimeout(timeout);
      reject(new Error('Worker terminated'));
    });
    this.pendingTasks.clear();

    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
      this.initialized = false;
      this.messageHandlers.clear();
      
      // Clean up Blob URL
      if (this.workerUrl) {
        URL.revokeObjectURL(this.workerUrl);
        this.workerUrl = null;
      }
    }
  }

  /**
   * Check if Worker is available
   */
  isAvailable(): boolean {
    return this.initialized && this.worker !== null;
  }
}

// Singleton instance
let workerManagerInstance: UploadWorkerManager | null = null;

/**
 * Get Worker Manager instance
 */
export function getUploadWorkerManager(): UploadWorkerManager {
  if (!workerManagerInstance) {
    workerManagerInstance = new UploadWorkerManager();
  }
  return workerManagerInstance;
}
