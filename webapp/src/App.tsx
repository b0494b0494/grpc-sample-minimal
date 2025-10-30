import React, { useState } from 'react';
import { Link, Routes, Route, Navigate } from 'react-router-dom';
import { Greeting, CounterStream, Chat, FileUpload, FileDownload } from './components';

function App() {
  const [storageProvider, setStorageProvider] = useState<string>('s3'); // Default to s3

  return (
    <div style={{ fontFamily: 'Arial, sans-serif', margin: '20px' }}>
      <nav style={{ marginBottom: '16px', borderBottom: '1px solid #ddd', paddingBottom: '8px' }}>
        <Link to="/" style={{ marginRight: 12 }}>Home</Link>
        <Link to="/greet" style={{ marginRight: 12 }}>Greet</Link>
        <Link to="/counter" style={{ marginRight: 12 }}>Counter</Link>
        <Link to="/chat" style={{ marginRight: 12 }}>Chat</Link>
        <Link to="/files">Files</Link>
      </nav>
      <div style={{ maxWidth: 760, margin: 'auto', padding: 20, border: '1px solid #ccc', borderRadius: 8 }}>
        <h1>gRPC Go Sample Application</h1>
        <Routes>
          <Route path="/" element={
            <div>
              <p>This sample demonstrates multiple features built on gRPC. Use the menu above to navigate.</p>
            </div>
          } />
          <Route path="/greet" element={
            <div>
              <p>Send your name and the gRPC server will respond with a greeting.</p>
              <Greeting />
            </div>
          } />
          <Route path="/counter" element={
            <div>
              <p>Receive counter values as a server stream and display them in real time.</p>
              <CounterStream />
            </div>
          } />
          <Route path="/chat" element={
            <div>
              <p>A simple bidirectional streaming chat. Messages are echoed via the server.</p>
              <Chat />
            </div>
          } />
          <Route path="/files" element={
            <div>
              <h2>File Operations</h2>
              <p>Select a file to upload/download. You can switch the destination using the Storage Provider.</p>
              <div style={{ marginBottom: 10 }}>
                <label>
                  Storage Provider:
                  <select value={storageProvider} onChange={(e) => setStorageProvider(e.target.value)} style={{ marginLeft: 10, padding: 5 }}>
                    <option value="s3">AWS S3 (Localstack)</option>
                    <option value="gcs">Google Cloud Storage (fake-gcs)</option>
                  </select>
                </label>
              </div>
              <FileUpload storageProvider={storageProvider} />
              <FileDownload storageProvider={storageProvider} />
            </div>
          } />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </div>
  );
}

export default App;
