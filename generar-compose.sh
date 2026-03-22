#!/bin/bash

NOMBRE_SCRIPT=$0
ARCHIVO_SALIDA=$1
CANTIDAD_CLIENTES=$2

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Uso correcto: $NOMBRE_SCRIPT <archivo_salida> <cantidad_clientes>"
    exit 1
fi

if ! [[ "$CANTIDAD_CLIENTES" =~ ^[0-9]+$ ]]; then
  echo "La cantidad de clientes debe ser un número"
  exit 1
fi

python3 generador-clientes.py "$ARCHIVO_SALIDA" "$CANTIDAD_CLIENTES"