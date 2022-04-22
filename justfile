OUTPUT_DIRECTORY := "build/"

_build_target target:
    @mkdir -p {{OUTPUT_DIRECTORY}}
    go build -o {{OUTPUT_DIRECTORY}} ./cmd/{{target}}

build-agent:
    @just _build_target

# Build the workspace.
build:
    @just _build_target menmos-agent
    @just _build_target menmos-platform-cli

# Test the workspace
test:
    go test -cover -race ./...

# Delete the build directory.
clean:
    rm -rf {{OUTPUT_DIRECTORY}}

# Run a debug menmos-agent.
debug:
    MENMOS_AGENT_DEBUG=true go run ./cmd/menmos-agent

# Run a production menmos-agent.
prod:
    go run ./cmd/menmos-agent
