#!/bin/bash
COMMAND="/opt/servicec"
PIDCMD="pgrep -f "$COMMAND"" 

if [[ $($PIDCMD) ]]
then
     pgrep -f "$COMMAND" | xargs kill $1
else
     echo "NO PID"
fi