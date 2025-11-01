# gRPC Sample Application in Go

This is a minimal gRPC sample application implemented in Go, demonstrating basic client-server communication using Protocol Buffers and gRPC.

## Project Structure

- `proto/`: Contains the Protocol Buffer definition (`greeter.proto`) and the generated Go code with namespace `grpc-sample-minimal/proto/`.
- `server/`: Implements the gRPC server with a layered architecture (domain, application, infrastructure) including authentication and logging interceptors.
  - `server/domain/`: Domain layer with storage service implementations (S3, GCS, Azure Blob Storage), queue services (Pub/Sub, SQS, Azure Queue), OCR services (Tesseract, EasyOCR), document converters, MultiOCRClient for managing multiple OCR endpoints, and SQLite database repository
  - `server/application/`: Application layer service orchestrating gRPC calls, file operations, and OCR task queuing
- `server/ocr/`: Standalone OCR service that processes images and documents using multiple OCR engines (Tesseract, EasyOCR). It continuously dequeues OCR tasks from provider-specific queues and processes them asynchronously. Each OCR engine runs in a separate container.
- `client/`: Implements the gRPC client with authentication and logging interceptors.
- `webapp/`: Contains a React frontend application (TypeScript with React Bootstrap, API service layer, custom hooks, and component-based structure) and a Go backend that exposes HTTP API endpoints for gRPC calls.
  - `webapp/src/components/`: React components including `AlertDialog` for modal dialogs and `OCRResults` for displaying OCR processing results
  - `webapp/src/hooks/`: Custom React hooks for file operations and OCR results
  - `webapp/src/services/`: API service layer for gRPC calls and OCR operations
  - `webapp/handlers/`: HTTP handlers for file operations and OCR endpoints
- `Dockerfile.server`: Dockerfile for building the gRPC server image.
- `Dockerfile.client`: Dockerfile for building the gRPC client image.
- `Dockerfile.webapp`: Dockerfile for building the web application image (React frontend + Go backend).
- `Dockerfile.tesseract`: Dockerfile for building the Tesseract OCR service image.
- `Dockerfile.easyocr`: Dockerfile for building the EasyOCR service image (Python-based, uses PyTorch).
- `Dockerfile.ocr`: Dockerfile for building the OCR service image (legacy, for reference).
- `docker-compose.yml`: Defines and runs the multi-container Docker application with storage emulators (Localstack, fake-gcs, Azurite) and multiple OCR services (one per engine).

## How to Run

This application uses Docker Compose for easy setup and execution.

1.  **Ensure Docker is Running:** Make sure Docker Desktop or Docker Engine is running on your system.

2.  **Navigate to the Project Root:** Open your terminal and navigate to the root directory of this project:
    ```bash
    cd /path/to/grpc-sample-minimal
    ```

3.  **Configure Environment Variables:** Create a `.env` file in the project root with the following content:
    ```
    AUTH_TOKEN=my-secret-token
    AWS_ACCESS_KEY_ID=test
    AWS_SECRET_ACCESS_KEY=test
    AWS_REGION=us-east-1
    S3_BUCKET_NAME=grpc-sample-bucket
    LOCALSTACK_ENDPOINT=http://localstack:4566
    GRPC_SERVER_PORT=50051
    OCR_SERVICE_PORT=50052
    # Multiple OCR engine endpoints (for multi-engine support)
    OCR_TESSERACT_ENDPOINT=http://ocr-tesseract-service:50052
    OCR_EASYOCR_ENDPOINT=http://ocr-easyocr-service:50053
    # Legacy single endpoint (for backward compatibility)
    # OCR_SERVICE_ENDPOINT=ocr-service:50052
    DB_PATH=/app/data/files.db
    # GCS emulator
    STORAGE_EMULATOR_HOST=fake-gcs:4443
    GOOGLE_CLOUD_PROJECT=dev-project
    GCS_BUCKET_NAME=grpc-sample-bucket
    # Pub/Sub emulator
    PUBSUB_EMULATOR_HOST=pubsub-emulator:8085
    # Azure Storage emulator
    AZURE_STORAGE_ENDPOINT=http://azurite:10000
    AZURE_STORAGE_ACCOUNT_NAME=devstoreaccount1
    # Note: This is a dummy key for Azurite emulator (development only)
    # For production, use your actual Azure Storage Account key or connection string
    AZURE_STORAGE_ACCOUNT_KEY=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==
    AZURE_STORAGE_CONTAINER_NAME=grpc-sample-container
    ```
    *Note: The `AUTH_TOKEN` is used for gRPC authentication. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `S3_BUCKET_NAME`, and `LOCALSTACK_ENDPOINT` are for Localstack S3 integration. `GRPC_SERVER_PORT` defines the port the gRPC server listens on. `OCR_TESSERACT_ENDPOINT` and `OCR_EASYOCR_ENDPOINT` specify the gRPC endpoints for each OCR engine service (default ports: 50052 for Tesseract, 50053 for EasyOCR). `OCR_SERVICE_PORT` is used by individual OCR service containers to set their listening port. `DB_PATH` specifies the path to the SQLite database file for file metadata and OCR results storage.*
    
    **GCS and Pub/Sub Emulators:** Google Cloud Storage and Pub/Sub emulator configuration is handled automatically in `docker-compose.yml`. The `STORAGE_EMULATOR_HOST` and `PUBSUB_EMULATOR_HOST` are set to connect to the emulators running in Docker containers.
    
    **Azure Storage Emulator:** Azure Blob Storage configuration is handled automatically in `docker-compose.yml`. For production use, you would set:
    - `AZURE_STORAGE_CONNECTION_STRING`: Connection string for real Azure Storage (when not using emulator)
    
    **Database:** The SQLite database is automatically created on first run. The database file is persisted in a Docker volume (`server-data`) mounted at `/app/data` in the server container.

