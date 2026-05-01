FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copiamos el archivo que acabamos de mover
COPY api-rust-bin /usr/local/bin/api-rust

EXPOSE 8080
# Le damos permisos de ejecución por si acaso
RUN chmod +x /usr/local/bin/api-rust

CMD ["api-rust"]