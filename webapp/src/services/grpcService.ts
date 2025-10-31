import { GreetingResponse, ChatMessage, FileUploadStatus, FileListResponse } from '../types';

const API_BASE_URL = ''; // Proxy to Go backend
const AUTH_TOKEN = 'my-secret-token';

const headers = {
  'Content-Type': 'application/json',
  'Authorization': AUTH_TOKEN,
};

export const greetService = async (name: string): Promise<GreetingResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/greet`, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify({ name }),
  });
  return response.json();
};

export const streamCounterService = (onMessage: (data: string) => void, onError: (event: Event) => void, onEnd: (data: string) => void): (() => void) => {
  const eventSource = new EventSource(`${API_BASE_URL}/api/stream-counter`);

  eventSource.onmessage = (event) => {
    onMessage(event.data);
  };

  eventSource.onerror = (event) => {
    onError(event);
    eventSource.close();
  };

  eventSource.addEventListener('end', (event) => {
    onEnd(event.data);
    eventSource.close();
  });

  return () => eventSource.close();
};

export const chatStreamService = (onMessage: (data: string) => void, onError: (event: Event) => void): (() => void) => {
  const eventSource = new EventSource(`${API_BASE_URL}/api/chat-stream`);

  eventSource.onmessage = (event) => {
    onMessage(event.data);
  };

  eventSource.onerror = (event) => {
    onError(event);
    eventSource.close();
  };

  return () => eventSource.close();
};

export const sendChatMessageService = async (user: string, message: string): Promise<Response> => {
  const response = await fetch(`${API_BASE_URL}/api/send-chat`, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify({ user, message }),
  });
  return response;
};

export const uploadFileService = async (file: File, storageProvider: string): Promise<FileUploadStatus> => {
  // Validate file name
  if (!file.name || file.name.trim() === '') {
    throw new Error('File must have a valid name');
  }

  console.log('Uploading file:', { name: file.name, size: file.size, type: file.type });

  const formData = new FormData();
  formData.append('uploadFile', file, file.name); // Explicitly set filename
  formData.append('storageProvider', storageProvider);

  const response = await fetch(`${API_BASE_URL}/api/upload-file`, {
    method: 'POST',
    headers: {
      'Authorization': AUTH_TOKEN,
    },
    body: formData,
  });

  if (!response.ok) {
    const errorText = await response.text();
    let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
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

  let data: any;
  try {
    const text = await response.text();
    if (!text) {
      throw new Error('Empty response from server');
    }
    data = JSON.parse(text);
  } catch (parseError) {
    console.error('Failed to parse response:', parseError);
    throw new Error('Failed to parse server response');
  }

  // Log for debugging
  console.log('Upload response:', data);

  return {
    filename: data.filename || file.name,
    bytesWritten: data.bytesWritten || '0',
    success: data.success !== undefined ? Boolean(data.success) : (data.message ? data.message.toLowerCase().includes('uploaded') : true),
    message: data.message || 'File uploaded successfully',
    storageProvider: data.storageProvider || storageProvider,
  };
};

export const downloadFileService = async (filename: string, storageProvider: string): Promise<Response> => {
  const response = await fetch(`${API_BASE_URL}/api/download-file?filename=${encodeURIComponent(filename)}&storageProvider=${encodeURIComponent(storageProvider)}`, {
    headers: {
      'Authorization': AUTH_TOKEN,
    },
  });
  return response;
};

export const listFilesService = async (storageProvider: string): Promise<FileListResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/list-files?storageProvider=${encodeURIComponent(storageProvider)}`, {
    headers: {
      'Authorization': AUTH_TOKEN,
    },
  });
  if (!response.ok) {
    throw new Error(`Failed to list files: ${response.statusText}`);
  }
  return response.json();
};

export interface DeleteFileResponse {
  success: boolean;
  message: string;
}

export const deleteFileService = async (filename: string, storageProvider: string): Promise<DeleteFileResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/delete-file?filename=${encodeURIComponent(filename)}&storageProvider=${encodeURIComponent(storageProvider)}`, {
    method: 'DELETE',
    headers: {
      'Authorization': AUTH_TOKEN,
    },
  });
  if (!response.ok) {
    throw new Error(`Failed to delete file: ${response.statusText}`);
  }
  return response.json();
};