4.  **Build and Run with Docker Compose:** Execute the following command to build the Docker images and start all services:
    ```bash
    docker-compose up --build
    ```
    
    Or run in detached mode:
    ```bash
    docker-compose up --build -d
    ```

    You should see output from the `server`, `client`, OCR services, and storage emulator services in your terminal. The `webapp` will be accessible via your browser.

    **Services:**
    - **server:** Main gRPC server handling file operations and gRPC requests (port 50051). Uses MultiOCRClient to manage connections to multiple OCR engines.
    - **ocr-tesseract-service:** Standalone OCR service for processing images and documents with Tesseract OCR engine (port 50052)
    - **ocr-easyocr-service:** Standalone OCR service for processing images and documents with EasyOCR engine (Python-based, uses PyTorch) (port 50053)
    - **webapp:** Web application with React frontend and Go backend (port 8080)

    **Storage Emulators:**
    - **Localstack:** The `localstack` service emulates AWS S3 and SQS locally. The server will attempt to create an S3 bucket named `grpc-sample-bucket` and an SQS queue named `ocr-tasks-queue` on startup (port 4566).
    - **fake-gcs:** The `fake-gcs` service emulates Google Cloud Storage locally (port 4443).
    - **pubsub-emulator:** The `pubsub-emulator` service emulates Google Cloud Pub/Sub locally. It automatically creates topics and subscriptions for OCR task processing (port 8085).
    - **Azurite:** The `azurite` service emulates Azure Blob Storage and Queue Storage locally:
      - **Blob Storage**: Port 10000 (for file uploads/downloads)
      - **Queue Storage**: Port 10001 (for OCR task queuing)
      - The application automatically uses the correct port based on the operation type

    **Example Client Output (from `client` service):**
    ```
    server-1  | 2025/10/30 12:52:28 server listening at [::]:50051
    server-1  | 2025/10/30 12:52:29 Auth successful for method: /proto.Greeter/SayHello
    server-1  | 2025/10/30 12:52:29 Incoming RPC: /proto.Greeter/SayHello, Request: name:"Docker"
    server-1  | 2025/10/30 12:52:29 Received: Docker
    client-1  | 2025/10/30 12:52:29 Outgoing RPC: /proto.Greeter/SayHello, Request: name:"Docker"
    client-1  | 2025/10/30 12:52:29 Incoming RPC: /proto.Greeter/SayHello, Response: message:"Hello Docker", Error: <nil>
    client-1  | 2025/10/30 12:52:29 Greeting: Hello Docker
    client-1  | 2025/10/30 12:52:29 Calling StreamCounter with limit 5
    server-1  | 2025/10/30 13:01:40 Auth successful for stream method: /proto.Greeter/StreamCounter
    server-1  | 2025/10/30 13:01:40 Received StreamCounter request with limit: 5
    client-1  | 2025/10/30 13:01:40 Outgoing stream RPC: /proto.Greeter/StreamCounter
    client-1  | 2025/10/30 13:01:41 Stream Counter: 1
    client-1  | 2025/10/30 13:01:42 Stream Counter: 2
    client-1  | 2025/10/30 13:01:43 Stream Counter: 3
    client-1  | 2025/10/30 13:01:44 Stream Counter: 4
    client-1  | 2025/10/30 13:01:45 Stream Counter: 5
    client-1  | 2025/10/30 13:01:45 Incoming stream RPC: /proto.Greeter/StreamCounter, Error: <nil>
    client-1  | 2025/10/30 13:01:45 StreamCounter finished
    client-1  | 2025/10/30 13:01:45 Calling Chat (bidirectional streaming)
    server-1  | 2025/10/30 13:01:45 Auth successful for stream method: /proto.Greeter/Chat
    client-1  | 2025/10/30 13:01:45 Outgoing stream RPC: /proto.Greeter/Chat
    server-1  | 2025/10/30 13:01:45 Chat message from Docker: Hello from client 0
    client-1  | 2025/10/30 13:01:45 Received chat message from Server: Echo: Hello from client 0
    server-1  | 2025/10/30 13:01:46 Chat message from Docker: Hello from client 1
    client-1  | 2025/10/30 13:01:46 Received chat message from Server: Echo: Hello from client 1
    server-1  | 2025/10/30 13:01:47 Chat message from Docker: Hello from client 2
    client-1  | 2025/10/30 13:01:47 Received chat message from Server: Echo: Hello from client 2
    client-1  | 2025/10/30 13:01:48 Incoming stream RPC: /proto.Greeter/Chat, Error: <nil>
    client-1  | 2025/10/30 13:01:48 Chat finished
    client-1 exited with code 0
    ```

