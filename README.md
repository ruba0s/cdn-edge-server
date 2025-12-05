# CDN Edge Server

A content delivery network (CDN) edge server implementation in Go with HTTP/1.0 support, FIFO caching, and an interactive CLI for testing.

## Architecture Overview

### System Components

```
┌─────────┐         ┌──────────────┐         ┌────────────────┐
│  Client │ ◄─────► │  Edge Server │ ◄─────► │ Origin Server  │
│  (CLI)  │         │  + Cache     │         │   + Storage    │
└─────────┘         └──────────────┘         └────────────────┘
```

**Client (CLI)**: Interactive terminal application for sending HTTP requests and testing server functionality.

**Edge Server**: Proxy server with FIFO cache that intercepts client requests:
- Cache hit → Serves from local cache
- Cache miss → Fetches from origin, caches result, returns to client
- Handles GET, HEAD, POST, PUT requests
- Invalidates cache on PUT/POST operations

**Origin Server**: Authoritative file storage server that stores and serves files from disk.

### Project Structure

```
cdn-edge-server/
├── cmd/
│   ├── cli/main.go          # CLI client entry point
│   ├── edge/main.go         # Edge server entry point
│   └── origin/main.go       # Origin server entry point
├── internal/
│   ├── cache/
│   │   ├── fifo.go          # FIFO cache implementation
│   │   └── files/           # Cached files storage
│   ├── edge/
│   │   ├── handler.go       # Edge server request handler
│   │   └── tcp_server.go    # TCP server wrapper
│   ├── origin/
│   │   └── handler.go       # Origin server request handler
│   ├── http/
│   │   ├── parser.go        # HTTP request/response parser
│   │   └── response.go      # HTTP response builder
│   ├── storage/
│   │   └── files/           # Origin server file storage
│   ├── ui/
│   │   └── terminal.go      # Interactive CLI implementation
│   └── config/
│       └── config.go        # Configuration loader
│
├── .env.template            # Environment template
├── go.mod
└── README.md
```

## Implementation Details

### Cache Implementation (FIFO)
- **Data structures**: 
  - `queue []string` - Maintains insertion order
  - `present map[string]bool` - O(1) lookup for cache hits
- **Capacity**: 5 files (configurable via `MaxCacheFiles`)
- **Eviction**: When cache is full, oldest file (front of queue) is removed
- **Cache invalidation**: PUT/POST requests remove stale cached files

### HTTP Protocol
- **Version**: HTTP/1.0
- **Connection model**: One request per connection (non-persistent)
- **Supported methods**: GET, HEAD, POST, PUT
- **Content-Type detection**: Based on file extension via `mime.TypeByExtension()`

### Concurrency
- **Edge server**: Each client connection handled in a separate goroutine
- **Origin server**: Each client connection handled in a separate goroutine
- **Thread safety**: Cache operations are single-process (no mutex needed as Go scheduler handles goroutines)

## Setup

### Prerequisites
- Go 1.25.1 or higher
- Terminal/Command prompt

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/ruba0s/cdn-edge-server
cd cdn-edge-server
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment**
```bash
cp .env.template .env
```

Default configuration:
```env
EDGE_HOST=127.0.0.1
EDGE_PORT=8080
ORIGIN_HOST=127.0.0.1
ORIGIN_PORT=4396
```

## Running the System

The system requires three separate terminal windows running concurrently.

### Terminal 1: Start Origin Server
```bash
go run cmd/origin/main.go
```

**Output:**
```
Origin server running on 127.0.0.1:4396 ...
```

The origin server stores and serves files from `internal/storage/files/`.

---

### Terminal 2: Start Edge Server
```bash
go run cmd/edge/main.go
```

**Output:**
```
Server listening on 127.0.0.1:8080
```

The edge server proxies requests and caches files in `internal/cache/files/`.

---

### Terminal 3: Start CLI Client
```bash
go run cmd/cli/main.go
```

**Output:**
```
   ____ ____  _   _   _____    _              ____                           
  / ___|  _ \| \ | | | ____|__| | __ _  ___  / ___|  ___ _ ____   _____ _ __ 
 | |   | | | |  \| | |  _| / _` |/ _` |/ _ \ \___ \ / _ \ '__\ \ / / _ \ '__/
 | |___| |_| | |\  | | |__| (_| | (_| |  __/  ___) |  __/ |   \ V /  __/ |   
  \____|____/|_| \_| |_____\__,_|\__, |\___| |____/ \___|_|    \_/ \___|_|   
                                 |___/                                        

┌─ Main Menu ─────────────────────────────┐
│ 1. Check Server Status                  │
│ 2. Send requests to edge server         │
│ 3. View Configuration                   │
│ 4. Exit                                 │
└─────────────────────────────────────────┘
```

## CLI Menu Options

### 1. Check Server Status
Verifies that both edge and origin servers are running and accessible.

**Example output:**
```
Checking Server Status...
═══════════════════════════════════════
Edge Server: Running on 127.0.0.1:8080
Origin Server: Running on 127.0.0.1:4396
═══════════════════════════════════════
```

Displays startup instructions if either server is not running.

---

### 2. Send Requests to Edge Server
Interactive submenu for sending HTTP requests to the edge server.

#### 2.1 GET Request
Retrieves a file from the edge server (cached or from origin).

**Steps:**
1. Select option `1` from the Send Requests menu
2. Enter filename (e.g., `test.txt`)

