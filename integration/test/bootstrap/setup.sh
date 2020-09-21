#!/bin/bash

cd project/
make

# Start the bootstrap process.
./apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"
