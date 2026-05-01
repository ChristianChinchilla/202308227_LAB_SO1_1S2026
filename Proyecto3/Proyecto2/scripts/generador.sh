#!/bin/bash
#generar 3 contenedores de bajo consumo
for i in {1..3}; do
    docker run -d alpine sleep 3600
done

#generar 2 contenedores de alto consumo
for i in {1..2}; do
    docker run -d alpine sh -c "while true; do :; done"
done
