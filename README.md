# gRPC Sample Application in Go

This is a minimal gRPC sample application implemented in Go, demonstrating basic client-server communication using Protocol Buffers and gRPC.

## Project Structure

- `proto/`: Contains the Protocol Buffer definition (`greeter.proto`) and the generated Go code.
- `server/`: Implements the gRPC server.
- `client/`: Implements the gRPC client.
- `Dockerfile.server`: Dockerfile for building the gRPC server image.
- `Dockerfile.client`: Dockerfile for building the gRPC client image.
- `docker-compose.yml`: Defines and runs the multi-container Docker application.

## How to Run

This application uses Docker Compose for easy setup and execution.

1.  **Ensure Docker is Running:** Make sure Docker Desktop or Docker Engine is running on your system.

2.  **Navigate to the Project Root:** Open your terminal and navigate to the root directory of this project:
    ```bash
    cd /path/to/grpc-sample-minimal
    ```

3.  **Build and Run with Docker Compose:** Execute the following command to build the Docker images and start the server and client containers:
    ```bash
    docker-compose up --build
    ```

    You should see output similar to this, indicating that the server is listening and the client has successfully made a call:
    ```
    server-1  | 2025/10/30 12:52:28 server listening at [::]:50051
    server-1  | 2025/10/30 12:52:29 Received: Docker
    client-1  | 2025/10/30 12:52:29 Greeting: Hello Docker
    client-1  | 2025/10/30 12:52:29 Calling SayHelloServerStream for Docker
    server-1  | 2025/10/30 12:52:29 Received: Docker for server stream
    client-1  | 2025/10/30 12:52:30 Stream Greeting: Hello Docker, count 0
    client-1  | 2025/10/30 12:52:31 Stream Greeting: Hello Docker, count 1
    client-1  | 2025/10/30 12:52:32 Stream Greeting: Hello Docker, count 2
    client-1  | 2025/10/30 12:52:33 Stream Greeting: Hello Docker, count 3
    client-1  | 2025/10/30 12:52:34 Stream Greeting: Hello Docker, count 4
    client-1  | 2025/10/30 12:52:34 SayHelloServerStream finished
    client-1 exited with code 0
    ```

4.  **Stop the Application:** To stop and remove the containers, press `Ctrl+C` in the terminal where `docker-compose up` is running. Then, you can optionally remove the volumes and networks:
    ```bash
    docker-compose down
    ```

## Exploring Further

This sample provides a basic foundation. You can extend it by:

- Implementing more complex RPC types (server-side streaming, client-side streaming, bidirectional streaming).
- Adding error handling and gRPC status codes.
- Incorporating metadata, authentication, or authorization.
- Exploring interceptors for logging or metrics.
- Using different data types in your `.proto` definitions.
