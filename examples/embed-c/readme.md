# Example: Invoking embedlib from c

## Introduction

This example invokes the embedlib from C.

## Commands

### Starting darlean runtime with test actor

For this example, it is necessary to have:
* A Darlean runtime running, which provides the message bus and actor registry
* An example typescript actor registered to the cluster.

Both functionalities are provided by the `runner` example.

To start the darlean runtime with the test actor, use:

```
$ ./scripts/run-runner-deno
```

Note: For this to work, it is required to have `deno` installed.

### Building embedlib go dll

To build the embedlib go dll, use the following command. This also copies the built dll and header file into this folder.

```
$ .\scripts\build-embedlib
```

### Building the c application

To build the c application, use:

```
$ .\scripts\build-examples-embed-c
```

### Running the c application

To run the C application, use:

```
$ .\scripts\run-examples-embed-c
```

It connects to Darlean, invokes the typescript test actor's `echo` action with `Moon` as argument,
and it should receive the lowercase version (`moon`) as an answer.

```
Invoking: Moon
Invoked: moon
```