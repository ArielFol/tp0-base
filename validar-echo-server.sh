#!/bin/bash

RED="tp0_testing_net"
SERVER="server"
PORT="12345"
MENSAJE="prueba"

RESPUESTA=$(docker run --rm --net $RED alpine sh -c "
            echo -n '$MENSAJE' | nc $SERVER $PORT
            ")
if [ "$RESPUESTA" = "$MENSAJE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
