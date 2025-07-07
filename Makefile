# Build and copy the Go binary, then build the Docker image for matlab
.PHONY: build-matlab-docker
# Allow MATLAB_RELEASE to be overridden from the command line, default to R2024b
MATLAB_RELEASE ?= R2024b
build-matlab-docker:
	$(MAKE) -C batch-mcp-main build-linux-amd64
	$(MAKE) prepare-matlab-binary
	docker buildx bake -f matlab/docker-bake.hcl matlab --set matlab.args.MATLAB_RELEASE=$(MATLAB_RELEASE) --set matlab.args.MATLAB_DEPS_OS=ubuntu22.04
# Makefile to prepare build context for matlab Docker build from project root

.PHONY: prepare-matlab-binary clean-matlab-binary

# Path to the built Go binary
BINARY_PATH=batch-mcp-main/build/batch-mcp-linux-amd64
# Path to where the binary should be copied for Docker build
TARGET_PATH=matlab/build/batch-mcp-linux-amd64

prepare-matlab-binary:
	@mkdir -p matlab/build
	cp $(BINARY_PATH) $(TARGET_PATH)
	@echo "Copied $(BINARY_PATH) to $(TARGET_PATH) for Docker build context."

clean-matlab-binary:
	rm -f matlab/build/batch-mcp-linux-amd64
	@echo "Removed copied Go binary from matlab/build context."
