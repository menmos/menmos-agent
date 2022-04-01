OUTPUT_DIRECTORY := "build/"

# Build the workspace.
build:
    mkdir -p {{OUTPUT_DIRECTORY}}
    go build -o {{OUTPUT_DIRECTORY}} ./cmd/menmos-agent

# Test the workspace
test:
    go test -cover -race -v ./...

# Delete the build directory.
clean:
    rm -rf {{OUTPUT_DIRECTORY}}

# Run a debug menmos-agent.
debug:
    MENMOS_AGENT_DEBUG=true go run ./cmd/menmos-agent

# Run a production menmos-agent.
prod:
    go run ./cmd/menmos-agent
