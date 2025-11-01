import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import './index.css';
import App from './App';
import { UploadQueueProvider } from './contexts/UploadQueueContext';

// Import test utility for browser console testing
import './utils/testWorker';

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');
const root = ReactDOM.createRoot(rootElement);
root.render(
  <React.StrictMode>
    <BrowserRouter>
      <UploadQueueProvider>
        <App />
      </UploadQueueProvider>
    </BrowserRouter>
  </React.StrictMode>
);
