# Arquitectura de la aplicación

Este proyecto es una aplicación web server-rendered escrita en Go. Go genera el HTML con `html/template`, Chi enruta las requests, los middlewares aplican comportamiento común y los paquetes de `pkg` reciben sus dependencias desde el punto de entrada.

La aplicación está en una etapa inicial: actualmente tiene las páginas Home y About, sesiones, protección CSRF y templates con Bootstrap cargado desde CDN. La guía de frontend, incluyendo HTMX, Tailwind, Alpine.js, TypeScript y ECharts, está documentada en [HTMX.md](HTMX.md) como camino de evolución; esas herramientas todavía no forman parte de las dependencias instaladas en este repositorio.

## Índice

- [Inicio rápido](#inicio-rápido)
- [Decisiones arquitectónicas](#decisiones-arquitectónicas)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Composición de la aplicación](#composición-de-la-aplicación)
- [Flujo de una request](#flujo-de-una-request)
- [Responsabilidad de cada paquete](#responsabilidad-de-cada-paquete)
- [Templates y renderizado](#templates-y-renderizado)
- [Configuración y cache](#configuración-y-cache)
- [Sesiones](#sesiones)
- [CSRF y formularios](#csrf-y-formularios)
- [Agregar una nueva ruta](#agregar-una-nueva-ruta)
- [Dependencias](#dependencias)
- [Estado actual y próximos pasos](#estado-actual-y-próximos-pasos)

## Inicio rápido

Requisitos: Go compatible con la versión declarada en `go.mod`.

```bash
go run ./cmd/web
```

La aplicación escucha en `http://localhost:8080`.

Para verificar que todos los paquetes compilan y que los tests pasan:

```bash
go test ./...
go vet ./...
```

Actualmente las rutas principales son:

- `GET /`: página Home; guarda la dirección remota en la sesión.
- `GET /about`: página About; muestra la dirección guardada en la sesión, si existe.

## Decisiones arquitectónicas

### Go genera HTML en el servidor

El servidor es la fuente de verdad para el contenido y las reglas de negocio. Los handlers preparan datos y `html/template` genera la respuesta HTML.

Esto mantiene el frontend simple y permite que una página funcione sin un framework SPA. Si más adelante una interacción necesita actualizar solo una parte de la página, HTMX puede solicitar un fragmento HTML sin cambiar esta arquitectura. Ver [HTMX.md](HTMX.md).

### `main` es el composition root

`cmd/web/main.go` crea y conecta las dependencias. Los paquetes no crean configuraciones globales ni buscan dependencias por su cuenta.

Ventajas:

- las dependencias son visibles;
- el ciclo de vida de cada componente está centralizado;
- los paquetes son más fáciles de probar;
- cambiar una implementación no obliga a modificar todos los consumidores.

### Inyección de dependencias, no variables globales

Los handlers reciben el renderer y la configuración a través de sus constructores:

```go
renderer := render.NewRenderer(app)
handler := handlers.NewHandler(app, renderer)
```

Una variable global ocultaría las dependencias y dificultaría los tests y el manejo de concurrencia.

### Separación entre HTTP, renderizado y sesiones

- `cmd/web` compone la aplicación, registra rutas y define middlewares.
- `pkg/handlers` traduce requests en respuestas y prepara datos.
- `pkg/render` carga y ejecuta templates.
- `pkg/session` configura el gestor de sesiones.
- `pkg/config` reúne dependencias compartidas.

Esta separación no intenta aplicar patrones por nombre. Por ejemplo, `pkg/session` no es un repository y `AppConfig` no es un `Repository`: un repository real debería abstraer persistencia, como una consulta a usuarios.

## Estructura del proyecto

```text
.
├── go.mod
├── go.sum
├── cmd/
│   └── web/
│       ├── main.go          # composición y arranque del servidor
│       ├── middleware.go    # CSRF y carga de sesión
│       └── routes.go        # router y orden de middlewares
├── pkg/
│   ├── config/
│   │   └── config.go        # AppConfig
│   ├── handlers/
│   │   └── handlers.go      # handlers Home y About
│   ├── models/
│   │   └── templateData.go  # datos comunes para templates
│   ├── render/
│   │   └── render.go        # cache y ejecución de templates
│   └── session/
│       └── sessions.go      # configuración de SCS
├── templates/
│   ├── about.page.tmpl
│   ├── base.layout.tmpl
│   └── home.page.tmpl
├── README.md
└── HTMX.md                 # guía de frontend y evolución propuesta
```

No hay todavía una carpeta `frontend/` ni una salida `static/` en el repositorio. Si se incorpora Vite, esa estructura y su flujo de build se explican en [HTMX.md](HTMX.md).

## Composición de la aplicación

El arranque actual sigue este orden:

```go
tc, err := render.CreateTemplateCache()
if err != nil {
    log.Fatal("cannot create template cache")
}

app := &config.AppConfig{
    UseCache:      false,
    TemplateCache: tc,
    InProduction:  false,
}

app.Session = session.New(app.InProduction)

renderer := render.NewRenderer(app)
handler := handlers.NewHandler(app, renderer)

srv := &http.Server{
    Addr:    ":8080",
    Handler: routes(handler, app),
}

log.Fatal(srv.ListenAndServe())
```

Los constructores pueden llamarse `NewRenderer` y `NewHandler` porque expresan con claridad qué tipo crean. `session.New` usa el nombre corto porque el paquete ya deja claro qué construye.

## Flujo de una request

```text
Request HTTP
    ↓
chi router
    ↓
Recoverer
    ↓
NoSurf
    ↓
SessionLoad
    ↓
Handler.Home o Handler.About
    ↓
TemplateData
    ↓
Renderer.RenderTemplate
    ↓
HTML generado con html/template
```

El orden está definido en `cmd/web/routes.go`:

```go
mux.Use(middleware.Recoverer)
mux.Use(NoSurf(app))
mux.Use(SessionLoad(app))
```

`Recoverer` evita que un panic cierre el proceso. `NoSurf` configura CSRF. `SessionLoad` carga la sesión antes del handler y guarda sus cambios al terminar.

## Responsabilidad de cada paquete

### `cmd/web`

Es el borde de la aplicación. Aquí se conoce el servidor HTTP concreto, el router Chi, las rutas y el orden de los middlewares.

### `pkg/config`

Define `AppConfig`, el objeto que contiene dependencias y opciones compartidas:

```go
type AppConfig struct {
    UseCache      bool
    TemplateCache map[string]*template.Template
    InfoLog       *log.Logger
    InProduction  bool
    Session       *scs.SessionManager
}
```

### `pkg/handlers`

Contiene los handlers HTTP. `Handler` conserva las dependencias que necesita cada request:

```go
type Handler struct {
    Renderer *render.Renderer
    App      *config.AppConfig
}
```

Un handler debería leer la request, ejecutar la lógica necesaria, construir los datos y delegar el HTML al renderer. Si la lógica de negocio crece, conviene extraerla a un servicio en vez de cargar más responsabilidades en el handler.

### `pkg/render`

Contiene `Renderer`, `CreateTemplateCache` y la ejecución de templates. El handler no necesita saber si el template salió de un cache o del disco.

### `pkg/models`

Contiene `TemplateData`, la estructura común que reciben los templates:

```go
type TemplateData struct {
    StringMap map[string]string
    IntMap    map[string]int
    FloatMap  map[string]float32
    Data      map[string]any
    CSRFToken string
    Flash     string
    Warning   string
    Error     string
}
```

`Data map[string]any` permite datos específicos de una página, pero sacrifica seguridad de tipos. Cuando una página gane complejidad, puede ser preferible definir un modelo específico.

### `pkg/session`

Centraliza la configuración de `github.com/alexedwards/scs/v2`. No conoce handlers ni templates; solo crea el gestor de sesiones.

## Templates y renderizado

Las páginas usan el sufijo `.page.tmpl` y el layout usa `.layout.tmpl`. `CreateTemplateCache` busca ambos grupos en `./templates`, combina cada página con los layouts y devuelve un `map[string]*template.Template`.

El layout actual (`templates/base.layout.tmpl`) usa Bootstrap 5.3.8 desde CDN y expone los bloques `css`, `content` y `js`. Las páginas `home.page.tmpl` y `about.page.tmpl` definen el bloque `content`.

Para agregar una página:

1. Crear `templates/nombre.page.tmpl`.
2. Incluir `{{ template "base" . }}`.
3. Definir `{{ define "content" }}`.
4. Agregar un método al handler.
5. Registrar la ruta en `cmd/web/routes.go`.

Ejemplo mínimo:

```gotemplate
{{ template "base" . }}

{{ define "content" }}
    <h1>Configuración</h1>
{{ end }}
```

Los valores de `html/template` se escapan automáticamente. No conviertas strings de usuarios a `template.HTML` salvo que el contenido sea controlado y entiendas el riesgo.

## Configuración y cache

`InProduction` y `UseCache` representan decisiones distintas:

- `InProduction` configura opciones propias de producción, como cookies `Secure`.
- `UseCache` decide si el renderer usa `TemplateCache` o vuelve a crear los templates.

No conviene reemplazar una variable por la otra. En desarrollo puede ser útil desactivar el cache para ver cambios sin reiniciar; en producción conviene crear el cache una vez y reutilizarlo.

El estado actual usa:

```go
UseCache:     false,
InProduction: false,
```

Eso es cómodo para aprender, pero no es una configuración adecuada para producción. Antes de desplegar, la aplicación debería leer estas opciones desde variables de entorno o una configuración explícita y manejar errores de renderizado sin terminar el proceso desde una request.

## Sesiones

La aplicación usa `github.com/alexedwards/scs/v2`:

```go
func New(inProduction bool) *scs.SessionManager {
    manager := scs.New()
    manager.Lifetime = 24 * time.Hour
    manager.Cookie.Persist = true
    manager.Cookie.SameSite = http.SameSiteLaxMode
    manager.Cookie.Secure = inProduction
    return manager
}
```

Configuración actual:

- duración máxima de 24 horas;
- cookie persistente;
- `SameSite=Lax`;
- `Secure` activado cuando `InProduction` es `true`;
- store en memoria, útil localmente pero no para varias instancias o reinicios de producción.

El ejemplo actual guarda `r.RemoteAddr` en `/` y lo recupera en `/about`. `RemoteAddr` puede pertenecer a un proxy en producción; no se debe confiar ciegamente en `X-Forwarded-For` sin configurar una lista de proxies confiables.

Para producción, reemplazar el store en memoria por uno compartido, como Redis o una base de datos.

## CSRF y formularios

`NoSurf(app)` protege las requests que modifican estado. Cuando se agregue un formulario, el handler debe incluir el token:

```go
data := &models.TemplateData{
    CSRFToken: nosurf.Token(r),
}
```

Y el template:

```gotemplate
<form method="post" action="/perfil">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    <!-- campos -->
</form>
```

Las rutas `GET` deben ser de lectura. Las operaciones `POST`, `PUT`, `PATCH` y `DELETE` deben validar autenticación, autorización, datos de entrada y CSRF cuando corresponda.

Actualmente no existen formularios de escritura: el middleware está conectado y el modelo tiene `CSRFToken`, pero todavía no hay un flujo que lo utilice.

## Agregar una nueva ruta

Para agregar una página o endpoint:

1. Definir el método en `pkg/handlers`.
2. Preparar un `TemplateData` o una respuesta específica.
3. Crear el template si la respuesta es HTML.
4. Registrar el método en `cmd/web/routes.go`.
5. Si modifica datos, usar el método HTTP adecuado y CSRF.
6. Agregar un test para la respuesta y los casos de error.

Ejemplo de ruta HTML:

```go
func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
    h.Renderer.RenderTemplate(w, "contact.page.tmpl", &models.TemplateData{})
}
```

```go
mux.Get("/contact", handler.Contact)
```

No agregues lógica de negocio, acceso a base de datos o dependencias globales directamente en `routes.go`.

## Dependencias

Las dependencias actuales de `go.mod` son:

| Dependencia | Uso |
| --- | --- |
| `github.com/go-chi/chi/v5` | Router y middlewares HTTP. |
| `github.com/justinas/nosurf` | Protección CSRF. |
| `github.com/alexedwards/scs/v2` | Gestión y persistencia de sesiones. |

La aplicación usa también la biblioteca estándar, principalmente `net/http`, `html/template`, `log` y `time`.

Para agregar una dependencia:

```bash
go get github.com/ejemplo/paquete
go mod tidy
```

Versioná `go.mod` y `go.sum` juntos.

## Estado actual y próximos pasos

El proyecto ya tiene:

- composición de dependencias en `cmd/web/main.go`;
- router Chi y middlewares;
- renderizado con layouts y cache opcional;
- sesiones con SCS;
- middleware CSRF;
- páginas Home y About.

Todavía falta, si el proyecto lo necesita:

- agregar tests HTTP y de paquetes;
- extraer servicios de negocio cuando la lógica crezca;
- incorporar persistencia y repositories reales;
- usar un store de sesiones compartido en producción;
- manejar configuración desde el entorno;
- decidir si se incorpora Vite, Tailwind, Alpine.js, HTMX y ECharts.

La decisión de frontend está documentada, con ejemplos y trade-offs, en [HTMX.md](HTMX.md).
