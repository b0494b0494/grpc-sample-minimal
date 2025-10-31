// Utility functions for file operations

/**
 * Determine file type from filename extension
 */
export const getFileType = (filename: string): 'image' | 'pdf' | 'text' | 'unknown' => {
  const ext = filename.toLowerCase().split('.').pop() || '';
  
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg', 'ico'];
  const textExts = ['txt', 'md', 'csv', 'json', 'xml', 'html', 'css', 'js', 'ts', 'jsx', 'tsx', 'yaml', 'yml'];
  
  if (imageExts.includes(ext)) return 'image';
  if (ext === 'pdf') return 'pdf';
  if (textExts.includes(ext)) return 'text';
  return 'unknown';
};

/**
 * Check if file can be previewed
 */
export const canPreview = (filename: string): boolean => {
  return getFileType(filename) !== 'unknown';
};
