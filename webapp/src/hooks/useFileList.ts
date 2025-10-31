import { useState, useEffect } from 'react';
import { listFilesService } from '../services/grpcService';
import { FileInfo } from '../types';

export const useFileList = (storageProvider: string) => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');

  const refreshFiles = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await listFilesService(storageProvider);
      // Handle null or undefined files array
      if (response.files === null || response.files === undefined) {
        setFiles([]);
      } else {
        setFiles(response.files);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load files');
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refreshFiles();
  }, [storageProvider]);

  return { files, loading, error, refreshFiles };
};
