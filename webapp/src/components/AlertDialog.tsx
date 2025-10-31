import React from 'react';
import { Modal, Button } from 'react-bootstrap';

interface AlertDialogProps {
  show: boolean;
  title: string;
  message?: string;
  variant?: 'success' | 'danger' | 'warning' | 'info';
  onClose: () => void;
  onConfirm?: () => void;
  confirmText?: string;
  cancelText?: string;
  showCancel?: boolean;
  children?: React.ReactNode; // Add children support
}

export const AlertDialog: React.FC<AlertDialogProps> = ({
  show,
  title,
  message,
  variant = 'info',
  onClose,
  onConfirm,
  confirmText = 'OK',
  cancelText = 'Cancel',
  showCancel = false,
  children,
}) => {
  const handleConfirm = () => {
    if (onConfirm) {
      onConfirm();
    }
    onClose();
  };

  const getVariantClass = () => {
    switch (variant) {
      case 'success':
        return 'text-success';
      case 'danger':
        return 'text-danger';
      case 'warning':
        return 'text-warning';
      case 'info':
      default:
        return 'text-info';
    }
  };

  return (
    <Modal show={show} onHide={onClose} centered>
      <Modal.Header closeButton>
        <Modal.Title className={getVariantClass()}>{title}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {children || message}
      </Modal.Body>
      <Modal.Footer>
        {showCancel && (
          <Button variant="secondary" onClick={onClose}>
            {cancelText}
          </Button>
        )}
        <Button variant={variant === 'danger' ? 'danger' : variant === 'success' ? 'success' : 'primary'} onClick={handleConfirm}>
          {confirmText}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};
