import React from 'react';
import { useChat } from '../hooks';

export const Chat: React.FC = () => {
  const { chatUser, setChatUser, chatMessageInput, setChatMessageInput, chatMessages, handleSendChatMessage } = useChat();

  return (
    <section>
      <h2>Bidirectional Streaming RPC: Chat</h2>
      <div style={{ marginTop: '20px', padding: '10px', border: '1px solid #eee', backgroundColor: '#f9f9f9', borderRadius: '4px', minHeight: '150px', maxHeight: '150px', overflowY: 'scroll' }}>
        {chatMessages.map((msg, index) => (
          <p key={index} style={{ padding: '2px 0', borderBottom: '1px dotted #eee' }}>{msg}</p>
        ))}
      </div>
      <form onSubmit={handleSendChatMessage} style={{ display: 'flex', marginTop: '10px' }}>
        <input
          type="text"
          value={chatUser}
          onChange={(e) => setChatUser(e.target.value)}
          placeholder="Your Name"
          style={{ width: '100px', padding: '8px', marginRight: '10px', border: '1px solid #ccc', borderRadius: '4px' }}
        />
        <input
          type="text"
          value={chatMessageInput}
          onChange={(e) => setChatMessageInput(e.target.value)}
          placeholder="Type your message..."
          style={{ flexGrow: 1, padding: '8px', marginRight: '10px', border: '1px solid #ccc', borderRadius: '4px' }}
        />
        <button type="submit" style={{ padding: '8px 15px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
          Send
        </button>
      </form>
    </section>
  );
};
