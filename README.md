Diagnostic Client API Documentation
Overview
The Diagnostic Client API provides real-time monitoring and analysis capabilities for system diagnostics, including file system monitoring, network traffic analysis, and log aggregation.
Base URL
Copyhttp://localhost:8080
WebSocket Endpoint
Real-time Updates Connection
CopyGET /ws
Once connected, the WebSocket will stream updates in the following formats:
File Update Message
jsonCopy{
    "type": "file_update",
    "payload": {
        "path": "/path/to/file",
        "parent_path": "/path/to",
        "name": "file",
        "is_directory": false,
        "size": 1024,
        "mod_time": "2024-11-02T03:18:43Z",
        "is_gzipped": false,
        "is_scraped": false
    }
}
Network Update Message
jsonCopy{
    "type": "network",
    "payload": [{
        "timestamp": "2024-11-02T03:18:43Z",
        "protocol": "TCP",
        "src_ip": "192.168.1.1",
        "dst_ip": "192.168.1.2",
        "src_port": 8080,
        "dst_port": 443,
        "length": 1024,
        "payload_size": 512,
        "tcp_flags": "ACK"
    }]
}
Log Update Message
jsonCopy{
    "type": "log",
    "payload": {
        "filename": "/var/log/system.log",
        "line": "Error: Connection refused",
        "line_num": 1234,
        "timestamp": "2024-11-02T03:18:43Z",
        "level": "ERROR"
    }
}
REST API Endpoints
File System Operations
Get File Tree
CopyGET /api/files
Retrieves a hierarchical view of the file system.
Query Parameters:

path (string, optional)

Root path to start traversal
Default: "/"


depth (integer, optional)

Depth of tree traversal
Default: 1
Maximum: 10



Success Response (200 OK):
jsonCopy[
    {
        "path": "/path/to/file",
        "parent_path": "/path/to",
        "name": "file",
        "is_directory": false,
        "size": 1024,
        "mod_time": "2024-11-02T03:18:43Z",
        "is_gzipped": false,
        "is_scraped": false
    }
]
Log Operations
Get Logs
CopyGET /api/logs
Retrieves log entries for a specific file with pagination.
Query Parameters:

file (string, required)

Path to the log file


before (string, optional)

ISO timestamp for pagination
Format: "2024-11-02T03:18:43Z"


limit (integer, optional)

Maximum number of entries to return
Default: 100
Maximum: 1000



Success Response (200 OK):
jsonCopy[
    {
        "filename": "/var/log/system.log",
        "line": "Error: Connection refused",
        "line_num": 1234,
        "timestamp": "2024-11-02T03:18:43Z",
        "level": "ERROR"
    }
]
Search Logs
CopyPOST /api/logs/search
Searches log entries across multiple files with full-text search capabilities.
Request Body:
jsonCopy{
    "query": "error connection",
    "files": ["/var/log/system.log", "/var/log/app.log"],
    "start_time": "2024-11-01T00:00:00Z",
    "end_time": "2024-11-02T00:00:00Z"
}
Success Response (200 OK):
jsonCopy[
    {
        "filename": "/var/log/system.log",
        "line": "Error: Connection refused",
        "line_num": 1234,
        "timestamp": "2024-11-02T03:18:43Z",
        "level": "ERROR"
    }
]
Network Operations
Get Network Metrics
CopyGET /api/network/metrics
Retrieves network traffic metrics and packet information.
Query Parameters:

start (string, optional)

Start time for metrics
Format: "2024-11-01T00:00:00Z"


end (string, optional)

End time for metrics
Format: "2024-11-02T00:00:00Z"


protocol (string[], optional)

Filter by protocols (e.g., TCP, UDP)



Success Response (200 OK):
jsonCopy{
    "packet_count": 1000,
    "total_bytes": 1048576,
    "avg_packet_size": 1024,
    "unique_sources": 10,
    "unique_destinations": 20,
    "protocol_count": 3,
    "protocol_stats": {
        "TCP": 800,
        "UDP": 150,
        "ICMP": 50
    },
    "packets": [
        {
            "timestamp": "2024-11-02T03:18:43Z",
            "protocol": "TCP",
            "src_ip": "192.168.1.1",
            "dst_ip": "192.168.1.2",
            "src_port": 8080,
            "dst_port": 443,
            "length": 1024,
            "payload_size": 512,
            "tcp_flags": "ACK"
        }
    ]
}
Error Responses
All endpoints use standard HTTP status codes and return errors in the following format:
jsonCopy{
    "error": "Detailed error message",
    "code": "ERROR_CODE",
    "details": {
        "field": "Additional error context"
    }
}
Common Status Codes:

200: Successful operation
400: Bad request (invalid parameters)
404: Resource not found
500: Internal server error

Common Error Codes:

INVALID_PATH: Invalid file path provided
FILE_NOT_FOUND: Requested file does not exist
INVALID_TIME_RANGE: Invalid time range specified
INVALID_PROTOCOL: Unsupported protocol specified
SEARCH_ERROR: Error during log search
DATABASE_ERROR: Database operation failed