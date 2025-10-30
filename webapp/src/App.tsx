import React, { useState } from 'react';
import {
  Greeting,
  CounterStream,
  Chat,
  FileUpload,
  FileDownload,
} from './components';

function App() {
  const [storageProvider, setStorageProvider] = useState<string>('s3'); // Default to s3

  return (
    <div style={{ fontFamily: 'Arial, sans-serif', margin: '20px' }}>
      <div style={{ maxWidth: '600px', margin: 'auto', padding: '20px', border: '1px solid #ccc', borderRadius: '8px' }}>
        <h1>gRPC Go Sample Application</h1>

        <Greeting />
        <CounterStream />
        <Chat />

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

        <FileUpload storageProvider={storageProvider} />
        <FileDownload storageProvider={storageProvider} />
      </div>
    </div>
  );
}

export default App;
