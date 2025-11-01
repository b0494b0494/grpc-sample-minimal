// Web Worker Message Types

export interface WorkerMessage {
  type: string;
  payload?: any;
}

export interface UploadTaskMessage extends WorkerMessage {
  type: 'UPLOAD_TASK';
  payload: {
    taskId: string;
    fileData: ArrayBuffer; // File objects cannot be sent to Worker, so use ArrayBuffer
    fileName: string;
    fileSize: number;
    fileType: string;
    storageProvider: string;
    apiBaseUrl: string;
    authToken: string;
  };
}

export interface CancelTaskMessage extends WorkerMessage {
  type: 'CANCEL_TASK';
  payload: {
    taskId: string;
  };
}

export interface ProgressUpdateMessage extends WorkerMessage {
  type: 'PROGRESS_UPDATE';
  payload: {
    taskId: string;
    progress: number;
  };
}

export interface TaskCompleteMessage extends WorkerMessage {
  type: 'TASK_COMPLETE';
  payload: {
    taskId: string;
    success: boolean;
    result?: {
      filename: string;
      bytesWritten: string;
      success: boolean;
      message: string;
      error?: string;
      storageProvider?: string;
    };
    error?: string;
  };
}

export interface TaskErrorMessage extends WorkerMessage {
  type: 'TASK_ERROR';
  payload: {
    taskId: string;
    error: string;
  };
}

export type WorkerRequestMessage = UploadTaskMessage | CancelTaskMessage;

export type WorkerResponseMessage = 
  | ProgressUpdateMessage 
  | TaskCompleteMessage 
  | TaskErrorMessage;
