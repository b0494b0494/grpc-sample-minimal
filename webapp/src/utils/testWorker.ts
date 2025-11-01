// Test utility to verify Worker functionality
// This can be called from browser console for testing

export async function testWorkerUpload() {
  const testFile = new File(['test content'], 'test.txt', { type: 'text/plain' });
  
  console.log('=== Worker Upload Test ===');
  console.log('Test file:', testFile.name, testFile.size, 'bytes');
  console.log('Window location:', window.location.origin);
  
  const { getUploadWorkerManager } = await import('./uploadWorkerManager');
  const workerManager = getUploadWorkerManager();
  
  console.log('Worker available:', workerManager.isAvailable());
  
  if (!workerManager.isAvailable()) {
    await workerManager.initialize();
    console.log('Worker initialized, available:', workerManager.isAvailable());
  }
  
  const testTaskId = `test-${Date.now()}`;
  const apiBaseUrl = window.location.origin;
  const authToken = 'my-secret-token';
  
  console.log('Uploading with:');
  console.log('  apiBaseUrl:', apiBaseUrl);
  console.log('  taskId:', testTaskId);
  
  try {
    // Set up message handlers for testing
    workerManager.onMessage('PROGRESS_UPDATE', (msg) => {
      console.log('[Worker] Progress:', msg.payload);
    });
    
    workerManager.onMessage('TASK_COMPLETE', (msg) => {
      console.log('[Worker] Complete:', msg.payload);
    });
    
    workerManager.onMessage('TASK_ERROR', (msg) => {
      console.error('[Worker] Error:', msg.payload);
    });
    
    await workerManager.uploadTask(testTaskId, testFile, 's3', apiBaseUrl, authToken);
    console.log('Upload task sent successfully');
  } catch (error) {
    console.error('Upload task failed:', error);
  }
}

// Make it available globally for browser console testing
if (typeof window !== 'undefined') {
  (window as any).testWorkerUpload = testWorkerUpload;
}
