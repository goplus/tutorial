#!/bin/bash

go build -o tutorial . && mkdir -p dist && cp -r * dist
