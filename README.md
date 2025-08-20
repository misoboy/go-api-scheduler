# Go API Scheduler

A simple web application built with Go that allows you to schedule repetitive API calls based on a user-defined schedule. The application includes a fake server for testing and logs the results in real-time.

## Features

* **Customizable Schedule:** Set a specific start time and a repeat interval (hours, minutes, or seconds).

* **API Calls:** Configure the API URL, HTTP method (GET/POST), and payload (as key-value pairs).

* **Real-time Logging:** View API call results and scheduler status on a console log screen.

* **Fake Server:** A built-in fake server for easy testing. It returns a `200 OK` status and the exact payload sent by the client.

* **Automatic Stop:** The scheduler automatically stops once an API call receives a `200 OK` response.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* Go (version 1.18 or higher)

* A web browser

### Project Structure

```
go-api-scheduler/
├── cmd/
│   └── api-scheduler/
│       └── main.go       # Main application entry point
├── internal/
│   ├── handler/
│   │   └── handler.go    # HTTP handlers and fake server logic
│   ├── logger/
│   │   └── logger.go     # Logging functionalities
│   └── scheduler/
│       └── scheduler.go  # Core scheduling logic
├── web/
│   └── static/
│       └── index.html    # Web UI (HTML, CSS, JS)
└── go.mod                # Go module file

```

### Building and Running the Application

1. **Create Project Directory:**
   Start by creating the main project directory and its subdirectories.

   ```
   mkdir go-api-scheduler
   cd go-api-scheduler
   mkdir -p cmd/api-scheduler internal/handler internal/logger internal/scheduler web/static
   
   ```

2. **Create Go Module:**
   Initialize the Go module in the root directory.

   ```
   go mod init go-api-scheduler
   go mod tidy
   
   ```

3. **Create Files:**
   Copy the code for `main.go`, `handler.go`, `logger.go`, `scheduler.go`, and `index.html` into their respective files within the project structure.

4. **Build the Project:**
   To avoid permission errors, it's recommended to build the executable first. Open a command prompt or PowerShell as an administrator.

   ```
   go build -o scheduler-app ./cmd/api-scheduler
   
   ```

5. **Run the Application:**
   Execute the built application.

   ```
   .\scheduler-app.exe
   
   ```

6. **Access the Web UI:**
   Open your web browser and navigate to `http://localhost:8080`.

You can now configure your scheduler and test it using the built-in fake server.
