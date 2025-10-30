import React, { useState } from 'react';
import { downloadFileService } from '../services/grpcService';

export const useFileDownload = (storageProvider: string) => {
  const [downloadFilename, setDownloadFilename] = useState<string>('');
  const [downloadStatus, setDownloadStatus] = useState<string>('');

  const handleFileDownload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!downloadFilename) {
      setDownloadStatus('Please enter a filename.');
      return;
    }

    setDownloadStatus('Downloading...');
    try {
      const response = await downloadFileService(downloadFilename, storageProvider);
      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = downloadFilename;
        document.body.appendChild(a);
        a.click();
        a.remove();
        window.URL.revokeObjectURL(url);
        setDownloadStatus('Download successful.');
      } else {
        setDownloadStatus(`Download failed: ${response.statusText}`);
      }
    } catch (error: any) {
      setDownloadStatus(`Network Error: ${error.message}`);
    }
  };

  return { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload };
};
