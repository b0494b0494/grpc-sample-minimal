# gRPC Sample Application in Go

This is a minimal gRPC sample application implemented in Go, demonstrating basic client-server communication using Protocol Buffers and gRPC.

## Project Structure

- `proto/`: Contains the Protocol Buffer definition (`greeter.proto`) and the generated Go code.
- `server/`: Implements the gRPC server with a layered architecture (domain, application, infrastructure) including authentication and logging interceptors.
  - `server/domain/`: Domain layer with storage service implementations (S3, GCS, Azure Blob Storage) and SQLite database repository
  - `server/application/`: Application layer service orchestrating gRPC calls and file operations
- `client/`: Implements the gRPC client with authentication and logging interceptors.
- `webapp/`: Contains a React frontend application (TypeScript with React Bootstrap, API service layer, custom hooks, and component-based structure) and a Go backend that exposes HTTP API endpoints for gRPC calls.
  - `webapp/src/components/`: React components including `AlertDialog` for modal dialogs
  - `webapp/src/hooks/`: Custom React hooks for file operations
  - `webapp/src/services/`: API service layer for gRPC calls
- `Dockerfile.server`: Dockerfile for building the gRPC server image.
- `Dockerfile.client`: Dockerfile for building the gRPC client image.
- `Dockerfile.webapp`: Dockerfile for building the web application image (React frontend + Go backend).
- `docker-compose.yml`: Defines and runs the multi-container Docker application with storage emulators (Localstack, fake-gcs, Azurite).

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
    DB_PATH=/app/data/files.db
    ```
    *Note: The `AUTH_TOKEN` is used for gRPC authentication. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `S3_BUCKET_NAME`, and `LOCALSTACK_ENDPOINT` are for Localstack S3 integration. `GRPC_SERVER_PORT` defines the port the gRPC server listens on. `DB_PATH` specifies the path to the SQLite database file for file metadata storage.*
    
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

    You should see output from the `server`, `client`, and storage emulator services in your terminal. The `webapp` will be accessible via your browser.

    **Storage Emulators:**
    - **Localstack:** The `localstack` service emulates AWS S3 locally. The server will attempt to create an S3 bucket named `grpc-sample-bucket` on startup.
    - **fake-gcs:** The `fake-gcs` service emulates Google Cloud Storage locally.
    - **Azurite:** The `azurite` service emulates Azure Blob Storage locally (runs on port 10000 for Blob service).

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
      - **Upload**: Select a file and upload it to the chosen storage provider
      - **List**: View all uploaded files with their metadata (filename, size, namespace)
      - **Download**: Download files from storage
      - **Delete**: Remove files from storage and database
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

## Features

- **gRPC Communication**: Unary, server streaming, client streaming, and bidirectional streaming
- **Authentication**: Token-based authentication for gRPC calls
- **File Operations**: 
  - Upload files to multiple cloud storage providers
  - Download files from storage
  - List uploaded files with metadata
  - Delete files from storage
  - Automatic file categorization by type (documents, media, others)
- **SQLite Database**: File metadata management for tracking uploaded files
- **Storage Emulators**: Local development support for AWS S3, GCS, and Azure Blob Storage
- **Web UI**: 
  - React-based frontend (TypeScript) with React Bootstrap components
  - Modal dialogs for user feedback and confirmations
  - File list with namespace badges (documents/media/others)
  - Storage provider selection (S3/GCS/Azure)
- **File Namespace Classification**: Files are automatically categorized into namespaces:
  - `documents/`: Document files (PDF, DOC, TXT, etc.)
  - `media/`: Image and video files (JPG, PNG, MP4, etc.)
  - `others/`: Other file types