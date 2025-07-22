# Testing Guide for Object Storage Gateway

This guide provides comprehensive instructions on how to test the Object Storage Gateway application.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Using the Makefile](#using-the-makefile)
3. [Testing Key Functionality](#testing-key-functionality)
4. [Troubleshooting](#troubleshooting)

## Prerequisites

Before you begin testing the Object Storage Gateway, ensure you have the following prerequisites installed:

1. **Docker and Docker Compose**: Required to run the Minio instances and the gateway application.
   ```bash
   # Check if Docker is installed
   docker --version

   # Check if Docker Compose is installed
   docker-compose --version
   ```

2. **Go (version 1.23 or later)**: Required to build and run the tests.
   ```bash
   # Check if Go is installed
   go version
   ```

3. **curl**: Useful for manual testing of the API endpoints.
   ```bash
   # Check if curl is installed
   curl --version
   ```

4. **Git**: Required to clone the repository.
   ```bash
   # Check if Git is installed
   git version
   ```

5. **Make**: Required to use the Makefile commands.
   ```bash
   # Check if Make is installed
   make --version
   ```

## Using the Makefile

A Makefile has been provided to simplify common testing tasks. Here are the available commands:

```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage and generate a report
make start          # Start the application
make stop           # Stop the application
make clean          # Clean build artifacts
make help           # Show help information
```

Using the Makefile is the recommended way to run tests and other common tasks.

## Testing Key Functionality

This section focuses on testing the specific functionality requirements mentioned in the README.md.

### Testing the PUT Endpoint

The PUT endpoint allows you to store an object with a specific ID. Here's how to test it:

```bash
# Start the application if it's not already running
make start

# Create a test file
echo "This is a test object" > test_object.txt

# Upload the object with ID "test123"
curl -X PUT --data-binary @test_object.txt http://localhost:3000/object/test123

# You should see a 200 OK response if the upload was successful
```

You can also test with different object sizes:

```bash
# Create a larger test file (1MB)
dd if=/dev/urandom of=large_test_object.bin bs=1M count=1

# Upload the large object
curl -X PUT --data-binary @large_test_object.bin http://localhost:3000/object/large123
```

### Testing the GET Endpoint

The GET endpoint allows you to retrieve an object by its ID. Here's how to test it:

```bash
# Retrieve the object you just uploaded
curl http://localhost:3000/object/test123 > retrieved_object.txt

# Compare the retrieved object with the original
diff test_object.txt retrieved_object.txt
# If there's no output, the files are identical

# Retrieve the large object
curl http://localhost:3000/object/large123 > retrieved_large_object.bin

# Compare the retrieved large object with the original
diff large_test_object.bin retrieved_large_object.bin
```

### Testing Error Handling

The application should handle various error conditions gracefully:

1. **Testing 404 Not Found**:
```bash
# Try to get a non-existent object
curl -v http://localhost:3000/object/nonexistent
# You should see a 404 Not Found response
```

2. **Testing Invalid Object ID**:
```bash
# Try to put an object with an invalid ID (e.g., with special characters)
curl -v -X PUT -d "test data" "http://localhost:3000/object/invalid@id"
# You should see a 400 Bad Request response
```

3. **Testing Empty Request Body**:
```bash
# Try to put an object with an empty body
curl -v -X PUT -d "" http://localhost:3000/object/test123
# You should see a 400 Bad Request response
```

## Verifying Consistent Instance Selection

The gateway should consistently choose the same Minio instance for the same object ID. You can verify this by:

1. **Checking the logs**:
```bash
# Start the application with logs visible
docker-compose up

# In another terminal, upload an object
curl -X PUT -d "test data" http://localhost:3000/object/test123

# Look for log messages indicating which Minio instance was used
# You should see something like: "Successfully stored object test123 in instance amazin-object-storage-node-X"

# Upload the same object again
curl -X PUT -d "updated data" http://localhost:3000/object/test123

# Check the logs again - it should use the same instance
```

2. **Using the Minio console**:
```bash
# The Minio console is available at:
# - Node 1: http://localhost:9001
# - Node 2: http://localhost:9002
# - Node 3: http://localhost:9003

# Log in using the credentials from docker-compose.yml:
# - Node 1: Username: ring, Password: treepotato
# - Node 2: Username: maglev, Password: baconpapaya
# - Node 3: Username: rendezvous, Password: bluegreen

# Navigate to the bucket for each node and check if your object exists
# It should only exist in one of the nodes
```

3. **Testing with different object IDs**:
```bash
# Upload objects with different IDs
curl -X PUT -d "data for object 1" http://localhost:3000/object/object1
curl -X PUT -d "data for object 2" http://localhost:3000/object/object2
curl -X PUT -d "data for object 3" http://localhost:3000/object/object3

# Check the logs to see which instance was used for each object
# They should be distributed across the available instances
```

## Troubleshooting

This section provides solutions to common issues you might encounter when testing the Object Storage Gateway.

### Application Won't Start

If the application fails to start, check the following:

1. **Docker daemon is running**:
   ```bash
   # Check if Docker daemon is running
   docker info
   ```

2. **Port conflicts**:
   ```bash
   # Check if port 3000 is already in use
   lsof -i :3000
   # If it is, stop the process using that port or change the port in server.go
   ```

3. **Docker network issues**:
   ```bash
   # Check if the amazin-object-storage network exists
   docker network ls | grep amazin-object-storage

   # If not, recreate it
   docker-compose down
   docker-compose up -d
   ```

### Cannot Connect to Minio Instances

If the gateway can't connect to Minio instances:

1. **Check if Minio containers are running**:
   ```bash
   docker ps | grep amazin-object-storage-node
   ```

2. **Check Minio logs**:
   ```bash
   docker logs amazin-object-storage-node-1
   ```

3. **Verify network connectivity**:
   ```bash
   # From the gateway container
   docker exec -it gateway-container ping amazin-object-storage-node-1
   ```

### PUT/GET Requests Failing

If your PUT or GET requests are failing:

1. **Check application logs**:
   ```bash
   docker logs gateway-container
   ```

2. **Verify the object ID format**:
   - Object IDs must be alphanumeric and up to 32 characters

3. **Check request format**:
   - For PUT requests, ensure you're sending a non-empty body
   - For GET requests, ensure the object ID is correct

### Inconsistent Instance Selection

If you notice that the same object ID is being routed to different Minio instances:

1. **Check for code changes in the hashing function**:
   - The `GetMinioInstanceForID` function in `utils/hash/hash.go` should be deterministic

2. **Verify that the list of Minio instances is stable**:
   - The instances should be sorted by name in `utils/docker/docker.go`

3. **Check for race conditions**:
   - The gateway uses a mutex to protect access to the instance list

### Test Failures

If tests are failing:

1. **Check Go version**:
   ```bash
   go version
   # Should be 1.23 or later
   ```

2. **Ensure dependencies are installed**:
   ```bash
   go mod download
   ```

3. **Run tests with verbose output**:
   ```bash
   go test -v ./...
   ```

4. **Check for environment-specific issues**:
   - Some tests may require Docker to be running
   - Integration tests may fail if network conditions change
