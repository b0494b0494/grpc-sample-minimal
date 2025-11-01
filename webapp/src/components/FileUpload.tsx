import React, { useState, useRef } from 'react';
import { Form, Button, Badge, ListGroup } from 'react-bootstrap';
import { useFileUpload } from '../hooks';
import { AlertDialog } from './AlertDialog';
import { UploadQueueView } from './UploadQueueView';

interface FileUploadProps {
  storageProvider: string;
}

export const FileUpload: React.FC<FileUploadProps> = ({ storageProvider }) => {
  const { 
    selectedFiles, 
    uploadStatus, 
    uploadTasks,
    isUploading,
    handleFileChange, 
    handleFileUpload,
    clearSelectedFiles,
    removeFile,
  } = useFileUpload(storageProvider);
  const [showDialog, setShowDialog] = useState(false);
  const [dialogTitle, setDialogTitle] = useState('');
  const [dialogMessage, setDialogMessage] = useState('');
  const [dialogVariant, setDialogVariant] = useState<'success' | 'danger' | 'warning' | 'info'>('info');
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Update dialog when upload status changes
  React.useEffect(() => {
    if (uploadStatus) {
      if (uploadStatus.includes('successful') || uploadStatus.includes('success')) {
        setDialogTitle('Upload Success');
        setDialogVariant('success');
      } else if (uploadStatus.includes('failed') || uploadStatus.includes('Error')) {
        setDialogTitle('Upload Failed');
        setDialogVariant('danger');
      } else if (uploadStatus.includes('Uploading')) {
        return; // Don't show dialog for "Uploading..." status
      } else {
        setDialogTitle('Upload Status');
        setDialogVariant('info');
      }
      setDialogMessage(uploadStatus);
      setShowDialog(true);
    }
  }, [uploadStatus]);

  // Drag and drop handlers
  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      const filesArray = Array.from(files);
      // Create a synthetic event for handleFileChange
      const syntheticEvent = {
        target: {
          files: filesArray,
        },
      } as unknown as React.ChangeEvent<HTMLInputElement>;
      handleFileChange(syntheticEvent);
    }
  };

  const handleDropAreaClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex align-items-center gap-2 mb-3">
        <h3 className="h5 fw-semibold mb-0">File Upload</h3>
        <Badge bg="primary">{storageProvider.toUpperCase()}</Badge>
      </div>
      <Form onSubmit={handleFileUpload} className="d-flex flex-column gap-3">
        <Form.Group>
          <Form.Label>Select File:</Form.Label>
          <input
            ref={fileInputRef}
            type="file"
            onChange={handleFileChange}
            className="d-none"
            accept="*/*"
            multiple
          />
          <div
            onClick={handleDropAreaClick}
            onDragEnter={handleDragEnter}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`border rounded p-5 text-center ${
              isDragging
                ? 'border-primary bg-primary bg-opacity-10'
                : 'border-secondary bg-white'
            }`}
            style={{
              cursor: 'pointer',
              transition: 'all 0.2s ease',
              borderStyle: isDragging ? 'dashed' : 'solid',
              borderWidth: '2px',
            }}
          >
            {selectedFiles.length > 0 ? (
              <div>
                <p className="mb-2">
                  <strong>{selectedFiles.length} file(s) selected</strong>
                </p>
                <p className="text-muted small mb-0">
                  Click to add more files or drag and drop files here
                </p>
              </div>
            ) : isDragging ? (
              <div>
                <p className="mb-0 text-primary">
                  <strong>Drop file here</strong>
                </p>
              </div>
            ) : (
              <div>
                <p className="mb-2">
                  <strong>Click to select</strong> or <strong>drag and drop</strong> a file here
                </p>
                <p className="text-muted small mb-0">
                  Supports any file type
                </p>
              </div>
            )}
          </div>
        </Form.Group>
        <div className="d-flex gap-2">
          <Button 
            variant="primary" 
            type="submit"
            disabled={selectedFiles.length === 0 || isUploading}
          >
            {isUploading ? 'Uploading...' : `Upload ${selectedFiles.length} File(s)`}
          </Button>
          {selectedFiles.length > 0 && !isUploading && (
            <Button 
              variant="outline-secondary" 
              type="button"
              onClick={clearSelectedFiles}
            >
              Clear
            </Button>
          )}
        </div>
      </Form>

      {/* Selected Files List */}
      {selectedFiles.length > 0 && !isUploading && (
        <div className="mt-3">
          <h6 className="mb-2">Selected Files:</h6>
          <ListGroup>
            {selectedFiles.map((file, index) => (
              <ListGroup.Item key={index} className="d-flex justify-content-between align-items-center">
                <div className="flex-grow-1">
                  <div className="fw-semibold">{file.name}</div>
                  <small className="text-muted">
                    {(file.size / 1024 / 1024).toFixed(2)} MB
                  </small>
                </div>
                <Button
                  variant="outline-danger"
                  size="sm"
                  onClick={() => removeFile(index)}
                >
                  Remove
                </Button>
              </ListGroup.Item>
            ))}
          </ListGroup>
        </div>
      )}

      {/* Upload Queue View */}
      {uploadTasks.length > 0 && (
        <UploadQueueView showCompleted={false} maxHeight="400px" />
      )}

      <AlertDialog
        show={showDialog}
        title={dialogTitle}
        message={dialogMessage}
        variant={dialogVariant}
        onClose={() => setShowDialog(false)}
      />
    </section>
  );
};
