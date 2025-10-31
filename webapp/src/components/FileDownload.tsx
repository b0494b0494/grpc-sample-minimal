import React from 'react';
import { Form, Button, Badge } from 'react-bootstrap';
import { useFileDownload } from '../hooks';

interface FileDownloadProps {
  storageProvider: string;
}

export const FileDownload: React.FC<FileDownloadProps> = ({ storageProvider }) => {
  const { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload } = useFileDownload(storageProvider);

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex align-items-center gap-2 mb-3">
        <h3 className="h5 fw-semibold mb-0">File Download</h3>
        <Badge bg="primary">{storageProvider.toUpperCase()}</Badge>
      </div>
      <Form onSubmit={handleFileDownload} className="d-flex flex-column gap-3">
        <Form.Group>
          <Form.Label>Filename:</Form.Label>
          <Form.Control
            type="text"
            value={downloadFilename}
            onChange={(e) => setDownloadFilename(e.target.value)}
            placeholder="Enter filename to download"
          />
        </Form.Group>
        <Button variant="primary" type="submit">
          Download File
        </Button>
        {downloadStatus && (
          <div className={`p-3 rounded-md ${
            downloadStatus.includes('successful') || downloadStatus.includes('success')
              ? 'bg-green-50 text-green-700 border border-green-200'
              : downloadStatus.includes('failed') || downloadStatus.includes('Error')
              ? 'bg-red-50 text-red-700 border border-red-200'
              : 'bg-blue-50 text-blue-700 border border-blue-200'
          }`}>
            <p className="text-sm font-medium">{downloadStatus}</p>
          </div>
        )}
      </Form>
    </section>
  );
};
