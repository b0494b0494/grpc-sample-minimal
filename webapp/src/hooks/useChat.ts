import React, { useState, useEffect } from 'react';
import { chatStreamService, sendChatMessageService } from '../services/grpcService';

export const useChat = () => {
  const [chatUser, setChatUser] = useState<string>('Anonymous');
  const [chatMessageInput, setChatMessageInput] = useState<string>('');
  const [chatMessages, setChatMessages] = useState<string[]>([]);

  useEffect(() => {
    const cleanup = chatStreamService(
      (data: string) => setChatMessages((prev) => [...prev, data]),
      (event: any) => {
        console.error('Chat EventSource failed:', event);
        setChatMessages((prev) => [...prev, `Error in chat stream: ${event.data || 'Unknown error'}`]);
      }
    );
    return cleanup;
  }, []);

  const handleSendChatMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await sendChatMessageService(chatUser, chatMessageInput);
      if (!response.ok) {
        const errorData = await response.json();
        console.error('Failed to send chat message:', errorData);
        setChatMessages((prev) => [...prev, `Error sending message: ${errorData.error || response.statusText}`]);
      } else {
        setChatMessageInput(''); // Clear input after sending
      }
    } catch (error: any) {
      console.error('Network Error sending chat message:', error);
      setChatMessages((prev) => [...prev, `Network Error: ${error.message}`]);
    }
  };

  return { chatUser, setChatUser, chatMessageInput, setChatMessageInput, chatMessages, handleSendChatMessage };
};
