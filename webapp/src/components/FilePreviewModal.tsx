import React, { useState, useEffect } from 'react';
import { Modal, Spinner, Alert } from 'react-bootstrap';
import { downloadFileAsBlob } from '../services/grpcService';
import { getFileType } from '../utils/fileUtils';

interface FilePreviewModalProps {
  show: boolean;
  filename: string;
  storageProvider: string;
  onHide: () => void;
}

const TextPreview: React.FC<{ url: string }> = ({ url }) => {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    fetch(url)
      .then(res => {
        if (!res.ok) {
          throw new Error(`Failed to fetch text: ${res.statusText}`);
        }
        return res.text();
      })
      .then(text => {
        setContent(text);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [url]);

  if (loading) {
    return (
      <div className="text-center py-3">
        <Spinner animation="border" size="sm" />
        <span className="ms-2">Loading text content...</span>
      </div>
    );
  }

  if (error) {
    return <Alert variant="danger">Error loading text: {error}</Alert>;
  }

  return (
    <pre className="bg-light p-3 rounded" style={{ maxHeight: '500px', overflow: 'auto', margin: 0 }}>
      <code style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{content}</code>
    </pre>
  );
};

export const FilePreviewModal: React.FC<FilePreviewModalProps> = ({
  show,
  filename,
  storageProvider,
  onHide,
}) => {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileType = getFileType(filename);

  useEffect(() => {
    if (show && filename) {
      loadPreview();
    } else {
      // Cleanup blob URL when modal closes
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
        setPreviewUrl(null);
      }
    }

    return () => {
      // Cleanup blob URL on unmount
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [show, filename]);

  const loadPreview = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const blob = await downloadFileAsBlob(filename, storageProvider);
      const url = URL.createObjectURL(blob);
      setPreviewUrl(url);
    } catch (err: any) {
      setError(err.message || 'Failed to load preview');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal show={show} onHide={onHide} size="lg" centered>
      <Modal.Header closeButton>
        <Modal.Title>Preview: {filename}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {loading && (
          <div className="text-center py-4">
            <Spinner animation="border" role="status">
              <span className="visually-hidden">Loading preview...</span>
            </Spinner>
            <p className="mt-2 text-muted">Loading file preview...</p>
          </div>
        )}
        
        {error && (
          <Alert variant="danger">
            <strong>Error:</strong> {error}
          </Alert>
        )}

        {!loading && !error && previewUrl && (
          <>
            {fileType === 'image' && (
              <div className="text-center">
                <img 
                  src={previewUrl} 
                  alt={filename} 
                  className="img-fluid rounded"
                  style={{ maxHeight: '70vh' }}
                />
              </div>
            )}
            
            {fileType === 'pdf' && (
              <iframe 
                src={previewUrl} 
                width="100%" 
                height="600px"
                style={{ border: 'none' }}
                title={`PDF Preview: ${filename}`}
              />
            )}
            
            {fileType === 'text' && (
              <TextPreview url={previewUrl} />
            )}

            {fileType === 'unknown' && (
              <Alert variant="warning">
                Preview not available for this file type.
              </Alert>
            )}
          </>
        )}
      </Modal.Body>
    </Modal>
  );
};
