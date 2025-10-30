import React, { useState, useEffect } from 'react';

const API_BASE_URL = ''; // Proxy to Go backend
const AUTH_TOKEN = 'my-secret-token';

function App() {
  const [name, setName] = useState('');
  const [greeting, setGreeting] = useState('');
  const [counterOutput, setCounterOutput] = useState([]);
  const [chatUser, setChatUser] = useState('Anonymous');
  const [chatMessageInput, setChatMessageInput] = useState('');
  const [chatMessages, setChatMessages] = useState([]);

  // Unary RPC: SayHello
  const handleSayHello = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_BASE_URL}/api/greet`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name }),
      });
      const data = await response.json();
      if (response.ok) {
        setGreeting(data.greeting);
      } else {
        setGreeting(`Error: ${data.error || response.statusText}`);
      }
    } catch (error) {
      setGreeting(`Network Error: ${error.message}`);
    }
  };

  // Server-Side Streaming RPC: StreamCounter
  const handleStartCounterStream = () => {
    setCounterOutput([]);
    const eventSource = new EventSource(`${API_BASE_URL}/api/stream-counter`);

    eventSource.onmessage = (event) => {
      setCounterOutput((prev) => [...prev, `Count: ${event.data}`]);
    };

    eventSource.onerror = (event) => {
      console.error('EventSource failed:', event);
      setCounterOutput((prev) => [...prev, `Error in stream: ${event.data || 'Unknown error'}`]);
      eventSource.close();
    };

    eventSource.addEventListener('end', (event) => {
      setCounterOutput((prev) => [...prev, event.data]);
      eventSource.close();
    });
  };

  // Bidirectional Streaming RPC: Chat
  useEffect(() => {
    const chatEventSource = new EventSource(`${API_BASE_URL}/api/chat-stream`);

    chatEventSource.onmessage = (event) => {
      setChatMessages((prev) => [...prev, event.data]);
    };

    chatEventSource.onerror = (event) => {
      console.error('Chat EventSource failed:', event);
      setChatMessages((prev) => [...prev, `Error in chat stream: ${event.data || 'Unknown error'}`]);
      chatEventSource.close();
    };

    return () => {
      chatEventSource.close();
    };
  }, []);

  const handleSendChatMessage = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_BASE_URL}/api/send-chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user: chatUser, message: chatMessageInput }),
      });
      if (!response.ok) {
        const errorData = await response.json();
        console.error('Failed to send chat message:', errorData);
        setChatMessages((prev) => [...prev, `Error sending message: ${errorData.error || response.statusText}`]);
      } else {
        setChatMessageInput(''); // Clear input after sending
      }
    } catch (error) {
      console.error('Network Error sending chat message:', error);
      setChatMessages((prev) => [...prev, `Network Error: ${error.message}`]);
    }
  };

  return (
    <div style={{ fontFamily: 'Arial, sans-serif', margin: '20px' }}>
      <div style={{ maxWidth: '600px', margin: 'auto', padding: '20px', border: '1px solid #ccc', borderRadius: '8px' }}>
        <h1>gRPC Go Sample Application</h1>

        <h2>Unary RPC: SayHello</h2>
        <form onSubmit={handleSayHello}>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
            style={{ width: 'calc(100% - 100px)', padding: '8px', marginRight: '10px', border: '1px solid #ccc', borderRadius: '4px' }}
          />
          <button type="submit" style={{ padding: '8px 15px', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
            Say Hello
          </button>
        </form>
        {greeting && <p style={{ marginTop: '20px', fontWeight: 'bold', color: greeting.startsWith('Error') || greeting.startsWith('Network Error') ? '#dc3545' : '#28a745' }}>{greeting}</p>}

        <h2>Server-Side Streaming RPC: StreamCounter</h2>
        <button onClick={handleStartCounterStream} style={{ padding: '8px 15px', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
          Start Counter Stream
        </button>
        <div style={{ marginTop: '20px', padding: '10px', border: '1px solid #eee', backgroundColor: '#f9f9f9', borderRadius: '4px', minHeight: '50px' }}>
          {counterOutput.map((item, index) => (
            <p key={index} style={{ padding: '5px 0', borderBottom: '1px dotted #ddd' }}>{item}</p>
          ))}
        </div>

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
      </div>
    </div>
  );
}

export default App;
