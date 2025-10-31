import React from 'react';
import { useFileDownload } from '../hooks';

interface FileDownloadProps {
  storageProvider: string;
}

export const FileDownload: React.FC<FileDownloadProps> = ({ storageProvider }) => {
  const { downloadFilename, setDownloadFilename, downloadStatus, handleFileDownload } = useFileDownload(storageProvider);

  return (
    <section className="bg-gray-50 rounded-lg p-6 border border-gray-200">
      <h3 className="text-xl font-semibold text-gray-900 mb-4">File Download</h3>
      <form onSubmit={handleFileDownload} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Filename:
          </label>
          <input
            type="text"
            value={downloadFilename}
            onChange={(e) => setDownloadFilename(e.target.value)}
            placeholder="Enter filename to download"
            className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <button 
          type="submit"
          className="px-6 py-2 bg-primary-600 text-white font-medium rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Download File
        </button>
        {downloadStatus && (
          <div className={`p-3 rounded-md ${
            downloadStatus.includes('successful') || downloadStatus.includes('success')
              ? 'bg-green-50 text-green-700 border border-green-200'
              : downloadStatus.includes('failed') || downloadStatus.includes('Error')
              ? 'bg-red-50 text-red-700 border border-red-200'
              : 'bg-blue-50 text-blue-700 border border-blue-200'
          }`}>
            <p className="text-sm font-medium">{downloadStatus}</p>
          </div>
        )}
      </form>
    </section>
  );
};
