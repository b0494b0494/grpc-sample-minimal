import { useState, useEffect } from 'react';
import { listOCRResultsService, getOCRResultService } from '../services/grpcService';
import { OCRResultSummary, OCRResultResponse } from '../types';

export const useOCRResults = (storageProvider: string) => {
  const [results, setResults] = useState<OCRResultSummary[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');

  const refreshResults = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await listOCRResultsService(storageProvider);
      setResults(response.results || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load OCR results');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refreshResults();
  }, [storageProvider]);

  return { results, loading, error, refreshResults };
};

export const useOCRResult = (filename: string, storageProvider: string, engineName: string = 'tesseract') => {
  const [result, setResult] = useState<OCRResultResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');

  const fetchResult = async () => {
    if (!filename || !storageProvider) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      const data = await getOCRResultService(filename, storageProvider, engineName);
      setResult(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load OCR result');
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResult();
  }, [filename, storageProvider, engineName]);

  return { result, loading, error, refreshResult: fetchResult };
};