**Example:**
```
Enter filename: test.txt

Sending GET request for 'test.txt'...

Response:
═══════════════════════════════════════
HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 12

Hello World!
═══════════════════════════════════════
```

**Cache behavior:**
- First request: Cache miss → Fetches from origin → Caches file → Returns to client
- Second request: Cache hit → Returns from cache (faster)

---

#### 2.2 HEAD Request
Retrieves only the headers/metadata of a file without the body.

**Steps:**
1. Select option `2` from the Send Requests menu
2. Enter filename (e.g., `image.jpg`)

**Example:**
```
Enter filename: image.jpg

 Sending HEAD request for 'image.jpg'...

 Response:
═══════════════════════════════════════
HTTP/1.0 200 OK
Content-Type: image/jpeg
Content-Length: 54321
═══════════════════════════════════════
```

**Use case:** Check if a file exists and get its size without downloading it.

---

#### 2.3 POST Request (Create File)
Creates a new file on the origin server.

**Steps:**
1. Select option `3` from the Send Requests menu
2. Enter filename (e.g., `newfile.txt`)
3. Enter body content (e.g., `This is new content`)

**Example:**
```
Enter filename: newfile.txt
Enter body content (optional): This is new content

 Sending POST request to create 'newfile.txt'...

 Response:
═══════════════════════════════════════
HTTP/1.0 200 OK
Created: newfile.txt
═══════════════════════════════════════
```

**Note:** POST fails if file already exists (returns 400 Bad Request).

---

#### 2.4 PUT Request (Update File)
Updates an existing file or creates it if it doesn't exist.

**Steps:**
1. Select option `4` from the Send Requests menu
2. Enter filename (e.g., `test.txt`)
3. Enter new body content (e.g., `Updated content`)

**Example:**
```
Enter filename: test.txt
Enter body content (optional): Updated content

 Sending PUT request to update 'test.txt'...

 Response:
═══════════════════════════════════════
HTTP/1.0 200 OK
Updated: test.txt
═══════════════════════════════════════
```

**Cache invalidation:** If the file was cached, it's automatically removed from the cache. The next GET request will fetch the updated version from origin.

---

### 3. View Configuration
Displays current server configuration and directory paths.

**Example output:**
```
  Configuration
═══════════════════════════════════════
Edge Server:   127.0.0.1:8080
Origin Server: 127.0.0.1:4396
Cache Dir:     /path/to/internal/cache/files
Storage Dir:   /path/to/internal/storage/files
```

---

### 4. Exit
Gracefully exits the CLI application.

```
CLI Exited
```

**Note:** Exiting the CLI does not stop the edge or origin servers. Stop them manually with `Ctrl+C` in their respective terminals.

## Testing Cache Behavior

### Test Scenario: Cache Hit vs Cache Miss

1. **Create a test file**
   - CLI → Send Requests → POST
   - Filename: `cache-test.txt`
   - Body: `Original content`

2. **First GET (Cache Miss)**
   ```
   GET cache-test.txt
   ```
   - Edge server logs: Fetching from origin
   - Response time: ~slower (network + origin)

3. **Second GET (Cache Hit)**
   ```
   GET cache-test.txt
   ```
   - Edge server logs: Cache hit
   - Response time: ~faster (local disk)

4. **Update file (Cache Invalidation)**
   - CLI → Send Requests → PUT
   - Filename: `cache-test.txt`
   - Body: `Updated content`
   - Cache is cleared for this file

5. **Third GET (Cache Miss Again)**
   ```
   GET cache-test.txt
   ```
   - Edge server fetches updated version from origin
   - Response body shows: `Updated content`

### Test Scenario: Cache Eviction (FIFO)

The cache holds a maximum of 5 files. When a 6th file is requested:

1. **Fill the cache** (GET 5 different files)
   - `file1.txt`, `file2.txt`, `file3.txt`, `file4.txt`, `file5.txt`
   - Cache queue: `[file1, file2, file3, file4, file5]`

2. **Request 6th file**
   - GET `file6.txt`
   - Cache queue: `[file2, file3, file4, file5, file6]`
   - `file1.txt` (oldest file) is evicted

3. **Request file1 again**
   - Cache miss → Fetches from origin again
   - Cache queue: `[file3, file4, file5, file6, file1]`
   - `file2.txt` is evicted

## Testing Concurrency

(To be added)

### Using curl
```bash
# Send 10 concurrent requests
for i in {1..10}; do
  curl http://127.0.0.1:8080/test.txt &
done
wait
```

All requests complete successfully, demonstrating proper concurrent handling.

## Error Handling

### Common HTTP Status Codes

| Code | Status | Meaning |
|------|--------|---------|
| 200 | OK | Request successful |
| 400 | Bad Request | Malformed request or POST to existing file |
| 404 | Not Found | File doesn't exist on origin |
| 405 | Method Not Allowed | Unsupported HTTP method |
| 500 | Internal Server Error | Edge server error (e.g., cache read failure) |
| 502 | Bad Gateway | Cannot connect to origin server |

### Troubleshooting

**Problem:** CLI shows "Edge server is not running"
- **Solution:** Start edge server in Terminal 2: `go run cmd/edge/main.go`

**Problem:** Edge server shows "502 Bad Gateway"
- **Solution:** Start origin server in Terminal 1: `go run cmd/origin/main.go`

**Problem:** GET returns 404
- **Solution:** File doesn't exist. Create it first with POST request.

**Problem:** POST returns 400
- **Solution:** File already exists. Use PUT to update instead.