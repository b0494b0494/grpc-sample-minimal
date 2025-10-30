const API_BASE_URL = ''; // Proxy to Go backend
const AUTH_TOKEN = 'my-secret-token';

const headers = {
  'Content-Type': 'application/json',
  'Authorization': AUTH_TOKEN,
};

export const greetService = async (name) => {
  const response = await fetch(`${API_BASE_URL}/api/greet`, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify({ name }),
  });
  return response.json();
};

export const streamCounterService = (onMessage, onError, onEnd) => {
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

export const chatStreamService = (onMessage, onError) => {
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

export const sendChatMessageService = async (user, message) => {
  const response = await fetch(`${API_BASE_URL}/api/send-chat`, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify({ user, message }),
  });
  return response;
};

export const uploadFileService = async (file) => {
  const formData = new FormData();
  formData.append('uploadFile', file);

  const response = await fetch(`${API_BASE_URL}/api/upload-file`, {
    method: 'POST',
    headers: {
      'Authorization': AUTH_TOKEN,
    },
    body: formData,
  });
  return response.json();
};

export const downloadFileService = async (filename) => {
  const response = await fetch(`${API_BASE_URL}/api/download-file?filename=${encodeURIComponent(filename)}`, {
    headers: {
      'Authorization': AUTH_TOKEN,
    },
  });
  return response;
};