5.  **Access the Web Application:** Open your web browser and go to `http://localhost:8080`.
    - The React application will load, and you can interact with the gRPC services through its UI.
    - Navigate to the **Files** page to test file operations:
      - **Upload**: Select a file (click or drag and drop) and upload it to the chosen storage provider
      - **List**: View all uploaded files with their metadata (filename, size, namespace)
      - **Preview**: Preview image, PDF, and text files directly in the browser
      - **Download**: Download files from storage
      - **Delete**: Remove files from storage and database
    - Navigate to the **OCR** page to process images and documents:
      - **Process OCR**: Select an uploaded image file and trigger OCR processing
      - **List Results**: View all OCR processing results with metadata
      - **Get Result**: View detailed OCR result for a specific file and engine
      - **Compare Results**: Compare OCR results from different engines (when multiple engines are available)
    - You can switch between different storage providers:
      - **AWS S3 (Localstack)**: Uses Localstack emulator
      - **Google Cloud Storage (fake-gcs)**: Uses fake-gcs-server emulator
      - **Azure Blob Storage (Azurite)**: Uses Azurite emulator
    - Files are automatically categorized into namespaces (documents/media/others) based on file type

6.  **Stop the Application:** To stop and remove the containers, press `Ctrl+C` in the terminal where `docker-compose up` is running (if running in foreground mode). For detached mode, or to clean up:
    ```bash
    docker-compose down
    ```
    
    To also remove volumes and networks:
    ```bash
    docker-compose down -v
    ```

## Storage Providers

This application supports multiple cloud storage providers with local emulators:

| Provider | Emulator | Port | Description |
|----------|----------|------|-------------|
| AWS S3 | Localstack | 4566 | Full S3 API emulation |
| Google Cloud Storage | fake-gcs-server | 4443 | GCS API emulation |
| Azure Blob Storage | Azurite | 10000 | Azure Storage emulation |

All storage providers can be selected from the web UI, and files uploaded/downloaded will be stored in the respective emulator.

## Queue System for OCR Processing

This application uses a queue-based architecture for asynchronous OCR task processing:

| Storage Provider | Queue Service | Emulator | Description |
|-----------------|---------------|----------|-------------|
| AWS S3 | AWS SQS | Localstack | SQS queue (`ocr-tasks-queue`) for OCR task queuing |
| Google Cloud Storage | GCP Pub/Sub | pubsub-emulator | Pub/Sub topic (`ocr-tasks`) and subscription (`ocr-tasks-subscription`) for OCR task processing |
| Azure Blob Storage | Azure Queue Storage | Azurite | Azure Queue Storage (`ocr-tasks-queue`) using Azurite Queue service (port 10001). Falls back to in-memory queue if emulator is unavailable |

**How it works:**
1. When a file is uploaded to `documents/` or `images/` namespace, an OCR task is automatically enqueued
2. The OCR service continuously dequeues tasks from the queue
3. OCR processing runs asynchronously in the background
4. Results are stored in the SQLite database

