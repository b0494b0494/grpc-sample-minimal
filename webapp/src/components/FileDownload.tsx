import React from 'react';
import { useFileDownload } from '../hooks';

interface FileDownloadProps {
  storageProvider: string;
}

export const FileDownload: React.FC<FileDownloadProps> = ({ storageProvider }) => {
  const { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload } = useFileDownload(storageProvider);

  return (
    <section>
      <h3>File Download</h3>
      <form onSubmit={handleFileDownload}>
        <input
          type="text"
          value={downloadFilename}
          onChange={(e) => setDownloadFilename(e.target.value)}
          placeholder="Enter filename to download"
        />
        <button type="submit">Download File</button>
        {downloadStatus && <p>{downloadStatus}</p>}
      </form>
    </section>
  );
};
