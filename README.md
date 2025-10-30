# gRPC Sample Application in Go

This is a minimal gRPC sample application implemented in Go, demonstrating basic client-server communication using Protocol Buffers and gRPC.

## Project Structure

- `proto/`: Contains the Protocol Buffer definition (`greeter.proto`) and the generated Go code.
- `server/`: Implements the gRPC server.
- `client/`: Implements the gRPC client.
- `webapp/`: Contains a simple Go web application that interacts with the gRPC server.
- `Dockerfile.server`: Dockerfile for building the gRPC server image.
- `Dockerfile.client`: Dockerfile for building the gRPC client image.
- `Dockerfile.webapp`: Dockerfile for building the web application image.
- `docker-compose.yml`: Defines and runs the multi-container Docker application.

## How to Run

This application uses Docker Compose for easy setup and execution.

1.  **Ensure Docker is Running:** Make sure Docker Desktop or Docker Engine is running on your system.

2.  **Navigate to the Project Root:** Open your terminal and navigate to the root directory of this project:
    ```bash
    cd /path/to/grpc-sample-minimal
    ```

3.  **Build and Run with Docker Compose:** Execute the following command to build the Docker images and start the server, client, and webapp containers:
    ```bash
    docker-compose up --build
    ```

    You should see output from the `server` and `client` services in your terminal. The `webapp` will be accessible via your browser.

    **Example Client Output (from `client` service):**
    ```
    server-1  | 2025/10/30 12:52:28 server listening at [::]:50051
    server-1  | 2025/10/30 12:52:29 Received: Docker
    client-1  | 2025/10/30 12:52:29 Greeting: Hello Docker
    client-1  | 2025/10/30 13:01:40 Received StreamCounter request with limit: 5
    client-1  | 2025/10/30 13:01:41 Stream Counter: 1
    client-1  | 2025/10/30 13:01:42 Stream Counter: 2
    client-1  | 2025/10/30 13:01:43 Stream Counter: 3
    client-1  | 2025/10/30 13:01:44 Stream Counter: 4
    client-1  | 2025/10/30 13:01:45 Stream Counter: 5
    client-1  | 2025/10/30 13:01:45 StreamCounter finished
    client-1  | 2025/10/30 13:01:45 Calling Chat (bidirectional streaming)
    server-1  | 2025/10/30 13:01:45 Chat message from Docker: Hello from client 0
    client-1  | 2025/10/30 13:01:45 Received chat message from Server: Echo: Hello from client 0
    server-1  | 2025/10/30 13:01:46 Chat message from Docker: Hello from client 1
    client-1  | 2025/10/30 13:01:46 Received chat message from Server: Echo: Hello from client 1
    server-1  | 2025/10/30 13:01:47 Chat message from Docker: Hello from client 2
    client-1  | 2025/10/30 13:01:47 Received chat message from Server: Echo: Hello from client 2
    client-1  | 2025/10/30 13:01:48 Chat finished
    client-1 exited with code 0
    ```

4.  **Access the Web Application:** Open your web browser and go to `http://localhost:8080`.
    - Enter a name in the form and click "Say Hello" to test the unary RPC.
    - Click "Start Counter Stream" to see the server-side streaming in action.
    - Use the chat interface to send messages and see them echoed back.

5.  **Stop the Application:** To stop and remove the containers, press `Ctrl+C` in the terminal where `docker-compose up` is running. Then, you can optionally remove the volumes and networks:
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
