import { GreetingResponse, ChatMessage, FileUploadStatus } from '../types';

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
  const formData = new FormData();
  formData.append('uploadFile', file);
  formData.append('storageProvider', storageProvider);

  const response = await fetch(`${API_BASE_URL}/api/upload-file`, {
    method: 'POST',
    headers: {
      'Authorization': AUTH_TOKEN,
    },
    body: formData,
  });
  return response.json();
};

export const downloadFileService = async (filename: string, storageProvider: string): Promise<Response> => {
  const response = await fetch(`${API_BASE_URL}/api/download-file?filename=${encodeURIComponent(filename)}&storageProvider=${encodeURIComponent(storageProvider)}`, {
    headers: {
      'Authorization': AUTH_TOKEN,
    },
  });
  return response;
};