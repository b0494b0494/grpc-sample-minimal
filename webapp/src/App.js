import React, { useState, useEffect } from 'react';
import {
  greetService,
  streamCounterService,
  chatStreamService,
  sendChatMessageService,
  uploadFileService,
  downloadFileService,
} from './services/grpcService';

function App() {
  const [name, setName] = useState('');
  const [greeting, setGreeting] = useState('');
  const [counterOutput, setCounterOutput] = useState([]);
  const [chatUser, setChatUser] = useState('Anonymous');
  const [chatMessageInput, setChatMessageInput] = useState('');
  const [chatMessages, setChatMessages] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [uploadStatus, setUploadStatus] = useState('');
  const [downloadFilename, setDownloadFilename] = useState('');
  const [downloadStatus, setDownloadStatus] = useState('');

  // Unary RPC: SayHello
  const handleSayHello = async (e) => {
    e.preventDefault();
    try {
      const data = await greetService(name);
      if (data.greeting) {
        setGreeting(data.greeting);
      } else {
        setGreeting(`Error: ${data.error || 'Unknown error'}`);
      }
    } catch (error) {
      setGreeting(`Network Error: ${error.message}`);
    }
  };

  // Server-Side Streaming RPC: StreamCounter
  const handleStartCounterStream = () => {
    setCounterOutput([]);
    const cleanup = streamCounterService(
      (data) => setCounterOutput((prev) => [...prev, `Count: ${data}`]),
      (event) => {
        console.error('EventSource failed:', event);
        setCounterOutput((prev) => [...prev, `Error in stream: ${event.data || 'Unknown error'}`]);
      },
      (data) => setCounterOutput((prev) => [...prev, data])
    );
    return cleanup; // Return cleanup function if needed
  };

  // Bidirectional Streaming RPC: Chat
  useEffect(() => {
    const cleanup = chatStreamService(
      (data) => setChatMessages((prev) => [...prev, data]),
      (event) => {
        console.error('Chat EventSource failed:', event);
        setChatMessages((prev) => [...prev, `Error in chat stream: ${event.data || 'Unknown error'}`]);
      }
    );
    return cleanup;
  }, []);

  const handleSendChatMessage = async (e) => {
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
    } catch (error) {
      console.error('Network Error sending chat message:', error);
      setChatMessages((prev) => [...prev, `Network Error: ${error.message}`]);
    }
  };

  // File Upload
  const handleFileChange = (e) => {
    setSelectedFile(e.target.files[0]);
  };

  const handleFileUpload = async (e) => {
    e.preventDefault();
    if (!selectedFile) {
      setUploadStatus('Please select a file first.');
      return;
    }

    setUploadStatus('Uploading...');
    try {
      const data = await uploadFileService(selectedFile);
      if (data.success) {
        setUploadStatus(`Upload successful: ${data.message} (${data.bytesWritten} bytes)`);
      } else {
        setUploadStatus(`Upload failed: ${data.message || 'Unknown error'}`);
      }
    } catch (error) {
      setUploadStatus(`Network Error: ${error.message}`);
    }
  };

  // File Download
  const handleFileDownload = async (e) => {
    e.preventDefault();
    if (!downloadFilename) {
      setDownloadStatus('Please enter a filename.');
      return;
    }

    setDownloadStatus('Downloading...');
    try {
      const response = await downloadFileService(downloadFilename);
      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = downloadFilename;
        document.body.appendChild(a);
        a.click();
        a.remove();
        window.URL.revokeObjectURL(url);
        setDownloadStatus('Download successful.');
      } else {
        setDownloadStatus(`Download failed: ${response.statusText}`);
      }
    } catch (error) {
      setDownloadStatus(`Network Error: ${error.message}`);
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

      <h2>File Upload</h2>
      <form onSubmit={handleFileUpload}>
        <input type="file" onChange={handleFileChange} />
        <button type="submit">Upload File</button>
        {uploadStatus && <p>{uploadStatus}</p>}
      </form>

      <h2>File Download</h2>
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