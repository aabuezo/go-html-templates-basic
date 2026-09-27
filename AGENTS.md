# Instrucciones para mantener la documentación

Este repositorio contiene material de aprendizaje, pero la documentación principal debe funcionar como una guía reutilizable para construir aplicaciones web con Go.

## Objetivo de la documentación

La documentación debe ayudar a una persona a iniciar un proyecto nuevo y avanzar paso a paso usando:

- Go como backend;
- `html/template` o un sistema equivalente de templates server-rendered;
- formularios y validación server-side;
- sesiones, autenticación y autorización;
- HTMX para actualizaciones parciales de HTML;
- Tailwind CSS para estilos;
- Alpine.js para interacciones pequeñas del navegador;
- Vite cuando sea conveniente para compilar CSS y JavaScript.

No debe convertirse en una descripción detallada de este repositorio de aprendizaje.

## Reglas para futuras actualizaciones

1. Leer este archivo antes de modificar README.md o cualquier documento dentro de `docs/`.
2. Mantener las ideas, decisiones y checklists generales. Evitar nombres de rutas, modelos, usuarios, mensajes y datos propios de esta aplicación.
3. Cuando se necesite un ejemplo, usar nombres genéricos como `Resource`, `CreateInput`, `/resources` o `example.com`.
4. Separar siempre:
   - lo conceptual;
   - el flujo recomendado;
   - el ejemplo de código;
   - el checklist de verificación;
   - las consideraciones de producción.
5. Marcar claramente si algo es:
   - `Implementado en este repositorio`;
   - `Ejemplo conceptual`;
   - `Decisión recomendada`;
   - `Pendiente de incorporar`.
6. No presentar como activo algo que solamente esté propuesto en la guía.
7. Mantener el README corto y específico del repositorio. La guía general debe vivir en `docs/`.
8. Si una herramienta o dependencia cambia, actualizar la explicación y revisar los comandos relacionados.
9. No documentar secretos, credenciales reales ni configuraciones inseguras como si fueran adecuadas para producción.
10. Después de editar documentación, comprobar enlaces internos, comandos, nombres de carpetas y referencias a `pkg`/`internal` contra el código actual.

## Estructura esperada

```text
README.md                         # contexto y arranque de este repositorio
AGENTS.md                         # estas reglas de mantenimiento
docs/
├── go-web-guide.md               # recorrido completo para un proyecto nuevo
├── forms-and-validation.md      # formularios y validación
├── sessions-auth-authorization.md
├── frontend.md                   # templates, HTMX, Tailwind, Alpine y Vite
├── testing-and-qa.md             # tests y pruebas manuales
└── production-security.md       # checklist de producción
```

## Estilo

Escribir en español claro, con lenguaje directo. Explicar el porqué de cada decisión, no solamente los comandos. Preferir ejemplos pequeños y genéricos. Evitar repetir la misma explicación en varios documentos; enlazar al documento especializado cuando corresponda.
