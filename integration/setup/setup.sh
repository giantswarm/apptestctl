#!/bin/bash

cd project/
go install .
apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"