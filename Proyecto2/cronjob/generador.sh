#!/bin/bash

COMANDOS=(
    "docker run -d alpine awk 'BEGIN {for(i=0;i<5000000;i++) a[i]=\"SO1_PROYECTO2_RAM_TEST_STRING_FILL\"; system(\"sleep 240\")}'"
    "docker run -d alpine sh -c \"while true; do echo '2^20' | bc > /dev/null; sleep 2; done\""
    "docker run -d alpine sleep 240"
)

echo "Generando 5 contenedores aleatorios..."

for i in {1..5}
do
    RANDOM_INDEX=$((RANDOM % 3))
    echo "Lanzando contenedor tipo $((RANDOM_INDEX + 1))..."
    eval ${COMANDOS[$RANDOM_INDEX]}
done

echo "Generación completada."
