import React from 'react';
import { Form, Button } from 'react-bootstrap';
import { useChat } from '../hooks';

export const Chat: React.FC = () => {
  const { chatUser, setChatUser, chatMessageInput, setChatMessageInput, chatMessages, handleSendChatMessage } = useChat();

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <h2 className="h4 fw-semibold mb-3">Bidirectional Streaming RPC: Chat</h2>
      <div className="mb-3 p-3 border bg-white rounded" style={{ minHeight: '150px', maxHeight: '200px', overflowY: 'auto' }}>
        {chatMessages.length === 0 ? (
          <p className="text-muted small">No messages yet. Start chatting!</p>
        ) : (
          chatMessages.map((msg, index) => (
            <p key={index} className="py-2 border-bottom border-light last:border-0 small">{msg}</p>
          ))
        )}
      </div>
      <Form onSubmit={handleSendChatMessage} className="d-flex flex-column flex-sm-row gap-2">
        <Form.Control
          type="text"
          value={chatUser}
          onChange={(e) => setChatUser(e.target.value)}
          placeholder="Your Name"
          className="w-auto"
          style={{ width: '140px' }}
        />
        <Form.Control
          type="text"
          value={chatMessageInput}
          onChange={(e) => setChatMessageInput(e.target.value)}
          placeholder="Type your message..."
          className="flex-grow-1"
        />
        <Button variant="success" type="submit">
          Send
        </Button>
      </Form>
    </section>
  );
};
