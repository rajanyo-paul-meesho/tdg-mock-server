# discovery-proto

This repository contains Protocol Buffer (`.proto`) definitions for the `pdp-iop` gRPC services.

---

## Prerequisites / Dependencies

Before generating the Go code from proto files, ensure the following tools are installed and available in your `PATH`:

- **protoc** (Protocol Buffer compiler)  
  Recommended version: v5.28.2  
  Installation:
    - Download from https://github.com/protocolbuffers/protobuf/releases
    - Extract and add the `bin` folder to your `PATH`

- **protoc-gen-go** (Go plugin for protoc)  
  Version: v1.35.2  
  Install with:  
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2

markdown
Copy
Edit

- **protoc-gen-go-grpc** (Go gRPC plugin for protoc)  
  Install with:  
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

yaml
Copy
Edit

Make sure your `GOPATH/bin` is in your system `PATH` so that `protoc` can find these plugins.

---

## Generating Go code from proto files

Run the following command **inside the `discovery-proto` directory**:

protoc
--proto_path=.
--go_out=../pdp-iop/client/grpc/pdp-iop/v1
--go_opt=paths=source_relative
--go-grpc_out=../pdp-iop/client/grpc/pdp-iop/v1
--go-grpc_opt=paths=source_relative
pdp-iop-grpc/v1/*.proto

yaml
Copy
Edit

- `--proto_path=.` tells `protoc` to look for imports relative to the current directory.
- The generated Go files will be placed in `../pdp-iop/client/grpc/pdp-iop/v1`, preserving relative paths.
- The `paths=source_relative` option keeps generated files alongside their source proto paths.

---

## Notes

- Do **not** modify import paths or `go_package` options inside the `.proto` files; this command works with the existing proto definitions.
- Ensure the relative paths in output flags align with your project structure.
- If you get errors like `protoc-gen-go-grpc: program not found`, verify that the plugin binaries are installed and your `PATH` includes `$GOPATH/bin`.

---

Feel free to open issues or ask for help integrating this into your build or CI pipelines.
