#!/bin/sh

if [ "$SERVER_PORT" = "8000" ];
then
    echo "Starting the agentd port 8000"
   /usr/local/tailtrace/agentd   -port 8000 -rpcport 50000 -filename trace1.data 2>&1 
elif [ "$SERVER_PORT" = "8001" ];  
then
    echo "Starting the agentd port 8001"
   /usr/local/tailtrace/agentd  -port 8001 -rpcport 50001  -filename trace2.data 2>&1 
elif  [ "$SERVER_PORT" = "8002" ];  
then
    echo "Starting the processd port 8002"
   /usr/local/tailtrace/processd -port 8002 -rpcport 50002   2>&1 
fi
tail -f $1
