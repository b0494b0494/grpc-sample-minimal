import React, { useState } from 'react';
import { Button, Badge, Table, Spinner, Alert } from 'react-bootstrap';
import { useFileDownload } from '../hooks/useFileDownload';
import { useFileList } from '../hooks/useFileList';
import { deleteFileService } from '../services/grpcService';

interface FileDownloadProps {
  storageProvider: string;
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const getNamespaceBadgeVariant = (namespace: string): string => {
  switch (namespace) {
    case 'documents':
      return 'info';
    case 'media':
      return 'success';
    default:
      return 'secondary';
  }
};

export const FileDownload: React.FC<FileDownloadProps> = ({ storageProvider }) => {
  const { files, loading, error, refreshFiles } = useFileList(storageProvider);
  const { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload } = useFileDownload(storageProvider);
  const [deleteStatus, setDeleteStatus] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);

  const handleDownload = async (filename: string) => {
    setDownloadFilename(filename);
    const e = { preventDefault: () => {} } as React.FormEvent;
    await handleFileDownload(e);
  };

  const handleDelete = async (filename: string) => {
    if (!window.confirm(`Are you sure you want to delete "${filename}"?`)) {
      return;
    }

    setDeleting(filename);
    setDeleteStatus(null);

    try {
      const response = await deleteFileService(filename, storageProvider);
      if (response.success) {
        setDeleteStatus(`File "${filename}" deleted successfully`);
        // Refresh the file list
        await refreshFiles();
      } else {
        setDeleteStatus(`Failed to delete file: ${response.message}`);
      }
    } catch (err) {
      setDeleteStatus(`Error deleting file: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setDeleting(null);
      setTimeout(() => setDeleteStatus(null), 5000);
    }
  };

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <div className="d-flex align-items-center gap-2">
          <h3 className="h5 fw-semibold mb-0">File List & Download</h3>
          <Badge bg="primary">{storageProvider.toUpperCase()}</Badge>
        </div>
        <Button variant="outline-secondary" size="sm" onClick={refreshFiles}>
          Refresh
        </Button>
      </div>

      {loading && (
        <div className="text-center py-4">
          <Spinner animation="border" role="status">
            <span className="visually-hidden">Loading...</span>
          </Spinner>
        </div>
      )}

      {error && (
        <Alert variant="danger" className="mb-3">
          {error}
        </Alert>
      )}

      {downloadStatus && (
        <Alert variant={downloadStatus.includes('successful') ? 'success' : 'warning'} className="mb-3">
          {downloadStatus}
        </Alert>
      )}

      {deleteStatus && (
        <Alert variant={deleteStatus.includes('successful') || deleteStatus.includes('deleted successfully') ? 'success' : 'danger'} className="mb-3">
          {deleteStatus}
        </Alert>
      )}

      {!loading && (!files || files.length === 0) && !error && (
        <Alert variant="info" className="mb-3">
          No files found. Upload some files first.
        </Alert>
      )}

      {!loading && files.length > 0 && (
        <div className="table-responsive">
          <Table striped bordered hover size="sm">
            <thead>
              <tr>
                <th>Namespace</th>
                <th>Filename</th>
                <th>Size</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {files.map((file, index) => (
                <tr key={index}>
                  <td>
                    <Badge bg={getNamespaceBadgeVariant(file.namespace)}>
                      {file.namespace}
                    </Badge>
                  </td>
                  <td className="font-monospace small">{file.filename}</td>
                  <td>{formatFileSize(file.size)}</td>
                  <td>
                    <div className="d-flex gap-2">
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => handleDownload(file.filename)}
                      >
                        Download
                      </Button>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => handleDelete(file.filename)}
                        disabled={deleting === file.filename}
                      >
                        {deleting === file.filename ? 'Deleting...' : 'Delete'}
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      )}
    </section>
  );
};
