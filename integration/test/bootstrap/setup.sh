#!/bin/bash

cd project/
go install .

# Start the bootstrap process.
apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"
