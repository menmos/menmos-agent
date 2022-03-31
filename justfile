OUTPUT_DIRECTORY := "build/"

# Build the workspace.
build:
    mkdir -p {{OUTPUT_DIRECTORY}}
    go build -o {{OUTPUT_DIRECTORY}} ./cmd/menmos-agent

# Delete the build directory.
clean:
    rm -rf {{OUTPUT_DIRECTORY}}

# Run a menmos agent locally.
menmos-agent:
    go run ./cmd/menmos-agent
