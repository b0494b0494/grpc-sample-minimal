package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	pb "grpc-sample-minimal/proto"
)

var (
	azureAccountName   = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	azureAccountKey    = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	azureContainerName = os.Getenv("AZURE_STORAGE_CONTAINER_NAME")
	azureEndpoint      = os.Getenv("AZURE_STORAGE_ENDPOINT") // For emulator, e.g., http://azurite:10000
)

type azureStorageService struct {
	blobClient     *azblob.Client
	containerName string
}

func NewAzureStorageService(ctx context.Context) (StorageService, error) {
	var blobClient *azblob.Client
	var err error

	containerName := azureContainerName
	if containerName == "" {
		containerName = "grpc-sample-container" // Default container name
	}

	// Use emulator if endpoint is set (for Azurite)
	if azureEndpoint != "" {
		// For Azurite emulator, use connection string format
		// Account name and key are typically: devstoreaccount1 / Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==
		accountName := azureAccountName
		if accountName == "" {
			accountName = "devstoreaccount1" // Default Azurite account name
		}
		accountKey := azureAccountKey
		if accountKey == "" {
			accountKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==" // Default Azurite account key
		}

		// Construct connection string for emulator
		// Format: DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://azurite:10000/devstoreaccount1;
		connStr := fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=%s;AccountKey=%s;BlobEndpoint=%s/%s;", accountName, accountKey, azureEndpoint, accountName)
		blobClient, err = azblob.NewClientFromConnectionString(connStr, nil)
	} else {
		// Use connection string from environment (for real Azure Storage)
		connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
		if connStr != "" {
			blobClient, err = azblob.NewClientFromConnectionString(connStr, nil)
		} else {
			return nil, fmt.Errorf("either AZURE_STORAGE_ENDPOINT (for emulator) or AZURE_STORAGE_CONNECTION_STRING must be set")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	// Ensure container exists
	serviceClient := blobClient.ServiceClient()
	containerClient := serviceClient.NewContainerClient(containerName)
	_, err = containerClient.Create(ctx, nil)
	if err != nil {
		// Check if container already exists (ignore that error)
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			// HTTP 409 Conflict means container already exists, which is fine
			if respErr.StatusCode == http.StatusConflict {
				log.Printf("Container %s already exists, continuing", containerName)
			} else {
				log.Printf("Warning: Failed to create Azure container %s: %v (status: %d)", containerName, err, respErr.StatusCode)
			}
		} else if strings.Contains(err.Error(), "ContainerAlreadyExists") || strings.Contains(err.Error(), "409") {
			log.Printf("Container %s already exists, continuing", containerName)
		} else {
			log.Printf("Warning: Failed to create Azure container %s: %v", containerName, err)
		}
	}

	return &azureStorageService{
		blobClient:     blobClient,
		containerName: containerName,
	}, nil
}

func (s *azureStorageService) UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error) {
	// Read content into bytes
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content for Azure upload: %w", err)
	}

	// Build storage path with namespace prefix (documents/, media/, or others/)
	storagePath := BuildStoragePath(filename)

	// Get block blob client
	serviceClient := s.blobClient.ServiceClient()
	containerClient := serviceClient.NewContainerClient(s.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(storagePath)

	// Upload the buffer
	_, err = blockBlobClient.UploadBuffer(ctx, contentBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to Azure Blob Storage: %w", err)
	}

	return &pb.FileUploadStatus{
		Filename:        filename,
		Success:         true,
		Message:         fmt.Sprintf("File %s uploaded to Azure Blob Storage at %s", filename, storagePath),
		StorageProvider: "azure",
	}, nil
}

func (s *azureStorageService) DownloadFile(ctx context.Context, filename string) (io.Reader, error) {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)

	// Get block blob client
	serviceClient := s.blobClient.ServiceClient()
	containerClient := serviceClient.NewContainerClient(s.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(storagePath)

	// Download the blob
	resp, err := blockBlobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from Azure Blob Storage: %w", err)
	}

	// Read all data from the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("failed to read Azure Blob object body: %w", err)
	}
	_ = resp.Body.Close()

	return bytes.NewReader(data), nil
}

func (s *azureStorageService) DeleteFile(ctx context.Context, filename string) error {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)

	// Get block blob client
	serviceClient := s.blobClient.ServiceClient()
	containerClient := serviceClient.NewContainerClient(s.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(storagePath)

	if _, err := blockBlobClient.Delete(ctx, nil); err != nil {
		return fmt.Errorf("failed to delete file from Azure Blob Storage: %w", err)
	}
	return nil
}

func (s *azureStorageService) ListFiles(ctx context.Context) ([]*pb.FileInfo, error) {
	var files []*pb.FileInfo
	
	serviceClient := s.blobClient.ServiceClient()
	containerClient := serviceClient.NewContainerClient(s.containerName)
	
	pager := containerClient.NewListBlobsFlatPager(nil)
	
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Azure blobs: %w", err)
		}
		
		for _, blob := range page.Segment.BlobItems {
			name := *blob.Name
			
			// Skip directory prefixes (names ending with "/")
			if strings.HasSuffix(name, "/") {
				continue
			}
			
			var namespace, filename string
			if strings.HasPrefix(name, "documents/") {
				namespace = "documents"
				filename = strings.TrimPrefix(name, "documents/")
			} else if strings.HasPrefix(name, "media/") {
				namespace = "media"
				filename = strings.TrimPrefix(name, "media/")
			} else if strings.HasPrefix(name, "others/") {
				namespace = "others"
				filename = strings.TrimPrefix(name, "others/")
			} else {
				// Legacy files without namespace
				namespace = "others"
				filename = name
			}
			
			size := int64(0)
			if blob.Properties != nil && blob.Properties.ContentLength != nil {
				size = *blob.Properties.ContentLength
			}
			
			files = append(files, &pb.FileInfo{
				Filename:  filename,
				Namespace: namespace,
				Size:      size,
			})
		}
	}
	
	return files, nil
}