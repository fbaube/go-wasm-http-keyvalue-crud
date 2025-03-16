# Works! :-D

This repo makes fixes to WasmCloud's
[example](https://github.com/wasmCloud/go/tree/main/examples/component/http-keyvalue-crud)
( described
[here](https://wasmcloud.com/blog/2025-01-23-walkthrough-a-wasmclod-crud-application-in-go/)
) so that it actually works.

This "early" version works with `wasm-tools` version 1.225.0. It does NOT work
with `wasm-tools` version 1.227.0, and the cause looks impossible to sort out.

It remains to be seen whether it works with `wash up -d`.

The router has to be passed to `wasi-http`, which chokes on the
new Go `ServeMux`, so we use a third-party router instead.

To exercise this example when running `wash dev`:
```
wash app list
curl -X POST localhost:8000/crud/mario -d '{"itsa": "me", "woo": "hoo"}'
curl localhost:8000/crud/mario
curl -X DELETE localhost:8000/crud/mario
```

# Go HTTP Key-Value CRUD

[This example](https://github.com/wasmCloud/go/tree/main/examples/component/http-keyvalue-crud)
is a WebAssembly component that demonstrates simple CRUD operations (Create, Read, Update, Destroy) with the
[`wasi:keyvalue/store`](https://github.com/WebAssembly/wasi-keyvalue) interface. 

## 📦 Dependencies

> [!WARNING]
> Due to incompatibilities introduced in `wasm-tools` v1.226.0, a version 
> of `wasm-tools` <= 1.225.0 is **required** for running this example.
>
> You can install `wasm-tools` [v1.225.0 from upstream releases](https://github.com/bytecodealliance/wasm-tools/releases/tag/v1.225.0), or use
> `cargo` ([Rust toolchain](https://doc.rust-lang.org/cargo/getting-started/installation.html)) -- (i.e. `cargo install --locked wasm-tools@1.225.0`)

Before starting, you need to have installed
[`tinygo`](https://tinygo.org/getting-started/install/),
[`wasm-tools`](https://github.com/bytecodealliance/wasm-tools#installation),
wasmCloud Shell [`wash`](https://wasmcloud.com/docs/installation).

## 👟 Run the example

In addition to the standard elements of a Go project, the
directory includes the following files and directories:

- `/build`: Target directory for compiled `.wasm` binaries
- `/gen`: Target directory for Go bindings of
[interfaces](https://wasmcloud.com/docs/concepts/interfaces)
- `/wit`: Directory for WIT packages that define interfaces
- `bindings.wadge.go`: Automatically generated test bindings
- `wadm.yaml`: Declarative app manifest
- `wasmcloud.lock`: Automatically generated lockfile for WIT packages
- `wasmcloud.toml`: Configuration file for a wasmCloud app

### Start a local development loop

Run `wash dev` to start a local development loop:

```shell
wash dev
```

The `wash dev` command will:

- Start a local wasmCloud environment
- Build this component
- Deploy your app and all dependencies to run the app locally
- Watch your code for changes and re-deploy when necessary.

### Clean up

You can stop the `wash dev` process with `Ctrl-C`.

## 📖 Further reading

When running this example with the `wash dev` command, wasmCloud uses
its included NATS key-value store to back key-value operations, but the
app could use another store like Redis with no change to the Go code. 

Learn more about capabilities like key-value storage are fulfilled
by swappable providers in the
[wasmCloud Quickstart](https://wasmcloud.com/docs/tour/hello-world).  
