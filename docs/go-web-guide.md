# Guía general para iniciar una aplicación web en Go

Esta guía describe un recorrido reutilizable para crear aplicaciones web server-rendered con Go. No depende de un dominio, una empresa, una ruta o un modelo concreto.

## 1. Definir el problema antes de escribir código

Documentar primero:

- qué usuarios existen;
- qué recursos administra el sistema;
- qué acciones puede realizar cada tipo de usuario;
- qué datos son obligatorios, sensibles o auditables;
- qué flujos necesitan una página completa y cuáles necesitan una actualización parcial;
- qué requisitos de disponibilidad, privacidad y rendimiento existen.

Para cada funcionalidad conviene definir:

```text
Actor → acción → recurso → resultado esperado → errores posibles
```

Esto evita diseñar handlers y tablas antes de entender el flujo.

## 2. Estructura inicial

Una estructura razonable para un servicio web es:

```text
.
├── cmd/
│   └── web/
│       ├── main.go
│       ├── middleware.go
│       └── routes.go
├── internal/
│   ├── config/
│   ├── forms/
│   ├── handlers/
│   ├── models/
│   ├── render/
│   ├── session/
│   └── services/
├── templates/
├── static/
├── migrations/
├── tests/
├── go.mod
└── README.md
```

No es una obligación. La regla importante es que el punto de entrada componga las dependencias y que los handlers no creen conexiones, configuraciones globales ni servicios por su cuenta.

`internal/` ayuda a evitar que los paquetes de la aplicación se importen desde otros módulos. `cmd/` contiene ejecutables; la lógica reutilizable vive en paquetes internos.

## 3. El composition root

`cmd/web/main.go` debería crear y conectar:

1. configuración;
2. logger;
3. conexiones a base de datos y cache;
4. repositorios;
5. servicios de dominio;
6. gestor de sesiones;
7. renderer;
8. handlers;
9. router y middlewares;
10. servidor HTTP y shutdown ordenado.

Ejemplo conceptual:

```go
repository := repository.New(db)
service := service.New(repository)
renderer := render.New(templateCache)
handler := handlers.New(service, renderer, sessionManager)
router := routes.New(handler, sessionManager)
```

Las interfaces deben introducirse cuando ayudan a probar o reemplazar una dependencia, no como requisito mecánico para cada struct.

## 4. Flujo de una request

Un flujo típico es:

```text
Request
  ↓
Router
  ↓
Recovery, logging, request ID
  ↓
CSRF, carga de sesión, autenticación
  ↓
Handler
  ↓
Servicio de dominio
  ↓
Repositorio o API externa
  ↓
Respuesta HTML, fragmento HTML, JSON o redirect
```

El handler traduce HTTP. Las reglas de negocio deberían vivir en servicios o funciones de dominio, y el acceso a datos en repositorios o adaptadores.

## 5. Elegir el tipo de respuesta

- Página completa: navegación normal o primera carga.
- Fragmento HTML: interacción con HTMX.
- JSON: integración con una API, un gráfico u otro cliente que realmente necesite datos estructurados.
- Redirect: después de un POST exitoso o cuando cambia el recurso de destino.

Usar `303 See Other` después de un POST exitoso suele evitar reenvíos accidentales al refrescar la página. No todos los formularios necesitan redirect: si el usuario debe continuar cargando datos, se puede renderizar la misma página con un mensaje de éxito y estado limpio.

## 6. Templates y datos de vista

Separar:

- modelo de persistencia;
- input del formulario;
- datos preparados para la vista.

Un `TemplateData` común puede transportar mensajes, token CSRF y el formulario actual, mientras que `Data` o una struct específica contiene los datos de la página. No conviene pasar directamente entidades de base de datos si la vista no necesita todos sus campos.

Con `html/template`, los valores se escapan como HTML por defecto. El uso de HTML confiable debe ser excepcional y revisado manualmente.

También se puede elegir Templ. Sus archivos `.templ` se convierten en código Go generado y sus componentes pueden componerse como funciones. La decisión no cambia el resto de la arquitectura: siguen existiendo handlers, datos de vista, validación, sesiones y respuestas HTTP.

## 7. Orden recomendado de implementación

Para cada recurso o flujo:

1. Definir casos de uso y permisos.
2. Crear el modelo de entrada y las validaciones.
3. Implementar el servicio de dominio.
4. Implementar persistencia o el adaptador externo.
5. Crear el handler GET.
6. Crear la vista y sus estados vacío, cargando, éxito y error.
7. Crear el handler POST/PUT/PATCH/DELETE.
8. Añadir CSRF, autenticación y autorización.
9. Añadir soporte HTMX si una actualización parcial mejora la experiencia.
10. Agregar tests y casos de QA.
11. Revisar logs, límites, seguridad y observabilidad.

## 8. Configuración

La configuración debe venir del entorno o de un proveedor de secretos, no de valores sensibles hardcodeados. Validar al iniciar:

- dirección de escucha;
- modo de ejecución;
- URL y credenciales de base de datos;
- clave de cifrado o autenticación de sesiones;
- configuración de cookies;
- endpoints externos;
- límites relevantes.

Separar configuración de desarrollo, test y producción. Los defaults locales pueden ser cómodos, pero una configuración de producción incompleta debería impedir el arranque.

## 9. Qué documentar por proyecto

Cada proyecto nuevo debería agregar un README propio con:

- propósito y alcance;
- cómo ejecutar la aplicación;
- dependencias externas;
- variables de entorno;
- comandos de test, lint y build;
- estructura de carpetas;
- tabla de rutas;
- modelo de autorización;
- migraciones y datos de desarrollo;
- decisiones que difieren de esta guía.

La guía general explica el proceso. El README del proyecto explica la realidad concreta.
