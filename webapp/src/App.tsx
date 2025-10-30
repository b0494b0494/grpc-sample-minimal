import React, { useState, useEffect } from 'react';
import {
  useGreeting,
  useCounterStream,
  useChat,
  useFileUpload,
  useFileDownload,
} from './hooks';

function App() {
  const { name, setName, greeting, handleSayHello } = useGreeting();
  const { counterOutput, handleStartCounterStream } = useCounterStream();
  const { chatUser, setChatUser, chatMessageInput, setChatMessageInput, chatMessages, handleSendChatMessage } = useChat();

  const [storageProvider, setStorageProvider] = useState<string>('s3'); // Default to s3
  const { selectedFile, uploadStatus, handleFileChange, handleFileUpload } = useFileUpload(storageProvider);
  const { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload } = useFileDownload(storageProvider);

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

      <h2>File Operations</h2>
      <div style={{ marginBottom: '10px' }}>
        <label>
          Storage Provider:
          <select value={storageProvider} onChange={(e) => setStorageProvider(e.target.value)} style={{ marginLeft: '10px', padding: '5px' }}>
            <option value="s3">AWS S3 (Localstack)</option>
            {/* Add options for Azure, GCP later */}
          </select>
        </label>
      </div>

      <h3>File Upload</h3>
      <form onSubmit={handleFileUpload}>
        <input type="file" onChange={handleFileChange} />
        <button type="submit">Upload File</button>
        {uploadStatus && <p>{uploadStatus}</p>}
      </form>

      <h3>File Download</h3>
      <form onSubmit={handleFileDownload}>
        <input
          type="text"
          value={downloadFilename}
          onChange={(e) => setDownloadFilename(e.target.value)}
          placeholder="Enter filename to download"
        />
        <button type="submit">Download File</button>
        {downloadStatus && <p>{downloadStatus}</p>}
      </form>
    </div>
  );
}

export default App;