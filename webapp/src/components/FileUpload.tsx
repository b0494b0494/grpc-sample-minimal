import React, { useState } from 'react';
import { Form, Button, Badge } from 'react-bootstrap';
import { useFileUpload } from '../hooks';
import { AlertDialog } from './AlertDialog';

interface FileUploadProps {
  storageProvider: string;
}

export const FileUpload: React.FC<FileUploadProps> = ({ storageProvider }) => {
  const { selectedFile, uploadStatus, handleFileChange, handleFileUpload } = useFileUpload(storageProvider);
  const [showDialog, setShowDialog] = useState(false);
  const [dialogTitle, setDialogTitle] = useState('');
  const [dialogMessage, setDialogMessage] = useState('');
  const [dialogVariant, setDialogVariant] = useState<'success' | 'danger' | 'warning' | 'info'>('info');

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

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex align-items-center gap-2 mb-3">
        <h3 className="h5 fw-semibold mb-0">File Upload</h3>
        <Badge bg="primary">{storageProvider.toUpperCase()}</Badge>
      </div>
      <Form onSubmit={handleFileUpload} className="d-flex flex-column gap-3">
        <Form.Group>
          <Form.Label>Select File:</Form.Label>
          <Form.Control 
            type="file" 
            onChange={handleFileChange}
            required
            accept="*/*"
          />
        </Form.Group>
        <Button 
          variant="primary" 
          type="submit"
          disabled={!selectedFile}
        >
          Upload File
        </Button>
      </Form>

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
