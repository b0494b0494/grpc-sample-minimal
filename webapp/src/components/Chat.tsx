import React from 'react';
import { useChat } from '../hooks';

export const Chat: React.FC = () => {
  const { chatUser, setChatUser, chatMessageInput, setChatMessageInput, chatMessages, handleSendChatMessage } = useChat();

  return (
    <section className="bg-gray-50 rounded-lg p-6 border border-gray-200">
      <h2 className="text-xl font-semibold text-gray-900 mb-4">Bidirectional Streaming RPC: Chat</h2>
      <div className="mb-4 p-4 border border-gray-300 bg-white rounded-md min-h-[150px] max-h-[200px] overflow-y-auto">
        {chatMessages.length === 0 ? (
          <p className="text-gray-400 text-sm">No messages yet. Start chatting!</p>
        ) : (
          chatMessages.map((msg, index) => (
            <p key={index} className="py-2 border-b border-gray-100 last:border-0 text-sm text-gray-700">{msg}</p>
          ))
        )}
      </div>
      <form onSubmit={handleSendChatMessage} className="flex gap-3">
        <input
          type="text"
          value={chatUser}
          onChange={(e) => setChatUser(e.target.value)}
          placeholder="Your Name"
          className="w-32 px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500"
        />
        <input
          type="text"
          value={chatMessageInput}
          onChange={(e) => setChatMessageInput(e.target.value)}
          placeholder="Type your message..."
          className="flex-1 px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500"
        />
        <button 
          type="submit"
          className="px-6 py-2 bg-green-600 text-white font-medium rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 transition-colors"
        >
          Send
        </button>
      </form>
    </section>
  );
};