**Queue Configuration:**
- Each storage provider uses its native queue service
- Queue services automatically fall back to in-memory queues if emulators are unavailable
- Pub/Sub uses a dedicated Receive loop with proper context management for reliable message delivery
- **Azure Queue Storage Implementation:**
  - Uses Azure Queue Storage SDK (`github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue`)
  - Queue name: `ocr-tasks-queue` (automatically created if it doesn't exist)
  - Uses `DequeueMessages` API with `NumberOfMessages=1` and `VisibilityTimeout=30` seconds
  - Queue endpoint: Port 10001 (Blob Storage uses port 10000, Queue Storage uses port 10001)
  - Automatically detects and converts `AZURE_STORAGE_ENDPOINT` from Blob port (10000) to Queue port (10001)
  - Singleton pattern ensures a single `QueueClient` instance is shared across enqueue/dequeue operations
  - Comprehensive debug logging for troubleshooting queue operations

## Multi-Engine OCR Architecture

This application supports multiple OCR engines running in separate containers:

### Supported OCR Engines

| Engine | Language | Container | Port | Implementation |
|--------|----------|-----------|------|----------------|
| **Tesseract** | Go (CGO) | `ocr-tesseract-service` | 50052 | C++ library wrapper |
| **EasyOCR** | Python | `ocr-easyocr-service` | 50053 | PyTorch-based deep learning |

### Architecture Benefits

- **Image Size Optimization**: Each container contains only the dependencies needed for its specific engine
- **Independent Scaling**: Scale each engine independently based on workload
- **Resource Isolation**: Each engine has its own memory and CPU resources
- **Easy Engine Addition**: Add new engines by creating new containers without affecting existing ones

### How It Works

1. **File Upload**: When a file is uploaded to `documents/` or `images/` namespace, an OCR task is enqueued
2. **Queue Distribution**: The task is added to the storage provider's queue (SQS/Pub/Sub/Azure Queue)
3. **Parallel Processing**: Both OCR engine containers dequeue tasks from the same queue
4. **Result Storage**: Each engine saves its results to the SQLite database with `engine_name` to distinguish them
5. **Result Retrieval**: The server's `MultiOCRClient` manages connections to all engine endpoints for result comparison

### Environment Variables for OCR

```bash
# Multiple OCR engine endpoints (recommended)
OCR_TESSERACT_ENDPOINT=http://ocr-tesseract-service:50052
OCR_EASYOCR_ENDPOINT=http://ocr-easyocr-service:50053

# Legacy single endpoint (backward compatibility)
# OCR_SERVICE_ENDPOINT=ocr-service:50052

# Individual service port configuration
OCR_SERVICE_PORT=50052  # For ocr-tesseract-service
OCR_SERVICE_PORT=50053  # For ocr-easyocr-service
OCR_ENGINES=tesseract   # Engine registration for each container
EASYOCR_ENABLED=true    # Enable EasyOCR (for ocr-easyocr-service)
```

### Docker Images

- **ocr-tesseract-service**: ~200MB (Alpine-based, Tesseract dependencies only)
- **ocr-easyocr-service**: ~12GB (Debian-based, includes PyTorch and EasyOCR)

For more details, see [doc/MULTI_ENGINE_OCR_PROGRESS.md](doc/MULTI_ENGINE_OCR_PROGRESS.md).

## Features

- **gRPC Communication**: Unary, server streaming, client streaming, and bidirectional streaming
- **Authentication**: Token-based authentication for gRPC calls
- **File Operations**: 
  - Upload files to multiple cloud storage providers (click to select or drag and drop)
  - Preview image, PDF, and text files directly in the browser
  - Download files from storage
  - List uploaded files with metadata
  - Delete files from storage
  - Automatic file categorization by type (documents, media, others)
- **OCR (Optical Character Recognition)**:
  - **Multi-Engine Support**: Process images and documents using multiple OCR engines (Tesseract and EasyOCR)
  - **Engine-Per-Container Architecture**: Each OCR engine runs in a separate Docker container for optimal resource usage and scalability
  - **Tesseract OCR**: Fast, C++-based OCR engine with Japanese and English language support
  - **EasyOCR**: Python-based OCR engine using deep learning (PyTorch) for improved accuracy
  - Extract text from uploaded image files
  - View OCR processing results with confidence scores
  - List and manage OCR results for multiple files and engines
  - Compare OCR results from different engines
  - OCR services run as separate microservices for scalability
  - Automatic OCR task queuing for `documents/` and `images/` files
  - Queue-based asynchronous processing using provider-specific queues (SQS for S3, Pub/Sub for GCS, Azure Queue for Azure)
  - **MultiOCRClient**: Manages connections to multiple OCR engine endpoints simultaneously
- **SQLite Database**: File metadata management for tracking uploaded files and OCR results
- **Storage Emulators**: Local development support for AWS S3, GCS, and Azure Blob Storage
- **Web UI**: 
  - React-based frontend (TypeScript) with React Bootstrap components
  - Drag and drop file selection for easy file uploads
  - File preview modal for images, PDFs, and text files
  - Modal dialogs for user feedback and confirmations
  - File list with namespace badges (documents/media/others)
  - Storage provider selection (S3/GCS/Azure)
  - OCR results page for viewing and managing OCR processing
- **File Namespace Classification**: Files are automatically categorized into namespaces:
  - `documents/`: Document files (PDF, DOC, TXT, etc.)
  - `media/`: Image and video files (JPG, PNG, MP4, etc.)
  - `others/`: Other file types