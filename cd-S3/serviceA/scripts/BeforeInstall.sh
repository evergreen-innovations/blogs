#!/bin/bash
DIRECTORY="/opt/servicea"
if [ ! -d "$DIRECTORY" ]; then
   sudo mkdir "$DIRECTORY"
fi
