#!/bin/bash

# Start the bootstrap process.
go install .
apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"
