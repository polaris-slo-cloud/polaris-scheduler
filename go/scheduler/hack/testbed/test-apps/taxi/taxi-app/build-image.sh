#!/usr/bin/env bash
#!/bin/bash

set -e

#docker build -t edgify:0.1.0 ./base
docker build -t rainbowh2020/taxi-async:0.0.1 ./taxi_exp
