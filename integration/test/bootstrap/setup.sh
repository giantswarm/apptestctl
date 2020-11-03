#!/bin/bash

# Start the bootstrap process.
make
apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"
