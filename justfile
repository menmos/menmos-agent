OUTPUT_DIRECTORY := "build/"

# Build the workspace.
build:
    mkdir -p {{OUTPUT_DIRECTORY}}
    go build -o {{OUTPUT_DIRECTORY}} ./cmd/menmos-agent

# Delete the build directory.
clean:
    rm -rf {{OUTPUT_DIRECTORY}}

# Run a debug menmos agent locally.
debug:
    MENMOS_AGENT_DEBUG=true go run ./cmd/menmos-agent

prod:
    go run ./cmd/menmos-agent
