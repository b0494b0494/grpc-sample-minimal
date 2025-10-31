import React, { useState } from 'react';
import { Table, Button, Badge, Card, Spinner, Accordion } from 'react-bootstrap';
import { useOCRResults, useOCRResult } from '../hooks';
import { AlertDialog } from './AlertDialog';
import { OCRResultSummary } from '../types';

interface OCRResultsProps {
  storageProvider: string;
}

export const OCRResults: React.FC<OCRResultsProps> = ({ storageProvider }) => {
  const { results, loading, error, refreshResults } = useOCRResults(storageProvider);
  const [selectedFilename, setSelectedFilename] = useState<string | null>(null);
  const [selectedEngine, setSelectedEngine] = useState<string>('tesseract');
  const { result: detailResult, loading: detailLoading, error: detailError } = useOCRResult(
    selectedFilename || '',
    storageProvider,
    selectedEngine
  );

  const formatDate = (timestamp: number): string => {
    if (!timestamp || timestamp === 0) return 'N/A';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
  };

  const getStatusBadgeVariant = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'completed':
        return 'success';
      case 'processing':
        return 'warning';
      case 'failed':
        return 'danger';
      default:
        return 'secondary';
    }
  };

  const handleViewResult = (filename: string, engineName: string) => {
    setSelectedFilename(filename);
    setSelectedEngine(engineName);
  };

  if (loading) {
    return (
      <Card>
        <Card.Body className="text-center py-5">
          <Spinner animation="border" role="status">
            <span className="visually-hidden">Loading OCR results...</span>
          </Spinner>
          <p className="mt-3 text-muted">Loading OCR results...</p>
        </Card.Body>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <Card.Body>
          <div className="alert alert-danger" role="alert">
            <strong>Error:</strong> {error}
          </div>
          <Button variant="primary" onClick={refreshResults}>
            Retry
          </Button>
        </Card.Body>
      </Card>
    );
  }

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="h4 fw-semibold mb-0">OCR Results</h2>
        <Button variant="outline-primary" size="sm" onClick={refreshResults}>
          Refresh
        </Button>
      </div>

      {results.length === 0 ? (
        <div className="text-center py-5 text-muted">
          <p>No OCR results found for {storageProvider}.</p>
          <p className="small">Upload document files to trigger OCR processing.</p>
        </div>
      ) : (
        <div className="table-responsive">
          <Table striped bordered hover size="sm">
            <thead>
              <tr>
                <th>Filename</th>
                <th>Engine</th>
                <th>Status</th>
                <th>Processed At</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {results.map((item: OCRResultSummary, index: number) => (
                <tr key={`${item.filename}-${item.engine_name}-${index}`}>
                  <td className="font-monospace small">{item.filename}</td>
                  <td>
                    <Badge bg="info">{item.engine_name}</Badge>
                  </td>
                  <td>
                    <Badge bg={getStatusBadgeVariant(item.status)}>
                      {item.status}
                    </Badge>
                  </td>
                  <td className="small">{formatDate(item.processed_at)}</td>
                  <td>
                    <Button
                      variant="outline-primary"
                      size="sm"
                      onClick={() => handleViewResult(item.filename, item.engine_name)}
                    >
                      {item.status === 'completed' ? 'View' : item.status === 'processing' ? 'Processing...' : 'View Details'}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      )}

      {/* OCR Result Detail Modal/Dialog */}
      {selectedFilename && (
        <AlertDialog
          show={!!selectedFilename}
          title={`OCR Result: ${selectedFilename}`}
          variant="info"
          onClose={() => setSelectedFilename(null)}
          showCancel={false}
          confirmText="Close"
        >
          {detailLoading ? (
            <div className="text-center py-3">
              <Spinner animation="border" size="sm" />
              <p className="mt-2 text-muted small">Loading details...</p>
            </div>
          ) : detailError ? (
            <div className="alert alert-danger">
              <strong>Error loading OCR result:</strong> {detailError}
              <br />
              <small className="text-muted">This may indicate that OCR processing failed or the result is not available.</small>
            </div>
          ) : detailResult ? (
            <div className="ocr-result-detail">
              <div className="mb-3">
                <strong>Engine:</strong> <Badge bg="info">{detailResult.engine_name}</Badge>{' '}
                <strong>Status:</strong>{' '}
                <Badge bg={getStatusBadgeVariant(detailResult.status)}>
                  {detailResult.status}
                </Badge>
              </div>
              {detailResult.confidence > 0 && (
                <div className="mb-3">
                  <strong>Confidence:</strong> {(detailResult.confidence * 100).toFixed(1)}%
                </div>
              )}
              {detailResult.error_message && (
                <div className="alert alert-danger mb-3">
                  <strong>Processing Error:</strong> {detailResult.error_message}
                </div>
              )}
              {detailResult.status === 'failed' && !detailResult.error_message && (
                <div className="alert alert-warning mb-3">
                  <strong>Status:</strong> OCR processing failed. No error details available.
                </div>
              )}
              {detailResult.status === 'processing' && (
                <div className="alert alert-info mb-3">
                  <Spinner animation="border" size="sm" className="me-2" />
                  <strong>Status:</strong> OCR processing is still in progress. Please refresh to check for updates.
                </div>
              )}
              {detailResult.pages && detailResult.pages.length > 1 ? (
                <Accordion>
                  {detailResult.pages.map((page, idx) => (
                    <Accordion.Item key={idx} eventKey={idx.toString()}>
                      <Accordion.Header>
                        Page {page.page_number}
                        {page.confidence > 0 && (
                          <span className="ms-2 small text-muted">
                            ({(page.confidence * 100).toFixed(1)}% confidence)
                          </span>
                        )}
                      </Accordion.Header>
                      <Accordion.Body>
                        <pre className="bg-light p-3 rounded small" style={{ whiteSpace: 'pre-wrap', maxHeight: '300px', overflowY: 'auto' }}>
                          {page.text || '(No text extracted)'}
                        </pre>
                      </Accordion.Body>
                    </Accordion.Item>
                  ))}
                </Accordion>
              ) : detailResult.status === 'completed' ? (
                <div>
                  <strong>Extracted Text:</strong>
                  <pre className="bg-light p-3 rounded mt-2 small" style={{ whiteSpace: 'pre-wrap', maxHeight: '400px', overflowY: 'auto' }}>
                    {detailResult.extracted_text || '(No text extracted)'}
                  </pre>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="alert alert-warning">
              <strong>No result available:</strong> OCR result not found for this file.
              <br />
              <small className="text-muted">The file may not have been processed yet, or processing may have failed.</small>
            </div>
          )}
        </AlertDialog>
      )}
    </section>
  );
};
