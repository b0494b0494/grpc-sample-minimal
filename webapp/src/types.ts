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

// Upload Queue related types
export interface UploadTask {
  id: string;
  file: File;
  storageProvider: string;
  status: 'pending' | 'uploading' | 'success' | 'error' | 'cancelled';
  progress: number; // 0-100
  error?: string;
  result?: FileUploadStatus;
  createdAt: number;
  startedAt?: number;
  completedAt?: number;
}

export interface UploadQueueState {
  tasks: UploadTask[];
  activeTaskId: string | null;
  maxConcurrent: number;
}

export interface FileDownloadRequest {
  filename: string;
  storageProvider?: string;
}

export interface FileInfo {
  filename: string;
  namespace: string;
  size: number;
  uploaded_at?: number; // Unix timestamp
}

export interface FileListResponse {
  files: FileInfo[];
}

// OCR related types
export interface OCRRequest {
  filename: string;
  storage_provider: string;
}

export interface OCRResponse {
  task_id: string;
  success: boolean;
  message: string;
}

export interface OCRPage {
  page_number: number;
  text: string;
  confidence: number;
}

export interface OCRResultResponse {
  filename: string;
  engine_name: string;
  extracted_text: string;
  pages: OCRPage[];
  status: string; // "processing", "completed", "failed"
  error_message: string;
  confidence: number;
  processed_at: number; // Unix timestamp
}

export interface OCRResultSummary {
  filename: string;
  engine_name: string;
  status: string;
  processed_at: number;
}

export interface OCRListResponse {
  results: OCRResultSummary[];
}

export interface OCRComparisonResponse {
  filename: string;
  storage_provider: string;
  results: OCRResultResponse[];
}

