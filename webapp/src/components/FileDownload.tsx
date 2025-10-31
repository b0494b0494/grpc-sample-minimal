import React, { useState } from 'react';
import { Button, Badge, Table, Spinner } from 'react-bootstrap';
import { useFileDownload } from '../hooks/useFileDownload';
import { useFileList } from '../hooks/useFileList';
import { deleteFileService } from '../services/grpcService';
import { AlertDialog } from './AlertDialog';

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
  
  // Dialog states
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [fileToDelete, setFileToDelete] = useState<string | null>(null);
  const [showStatusDialog, setShowStatusDialog] = useState(false);
  const [statusDialogTitle, setStatusDialogTitle] = useState('');
  const [statusDialogMessage, setStatusDialogMessage] = useState('');
  const [statusDialogVariant, setStatusDialogVariant] = useState<'success' | 'danger' | 'warning' | 'info'>('info');

  const handleDownload = async (filename: string) => {
    setDownloadFilename(filename);
    const e = { preventDefault: () => {} } as React.FormEvent;
    await handleFileDownload(e);
  };

  const handleDeleteClick = (filename: string) => {
    setFileToDelete(filename);
    setShowDeleteConfirm(true);
  };

  const handleDeleteConfirm = async () => {
    if (!fileToDelete) return;

    setDeleting(fileToDelete);
    setShowDeleteConfirm(false);

    try {
      const response = await deleteFileService(fileToDelete, storageProvider);
      if (response.success) {
        setStatusDialogTitle('Delete Success');
        setStatusDialogMessage(`File "${fileToDelete}" deleted successfully`);
        setStatusDialogVariant('success');
        // Refresh the file list
        await refreshFiles();
      } else {
        setStatusDialogTitle('Delete Failed');
        setStatusDialogMessage(`Failed to delete file: ${response.message}`);
        setStatusDialogVariant('danger');
      }
      setShowStatusDialog(true);
    } catch (err) {
      setStatusDialogTitle('Delete Error');
      setStatusDialogMessage(`Error deleting file: ${err instanceof Error ? err.message : 'Unknown error'}`);
      setStatusDialogVariant('danger');
      setShowStatusDialog(true);
    } finally {
      setDeleting(null);
      setFileToDelete(null);
    }
  };

  // Update status dialog when download status changes
  React.useEffect(() => {
    if (downloadStatus) {
      if (downloadStatus.includes('successful')) {
        setStatusDialogTitle('Download Success');
        setStatusDialogVariant('success');
      } else if (downloadStatus.includes('failed') || downloadStatus.includes('Error')) {
        setStatusDialogTitle('Download Failed');
        setStatusDialogVariant('danger');
      } else {
        return; // Don't show dialog for "Downloading..." status
      }
      setStatusDialogMessage(downloadStatus);
      setShowStatusDialog(true);
    }
  }, [downloadStatus]);

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
        <div className="alert alert-danger mb-3" role="alert">
          {error}
        </div>
      )}

      {!loading && (!files || files.length === 0) && !error && (
        <div className="alert alert-info mb-3" role="alert">
          No files found. Upload some files first.
        </div>
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
                        onClick={() => handleDeleteClick(file.filename)}
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

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        show={showDeleteConfirm}
        title="Confirm Delete"
        message={`Are you sure you want to delete "${fileToDelete}"?`}
        variant="danger"
        onClose={() => {
          setShowDeleteConfirm(false);
          setFileToDelete(null);
        }}
        onConfirm={handleDeleteConfirm}
        confirmText="Delete"
        cancelText="Cancel"
        showCancel={true}
      />

      {/* Status Dialog */}
      <AlertDialog
        show={showStatusDialog}
        title={statusDialogTitle}
        message={statusDialogMessage}
        variant={statusDialogVariant}
        onClose={() => setShowStatusDialog(false)}
      />
    </section>
  );
};
