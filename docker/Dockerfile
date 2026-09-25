# ==============================================================================
# IPVN7 Network OS - Zero-Privilege Scratch Multi-Stage Container
# ==============================================================================

# Etapa 1: Compilador estático reproducible
FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates

# Copiar manifiestos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente completo
COPY . .

# Compilación estática Pure Go sin CGO ni dependencias dinámicas
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /build/ipvn7 ./cmd/ipvn7

# Etapa 2: Imagen mínima sin privilegios (Superficie de ataque cero)
FROM scratch

# Certificados raíz para integraciones salientes
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binario estático compilado
COPY --from=builder /build/ipvn7 /ipvn7

# Archivos estáticos de frontend si se sirven localmente
COPY web/ /web/

# Volúmenes para persistencia de claves soberanas y base de datos local
VOLUME ["/keystore", "/data"]

# Puertos estándar IPVN7:
# 7777/udp   - Transporte físico de malla P2P (Kleinberg)
# 7070/tcp   - API Smart Component Gateway y Dashboard
# 10807/tcp  - Proxy SOCKS5 en espacio de usuario
# 10808/tcp  - Proxy HTTP CONNECT en espacio de usuario
EXPOSE 7777/udp 7070/tcp 10807/tcp 10808/tcp

# Usuario sin privilegios de sistema
USER 10007:10007

ENTRYPOINT ["/ipvn7"]
CMD ["-port", "7777", "-web-port", "7070", "-keystore", "/keystore/node_identity.key"]
