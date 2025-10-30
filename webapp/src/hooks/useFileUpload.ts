import React, { useState } from 'react';
import { uploadFileService } from '../services/grpcService';
import { FileUploadStatus } from '../types';

export const useFileUpload = (storageProvider: string) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadStatus, setUploadStatus] = useState<string>('');

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleFileUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) {
      setUploadStatus('Please select a file first.');
      return;
    }

    setUploadStatus('Uploading...');
    try {
      const data: FileUploadStatus = await uploadFileService(selectedFile, storageProvider);
      if (data.success) {
        setUploadStatus(`Upload successful: ${data.message} (${data.bytesWritten} bytes) to ${data.storageProvider}`);
      } else {
        setUploadStatus(`Upload failed: ${data.message || 'Unknown error'}`);
      }
    } catch (error: any) {
      setUploadStatus(`Network Error: ${error.message}`);
    }
  };

  return { selectedFile, uploadStatus, handleFileChange, handleFileUpload };
};
