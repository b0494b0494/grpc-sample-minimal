import React from 'react';
import { ListGroup, Badge, ProgressBar, Button, Card } from 'react-bootstrap';
import { useUploadQueue } from '../hooks/useUploadQueue';
import { UploadTask } from '../types';

interface UploadQueueViewProps {
  showCompleted?: boolean; // Show completed tasks
  maxHeight?: string; // Max height for scrollable list
}

export const UploadQueueView: React.FC<UploadQueueViewProps> = ({
  showCompleted = false,
  maxHeight = '400px',
}) => {
  const {
    tasks,
    activeTaskId,
    updateTask,
    removeTask,
    cancelTask,
    getQueueStats,
  } = useUploadQueue();

  const stats = getQueueStats();

  // Filter tasks based on showCompleted
  const displayTasks = showCompleted
    ? tasks
    : tasks.filter(task => task.status !== 'success' && task.status !== 'cancelled');

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
  };

  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  const getTaskDuration = (task: UploadTask): string | null => {
    if (task.completedAt && task.startedAt) {
      return formatDuration(task.completedAt - task.startedAt);
    }
    if (task.startedAt) {
      return `Running: ${formatDuration(Date.now() - task.startedAt)}`;
    }
    return null;
  };

  const getStatusBadgeVariant = (status: UploadTask['status']): string => {
    switch (status) {
      case 'success':
        return 'success';
      case 'error':
        return 'danger';
      case 'uploading':
        return 'primary';
      case 'pending':
        return 'secondary';
      case 'cancelled':
        return 'warning';
      default:
        return 'secondary';
    }
  };

  if (tasks.length === 0) {
    return (
      <Card className="mt-3">
        <Card.Body className="text-center text-muted">
          <p className="mb-0">No upload tasks</p>
        </Card.Body>
      </Card>
    );
  }

  return (
    <Card className="mt-3">
      <Card.Header className="d-flex justify-content-between align-items-center">
        <div>
          <strong>Upload Queue</strong>
          <Badge bg="info" className="ms-2">
            {stats.total} total
          </Badge>
          {stats.uploading > 0 && (
            <Badge bg="primary" className="ms-1">
              {stats.uploading} uploading
            </Badge>
          )}
          {stats.pending > 0 && (
            <Badge bg="secondary" className="ms-1">
              {stats.pending} pending
            </Badge>
          )}
        </div>
        <div>
          {stats.success > 0 && (
            <Badge bg="success" className="ms-1">
              {stats.success} success
            </Badge>
          )}
          {stats.error > 0 && (
            <Badge bg="danger" className="ms-1">
              {stats.error} failed
            </Badge>
          )}
        </div>
      </Card.Header>
      <Card.Body style={{ maxHeight, overflowY: 'auto' }}>
        <ListGroup variant="flush">
          {displayTasks.map((task) => (
            <ListGroup.Item key={task.id}>
              <div className="d-flex justify-content-between align-items-start mb-2">
                <div className="flex-grow-1">
                  <div className="d-flex align-items-center mb-1">
                    <span className="fw-semibold me-2">{task.file.name}</span>
                    {task.id === activeTaskId && (
                      <Badge bg="info">Active</Badge>
                    )}
                  </div>
                  <div className="d-flex align-items-center gap-2 mb-2">
                    <Badge bg={getStatusBadgeVariant(task.status)}>
                      {task.status}
                    </Badge>
                    <small className="text-muted">
                      {formatFileSize(task.file.size)}
                    </small>
                    {getTaskDuration(task) && (
                      <small className="text-muted">
                        ? {getTaskDuration(task)}
                      </small>
                    )}
                  </div>
                  {task.status === 'uploading' && (
                    <ProgressBar
                      now={task.progress}
                      label={`${task.progress}%`}
                      className="mb-2"
                    />
                  )}
                  {task.status === 'success' && task.result && (
                    <small className="text-success d-block">
                      ? {task.result.message}
                    </small>
                  )}
                  {task.status === 'error' && task.error && (
                    <small className="text-danger d-block">
                      ? {task.error}
                    </small>
                  )}
                  {task.status === 'cancelled' && (
                    <small className="text-warning d-block">
                      Cancelled
                    </small>
                  )}
                </div>
                <div className="d-flex gap-1 ms-2">
                  {(task.status === 'pending' || task.status === 'uploading') && (
                    <Button
                      variant="outline-danger"
                      size="sm"
                      onClick={() => cancelTask(task.id)}
                    >
                      Cancel
                    </Button>
                  )}
                  {(task.status === 'success' || task.status === 'error' || task.status === 'cancelled') && (
                    <Button
                      variant="outline-secondary"
                      size="sm"
                      onClick={() => removeTask(task.id)}
                    >
                      Remove
                    </Button>
                  )}
                </div>
              </div>
            </ListGroup.Item>
          ))}
        </ListGroup>
      </Card.Body>
      {!showCompleted && stats.completed > 0 && (
        <Card.Footer className="text-center">
          <small className="text-muted">
            {stats.completed} completed task(s) hidden.{' '}
            <Button
              variant="link"
              size="sm"
              className="p-0"
              onClick={() => {
                // This would need to be handled by parent component
                // or we could add a state management here
              }}
            >
              Show all
            </Button>
          </small>
        </Card.Footer>
      )}
    </Card>
  );
};
