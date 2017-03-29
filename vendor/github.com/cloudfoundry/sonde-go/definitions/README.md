# dropsonde-protocol

Dropsonde is a two-way protocol for emitting events and metrics in one direction. Messages are encoded in the Google [Protocol Buffer](https://developers.google.com/protocol-buffers) binary wire format.

It is a goal of the system to reduce the need for low-level metric events (e.g. `ValueMetric` and `CounterEvent` messages). Though in the early stages, we include types such as `HttpStartEvent`, `HttpStopEvent` and `HttpStartStopEvent` to allow metric generation and aggregation at a higher level of abstraction, and to offload the work of aggregation to downstream receivers. Emitting applications should focus on delivering events, not processing them or computing statistics.

This protocol forms the backbone of the [Doppler](https://github.com/cloudfoundry/loggregator) system of Cloud Foundry.

## Message types

Please see the following for detailed descriptions of each type:

* [events README](events/README.md)

## Library using this protocol

* [Sonde-Go](https://github.com/cloudfoundry/sonde-go) is a generated Go library for components that wish to emit messages to or consume messages from the Cloud Foundry [metric system](https://github.com/cloudfoundry/loggregator).

## Generating code

Note: Due to [maps not being supported in protoc v2.X](https://github.com/google/protobuf/issues/799#issuecomment-138207911), the proto definitions in this repository require protoc v3.0.0 or higher.

### Go

Code generation for Go has moved to the [Sonde-Go](https://github.com/cloudfoundry/sonde-go) library.

### Java

1. Build the go code first (see above) so that all the imports are available

2. Generate the Java code (optionally providing a target path as a directory)
   ```
   ./generate-java.sh [TARGET_PATH]
   ```

### Other languages

For C++ and Python, Google provides [tutorials](https://developers.google.com/protocol-buffers/docs/tutorials).

Please see [this list](https://github.com/google/protobuf/wiki/Third-Party-Add-ons#Programming_Languages) for working with protocol buffers in other languages.

### Message documentation

Each package's documentation is auto-generated with [protoc-gen-doc](https://github.com/estan/protoc-gen-doc). After installing the tool, run:
```
cd events
protoc --doc_out=markdown,README.md:. *.proto
```

## Communication protocols

### Event emission

Dropsonde is intended to be a "fire and forget" protocol, in the sense that an emitter should send events to its receiver with no expectation of acknowledgement. There is no "handshake" step; the emitter simply begins emitting to a known address of an expected recipient.
