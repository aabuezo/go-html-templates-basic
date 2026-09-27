# Testing y QA

La calidad se revisa en varias capas. Un test unitario no reemplaza una prueba de navegador, y una prueba manual no reemplaza una regresión automatizada.

## Capas de pruebas

### Unitarias

Usarlas para validadores, reglas de dominio, transformaciones de datos, autorización y funciones puras.

### Integración

Usarlas para handlers y routing, persistencia, sesiones, middlewares y respuestas HTML, fragmentos y redirects. Verificar status, headers, cookies, cuerpo y efectos persistidos.

### End-to-end o navegador

Usarlas para navegación, formularios, login y logout, permisos, interacciones HTMX, responsive y accesibilidad básica.

## Matriz de QA para una funcionalidad

```text
Funcionalidad: [nombre]
Actor: [tipo de usuario]
Ruta/acción: [método y URL]
Resultado válido: [status y efecto]
Errores esperados: [status y mensaje]
Permiso requerido: [permiso]
```

Casos mínimos:

- datos válidos;
- campos obligatorios vacíos;
- límites inferior y superior;
- formato inválido;
- combinación inválida entre campos;
- usuario sin autenticación;
- usuario sin permiso;
- recurso fuera de alcance;
- request sin CSRF cuando corresponda;
- refresh después de un POST;
- doble envío.

## Formulario

Comprobar que el navegador muestra errores útiles, el servidor rechaza valores inválidos aunque se desactive JavaScript, los valores no sensibles se conservan, los secretos no vuelven al HTML, el foco llega al error y el status y redirect son los esperados.

## Sesiones y autorización

Comprobar acceso sin sesión, sesión válida, sesión expirada, logout, cambio de usuario en el mismo navegador, usuario con permiso de lectura pero no de escritura, recurso de otro tenant o propietario y regeneración de sesión.

## Comandos

Documentar en cada proyecto los comandos reales. Como base:

```bash
go test ./...
go vet ./...
go test -race ./...
```

Si existe frontend:

```bash
npm ci
npm run build
npm run lint
```

No documentar comandos que el repositorio no pueda ejecutar sin aclararlo.

## Evidencia para QA

Para un bug o una prueba fallida registrar versión o commit, entorno y navegador, pasos reproducibles, resultado esperado, resultado actual, status, URL, request relevante y logs sin secretos.
