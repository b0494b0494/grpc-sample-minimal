export interface GreetingResponse {
  greeting: string;
  error?: string;
}

export interface ChatMessage {
  user: string;
  message: string;
}

export interface FileUploadStatus {
  filename: string;
  bytesWritten: string;
  success: boolean;
  message: string;
  error?: string;
  storageProvider?: string;
}

export interface FileDownloadRequest {
  filename: string;
  storageProvider?: string;
}

export interface FileInfo {
  filename: string;
  namespace: string;
  size: number;
}

export interface FileListResponse {
  files: FileInfo[];
}

