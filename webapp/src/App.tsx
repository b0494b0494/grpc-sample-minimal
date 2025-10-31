import React, { useState } from 'react';
import { Link, Routes, Route, Navigate } from 'react-router-dom';
import { Greeting, CounterStream, Chat, FileUpload, FileDownload } from './components';

function App() {
  const [storageProvider, setStorageProvider] = useState<string>('s3'); // Default to s3

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8 py-4">
            <Link to="/" className="text-gray-700 hover:text-primary-600 font-medium transition-colors">Home</Link>
            <Link to="/greet" className="text-gray-700 hover:text-primary-600 font-medium transition-colors">Greet</Link>
            <Link to="/counter" className="text-gray-700 hover:text-primary-600 font-medium transition-colors">Counter</Link>
            <Link to="/chat" className="text-gray-700 hover:text-primary-600 font-medium transition-colors">Chat</Link>
            <Link to="/files" className="text-gray-700 hover:text-primary-600 font-medium transition-colors">Files</Link>
          </div>
        </div>
      </nav>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white rounded-lg shadow-md p-6">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">gRPC Go Sample Application</h1>
        <Routes>
          <Route path="/" element={
            <div className="py-4">
              <p className="text-gray-600 text-lg">This sample demonstrates multiple features built on gRPC. Use the menu above to navigate.</p>
            </div>
          } />
          <Route path="/greet" element={
            <div className="py-4">
              <p className="text-gray-600 mb-4">Send your name and the gRPC server will respond with a greeting.</p>
              <Greeting />
            </div>
          } />
          <Route path="/counter" element={
            <div className="py-4">
              <p className="text-gray-600 mb-4">Receive counter values as a server stream and display them in real time.</p>
              <CounterStream />
            </div>
          } />
          <Route path="/chat" element={
            <div className="py-4">
              <p className="text-gray-600 mb-4">A simple bidirectional streaming chat. Messages are echoed via the server.</p>
              <Chat />
            </div>
          } />
          <Route path="/files" element={
            <div className="py-4">
              <h2 className="text-2xl font-semibold text-gray-900 mb-2">File Operations</h2>
              <p className="text-gray-600 mb-6">Select a file to upload/download. You can switch the destination using the Storage Provider.</p>
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Storage Provider:
                </label>
                <select 
                  value={storageProvider} 
                  onChange={(e) => setStorageProvider(e.target.value)}
                  className="block w-full max-w-xs px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 bg-white"
                >
                  <option value="s3">AWS S3 (Localstack)</option>
                  <option value="gcs">Google Cloud Storage (fake-gcs)</option>
                  <option value="azure">Azure Blob Storage (Azurite)</option>
                </select>
              </div>
              <div className="space-y-6">
                <FileUpload storageProvider={storageProvider} />
                <FileDownload storageProvider={storageProvider} />
              </div>
            </div>
          } />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        </div>
      </div>
    </div>
  );
}

export default App;
