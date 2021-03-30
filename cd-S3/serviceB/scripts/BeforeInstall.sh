#!/bin/bash
DIRECTORY="/opt/serviceb"
if [ ! -d "$DIRECTORY" ]; then
   sudo mkdir "$DIRECTORY"
fi
