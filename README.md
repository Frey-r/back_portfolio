# Backend Go — Portfolio Eduardo Bachmann

Este es el backend para el portafolio de Eduardo Bachmann. Desarrollado en Go 1.24, utilizando únicamente la librería estándar para routing (`net/http`) y SQLite como base de datos embebida.

## Funcionalidades
- **Contacto**: API para recibir, validar y guardar mensajes del formulario de contacto.
- **Rate Limiting**: Mitigación en memoria por IP para evitar el abuso del endpoint de contacto.
- **LinkedIn**: Cacheo concurrente y rescate del post principal de LinkedIn (soporte para fallback sin autenticación real).
- **Graceful Shutdown**: Cierre seguro para asegurar que las transacciones y requests finalizan adecuadamente.
- **Cross-Origin Resource Sharing (CORS)**: Configuraciones optimizadas para operar con Astro.

## Iniciar (Local)

Existen opciones automatizadas en el `Makefile`:

```bash
# Instalar dependencias
go mod tidy

# Iniciar servidor local
make dev
```

El servidor quedará montado por defecto en `http://localhost:8080`.

## Configuración y Entorno

Sobreescribe los valores modificando el archivo `.env`. (puedes basarte en `.env.example`).

```env
PORT=8080
ALLOWED_ORIGINS=http://localhost:4321
SQLITE_PATH=./data/portfolio.db
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/health` | Estado del backend. |
| `GET`  | `/api/status` | Info de disponibilidad / estado del perfil. |
| `GET`  | `/api/linkedin/top-post` | Fetch post de LinkedIn desde el cache interno. |
| `POST` | `/api/contact` | Guardar un nuevo mensaje. |
