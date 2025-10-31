import React, { useState } from 'react';
import { Link, Routes, Route, Navigate } from 'react-router-dom';
import { Container, Navbar, Nav } from 'react-bootstrap';
import { Greeting, CounterStream, Chat, FileUpload, FileDownload, OCRResults } from './components';

function App() {
  const [storageProvider, setStorageProvider] = useState<string>('s3'); // Default to s3

  return (
    <div className="min-h-screen bg-gray-50 w-full">
      <Navbar bg="white" expand="lg" className="shadow-sm border-b border-gray-200 sticky top-0 z-10">
        <Container fluid="lg">
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse id="basic-navbar-nav">
            <Nav className="mx-auto">
              <Nav.Link as={Link} to="/" className="px-3">Home</Nav.Link>
              <Nav.Link as={Link} to="/greet" className="px-3">Greet</Nav.Link>
              <Nav.Link as={Link} to="/counter" className="px-3">Counter</Nav.Link>
              <Nav.Link as={Link} to="/chat" className="px-3">Chat</Nav.Link>
              <Nav.Link as={Link} to="/files" className="px-3">Files</Nav.Link>
              <Nav.Link as={Link} to="/ocr" className="px-3">OCR Results</Nav.Link>
            </Nav>
          </Navbar.Collapse>
        </Container>
      </Navbar>
      <Container className="my-5">
        <div className="bg-white rounded-lg shadow-md p-4 p-md-5">
          <div className="mb-4 mb-md-5 pb-4 border-bottom">
            <h1 className="display-5 fw-bold text-gray-900 mb-2">gRPC Go Sample Application</h1>
            <p className="text-muted lead">A comprehensive sample demonstrating gRPC communication patterns</p>
          </div>
          <Routes>
          <Route path="/" element={
            <div className="py-3">
              <p className="text-muted fs-5">This sample demonstrates multiple features built on gRPC. Use the menu above to navigate.</p>
            </div>
          } />
          <Route path="/greet" element={
            <div className="py-3">
              <p className="text-muted mb-3">Send your name and the gRPC server will respond with a greeting.</p>
              <Greeting />
            </div>
          } />
          <Route path="/counter" element={
            <div className="py-3">
              <p className="text-muted mb-3">Receive counter values as a server stream and display them in real time.</p>
              <CounterStream />
            </div>
          } />
          <Route path="/chat" element={
            <div className="py-3">
              <p className="text-muted mb-3">A simple bidirectional streaming chat. Messages are echoed via the server.</p>
              <Chat />
            </div>
          } />
          <Route path="/files" element={
            <div className="py-3">
              <h2 className="h3 fw-semibold mb-2">File Operations</h2>
              <p className="text-muted mb-4">Select a file to upload/download. You can switch the destination using the Storage Provider.</p>
              <div className="mb-4">
                <label className="form-label">
                  Storage Provider:
                </label>
                <select 
                  value={storageProvider} 
                  onChange={(e) => setStorageProvider(e.target.value)}
                  className="form-select w-auto"
                >
                  <option value="s3">AWS S3 (Localstack)</option>
                  <option value="gcs">Google Cloud Storage (fake-gcs)</option>
                  <option value="azure">Azure Blob Storage (Azurite)</option>
                </select>
              </div>
              <div className="d-flex flex-column gap-4">
                <FileUpload storageProvider={storageProvider} />
                <FileDownload storageProvider={storageProvider} />
              </div>
            </div>
          } />
          <Route path="/ocr" element={
            <div className="py-3">
              <h2 className="h3 fw-semibold mb-2">OCR Results</h2>
              <p className="text-muted mb-4">View OCR processing results for document files.</p>
              <div className="mb-4">
                <label className="form-label">
                  Storage Provider:
                </label>
                <select 
                  value={storageProvider} 
                  onChange={(e) => setStorageProvider(e.target.value)}
                  className="form-select w-auto"
                >
                  <option value="s3">AWS S3 (Localstack)</option>
                  <option value="gcs">Google Cloud Storage (fake-gcs)</option>
                  <option value="azure">Azure Blob Storage (Azurite)</option>
                </select>
              </div>
              <OCRResults storageProvider={storageProvider} />
            </div>
          } />
          <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
      </Container>
    </div>
  );
}

export default App;
