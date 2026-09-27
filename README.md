# Go web learning project

Este repositorio es un proyecto de aprendizaje sobre aplicaciones web server-rendered con Go. El código actual muestra routing, middlewares, templates, formularios, validación, sesiones y protección CSRF mediante un ejemplo pequeño.

No es una aplicación de producción. Las decisiones y checklists reutilizables están documentadas en `docs/`; las instrucciones para mantener esa documentación están en [AGENTS.md](AGENTS.md).

## Ejecutar el proyecto actual

Requisito: una versión de Go compatible con `go.mod`.

```bash
go run ./cmd/web
```

La aplicación escucha en `http://localhost:8080`.

Para comprobar que el código compila:

```bash
go test ./...
go vet ./...
```

Rutas actuales del ejemplo:

- `GET /`: página pública.
- `GET /about`: otra página pública.
- `GET /login` y `POST /login`: formulario educativo con validación.
- `GET /user-home`: página que demuestra el uso de datos de sesión.

## Documentación reutilizable

La guía está pensada para poder acompañar la creación de un proyecto nuevo, no solamente para explicar este repositorio.

- [Guía completa para iniciar un proyecto web en Go](docs/go-web-guide.md)
- [Formularios y validación](docs/forms-and-validation.md)
- [Sesiones, autenticación y autorización](docs/sessions-auth-authorization.md)
- [Templates, HTMX, Tailwind, Alpine.js y Vite](docs/frontend.md)
- [Testing y QA](docs/testing-and-qa.md)
- [Checklist de seguridad para producción](docs/production-security.md)

## Cómo usar la guía en un proyecto nuevo

1. Leer la guía general y definir el dominio, los usuarios y los flujos principales.
2. Crear el proyecto y tomar decisiones explícitas sobre templates, assets, sesiones y persistencia.
3. Seguir las checklists de formularios, frontend, testing y seguridad mientras se implementa cada funcionalidad.
4. Crear en el nuevo proyecto un README específico con sus rutas, comandos, variables de entorno, dependencias y decisiones particulares.

## Alcance

La guía prioriza una arquitectura server-rendered y progresiva: Go es la fuente de verdad, HTML es la interfaz principal, HTMX sirve para actualizaciones parciales, Alpine.js para estado local pequeño y Tailwind/Vite para los assets. Las herramientas pueden reemplazarse si los requisitos del proyecto lo justifican.
