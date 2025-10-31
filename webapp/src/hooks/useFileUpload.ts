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

    // Validate file name
    if (!selectedFile.name || selectedFile.name.trim() === '') {
      setUploadStatus('Please select a file with a valid filename.');
      return;
    }

    setUploadStatus('Uploading...');
    try {
      const data: FileUploadStatus = await uploadFileService(selectedFile, storageProvider);
      
      console.log('Upload result:', data);
      
      // Determine success: check success field first, then message content
      const isSuccess = data.success === true || 
                       (data.success !== false && 
                        data.message && 
                        (data.message.toLowerCase().includes('uploaded') || 
                         data.message.toLowerCase().includes('success')));
      
      if (isSuccess) {
        setUploadStatus(`Upload successful: ${data.message || 'File uploaded successfully'}`);
        // Clear the selected file after successful upload
        setSelectedFile(null);
      } else {
        const failureMessage = data.message || data.error || 'Upload failed';
        setUploadStatus(`Upload failed: ${failureMessage}`);
        console.error('Upload failed:', data);
      }
    } catch (error: any) {
      console.error('Upload error details:', error);
      const errorMsg = error?.message || error?.toString() || 'Unknown error';
      setUploadStatus(`Upload failed: ${errorMsg}`);
    }
  };

  return { selectedFile, uploadStatus, handleFileChange, handleFileUpload };
};
