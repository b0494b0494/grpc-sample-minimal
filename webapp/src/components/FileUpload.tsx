import React from 'react';
import { Form, Button, Badge } from 'react-bootstrap';
import { useFileUpload } from '../hooks';

interface FileUploadProps {
  storageProvider: string;
}

export const FileUpload: React.FC<FileUploadProps> = ({ storageProvider }) => {
  const { selectedFile, uploadStatus, handleFileChange, handleFileUpload } = useFileUpload(storageProvider);

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex align-items-center gap-2 mb-3">
        <h3 className="h5 fw-semibold mb-0">File Upload</h3>
        <Badge bg="primary">{storageProvider.toUpperCase()}</Badge>
      </div>
      <Form onSubmit={handleFileUpload} className="d-flex flex-column gap-3">
        <Form.Group>
          <Form.Label>Select File:</Form.Label>
          <Form.Control type="file" onChange={handleFileChange} />
        </Form.Group>
        <Button variant="primary" type="submit">
          Upload File
        </Button>
        {uploadStatus && (
          <div className={`p-3 rounded-md ${
            uploadStatus.includes('successful') || uploadStatus.includes('success')
              ? 'bg-green-50 text-green-700 border border-green-200'
              : uploadStatus.includes('failed') || uploadStatus.includes('Error')
              ? 'bg-red-50 text-red-700 border border-red-200'
              : 'bg-blue-50 text-blue-700 border border-blue-200'
          }`}>
            <p className="text-sm font-medium">{uploadStatus}</p>
          </div>
        )}
      </Form>
    </section>
  );
};
