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
}

export interface FileDownloadRequest {
  filename: string;
}
