import React from 'react';
import { useFileUpload } from '../hooks';

interface FileUploadProps {
  storageProvider: string;
}

export const FileUpload: React.FC<FileUploadProps> = ({ storageProvider }) => {
  const { selectedFile, uploadStatus, handleFileChange, handleFileUpload } = useFileUpload(storageProvider);

  return (
    <section>
      <h3>File Upload</h3>
      <form onSubmit={handleFileUpload}>
        <input type="file" onChange={handleFileChange} />
        <button type="submit">Upload File</button>
        {uploadStatus && <p>{uploadStatus}</p>}
      </form>
    </section>
  );
};
