# Arquitectura de la aplicación

Este proyecto es una aplicación web escrita en Go. La aplicación está organizada en paquetes con responsabilidades separadas y utiliza inyección de dependencias para conectar sus componentes.

## Idea principal

`main` funciona como el punto de composición de la aplicación:

1. Crea la configuración general.
2. Crea el cache de templates.
3. Crea el renderer.
4. Crea los handlers e inyecta sus dependencias.
5. Registra las rutas y sus middlewares.
6. Inicia el servidor HTTP.

El resto de los paquetes no crea dependencias globales por su cuenta. Recibe lo que necesita desde `main`.

```text
main
 └── AppConfig
      ├── SessionManager
      └── Renderer
           └── Handler
```

En código, el armado principal se ve así:

```go
app := &config.AppConfig{
    UseCache:      false,
    TemplateCache: tc,
    InProduction:  false,
}

app.Session = session.New(app.InProduction)

renderer := render.NewRenderer(app)
handler := handlers.NewHandler(app, renderer)
```

## Estructura de paquetes

```text
cmd/web/main.go       Punto de entrada y composición de la aplicación
cmd/web/routes.go      Router y orden de los middlewares
cmd/web/middleware.go  Middlewares de recuperación, CSRF y sesiones
pkg/config             Configuración compartida
pkg/handlers           Handlers HTTP
pkg/models             Estructuras de datos
pkg/render             Renderizado de templates
pkg/session            Creación y configuración del gestor de sesiones
templates              Templates HTML
```

### `pkg/config`

Define `AppConfig`, que contiene dependencias y opciones compartidas:

```go
type AppConfig struct {
    UseCache      bool
    TemplateCache map[string]*template.Template
    InfoLog       *log.Logger
    InProduction  bool
    Session       *scs.SessionManager
}
```

`InProduction` activa opciones seguras de cookies, como `Secure`. `Session` contiene el gestor de sesiones que se crea en `main` y se inyecta en el flujo HTTP.

### `pkg/render`

Contiene `Renderer`, que sabe cómo cargar y ejecutar templates:

```go
type Renderer struct {
    app *config.AppConfig
}

func NewRenderer(app *config.AppConfig) *Renderer {
    return &Renderer{app: app}
}
```

El renderer puede usar el cache configurado en `AppConfig` o crear los templates nuevamente, según el valor de `UseCache`.

### `pkg/handlers`

Contiene los handlers HTTP y sus dependencias:

```go
type Handler struct {
    App      *config.AppConfig
    Renderer *render.Renderer
}
```

Los handlers reciben un `Renderer` y lo utilizan para responder las requests:

```go
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    h.Renderer.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}
```

Los handlers deberían concentrarse en la lógica de la request: leer parámetros, ejecutar la lógica necesaria, preparar los datos y devolver una respuesta.

### `pkg/models`

Contiene estructuras de datos compartidas. `TemplateData` representa el formato común de datos que reciben los templates:

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

Tener una estructura común evita que cada handler pase un tipo completamente diferente al renderer. También permite agregar información transversal, como mensajes, errores o tokens CSRF.

`Data map[string]any` conserva flexibilidad para datos específicos de una página. Como contrapartida, esos datos tienen menos seguridad de tipos.

## Inyección de dependencias

La inyección de dependencias consiste en entregar un objeto desde afuera, en lugar de crearlo dentro del componente que lo usa.

Por ejemplo, `Renderer` necesita `AppConfig`. En vez de crear la configuración internamente, la recibe:

```go
renderer := render.NewRenderer(app)
```

Esto tiene varias ventajas:

- las dependencias son explícitas;
- `main` controla cómo se arma la aplicación;
- los paquetes quedan menos acoplados;
- es más sencillo reemplazar una dependencia durante un test;
- se evita el estado global compartido.

No todos los paquetes necesitan obligatoriamente un struct y un constructor. Se usan cuando un componente tiene estado o dependencias que debe conservar. Una función auxiliar simple puede existir sin constructor.

## Por qué se evitan variables globales

Una versión anterior utilizaba variables como:

```go
var Repo *Repository
var app *config.AppConfig
```

Eso permitía que cualquier función accediera a dependencias sin recibirlas explícitamente. El problema es que las dependencias quedaban ocultas:

```go
func Home(w http.ResponseWriter, r *http.Request) {
    // No se ve en la firma, pero depende de una variable global.
}
```

Las variables globales dificultan los tests, pueden generar problemas cuando hay concurrencia y hacen más difícil entender el ciclo de vida de los objetos.

La versión actual guarda las dependencias dentro de `Handler` y `Renderer`, que se crean desde `main`.

## Repository Pattern

El nombre `Repository` utilizado originalmente puede resultar confuso. Un repository, en sentido estricto, normalmente abstrae el acceso a datos:

```go
type UserRepository interface {
    FindByID(id int) (*User, error)
}
```

Una implementación podría consultar una base de datos:

```go
type PostgresUserRepository struct {
    DB *sql.DB
}
```

El objeto llamado `Repository` en la versión original no hacía eso. Simplemente contenía `AppConfig` para que los handlers pudieran acceder a la configuración. Por lo tanto, era más parecido a un contenedor de dependencias que a un Repository Pattern.

Si en el futuro se agrega persistencia, podría existir un paquete separado:

```text
pkg/repository
    user_repository.go
```

Ese repository tendría como responsabilidad consultar y guardar datos, no renderizar templates ni manejar requests HTTP.

## Flujo de una request

```text
Request HTTP
    ↓
Recoverer
    ↓
NoSurf (protección CSRF)
    ↓
SessionLoad (carga y guardado de la sesión)
    ↓
Handler.Home o Handler.About
    ↓
Construcción de TemplateData
    ↓
Renderer.RenderTemplate
    ↓
Obtención del template desde el cache o desde los archivos
    ↓
Respuesta HTML
```

El handler decide qué datos necesita la página. El renderer decide cómo ejecutar el template. Esta separación permite cambiar la forma de cargar templates sin modificar cada handler.

## CSRF y `CSRFToken`

CSRF (Cross-Site Request Forgery) es un ataque en el que un sitio externo intenta provocar una request con las credenciales de un usuario que ya está logueado en esta aplicación. El token CSRF agrega un valor impredecible que el sitio externo no puede conocer y que debe acompañar a las requests que modifican datos.

El middleware `NoSurf` se registra en `cmd/web/routes.go` y usa la librería `github.com/justinas/nosurf`. Su responsabilidad es:

- generar y validar el token CSRF;
- agregar la cookie base del mecanismo CSRF;
- proteger las requests que cambian estado, especialmente las `POST`;
- configurar la cookie con `HttpOnly`, `SameSite=Lax` y `Secure` cuando `InProduction` es `true`.

`TemplateData` tiene el campo `CSRFToken` para que un handler pueda pasarlo al template:

```go
data := &models.TemplateData{
    CSRFToken: nosurf.Token(r),
}
```

Luego el formulario debe enviarlo como un campo oculto:

```html
<input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
```

Actualmente el middleware está conectado, pero los handlers existentes (`Home` y `About`) todavía no construyen un formulario ni cargan explícitamente `CSRFToken`. Por eso el campo del modelo está preparado para los próximos formularios, pero todavía no hay un flujo de escritura protegido por CSRF.

Las requests `GET` se usan para consultar recursos y no deberían modificar datos. Las operaciones `POST`, `PUT`, `PATCH` y `DELETE` deben incluir el token CSRF cuando se incorporen.

## Middlewares

Los middlewares envuelven al router y permiten aplicar comportamiento común a varias rutas sin repetir código en cada handler.

En `routes.go` se registran en este orden:

1. `middleware.Recoverer`: recupera un `panic` y evita que la aplicación se cierre por un error inesperado de un handler.
2. `NoSurf(app)`: agrega la protección CSRF y configura su cookie.
3. `SessionLoad(app)`: carga los datos de sesión al comienzo de la request y guarda los cambios al terminarla.

`SessionLoad` no crea la sesión. Usa el gestor que ya fue creado en `main`:

```go
app.Session = session.New(app.InProduction)
```

Esto mantiene la creación de dependencias en el punto de composición de la aplicación y permite que los handlers accedan a la sesión mediante el contexto de la request cuando sea necesario.

## Sesiones

El paquete `pkg/session` centraliza la configuración del gestor de sesiones de `github.com/alexedwards/scs/v2`.

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

- la sesión dura como máximo 24 horas;
- la cookie se conserva al cerrar el navegador;
- `SameSite=Lax` reduce el envío de la cookie en requests cross-site;
- `Secure=true` en producción hace que la cookie viaje únicamente por HTTPS;
- SCS usa una cookie `HttpOnly` por defecto, por lo que JavaScript no puede leer directamente el identificador de sesión;
- el store predeterminado es en memoria (`memstore`), adecuado para desarrollo local pero no para varias instancias o reinicios de producción.

En producción, el store debería reemplazarse por uno compartido, como Redis o una base de datos, para que todas las instancias de la aplicación puedan encontrar las mismas sesiones.

El acceso a los datos de sesión se hace desde el contexto de la request usando el `SessionManager`. Por ejemplo, un handler podría leer o guardar un valor con métodos como `GetString`, `Put` o `GetBool`, sin acceder directamente a la cookie.

### Ejemplo actual: guardar y recuperar la IP

El ejemplo de los handlers usa la sesión para compartir un dato entre dos requests:

```go
// Home: guarda el valor en la sesión.
remoteIP := r.RemoteAddr
h.App.Session.Put(r.Context(), "remote_ip", remoteIP)

// About: recupera el valor guardado.
remoteIP := h.App.Session.GetString(r.Context(), "remote_ip")
```

El flujo para probarlo es:

1. visitar `/` para guardar `remote_ip` en la sesión;
2. visitar `/about` para leer el valor y mostrarlo en la página;
3. si no existe una sesión previa, `/about` muestra el mensaje de que todavía no conoce la IP.

`SessionLoad` se encarga de cargar la sesión antes de ejecutar el handler y de guardar los cambios al terminar la request. El navegador mantiene el identificador en la cookie y lo envía nuevamente al visitar `/about`, permitiendo recuperar el dato.

`r.RemoteAddr` representa la dirección remota que ve el servidor y normalmente incluye el puerto, por ejemplo `127.0.0.1:54321`. En producción puede representar la dirección de un proxy o load balancer; para obtener la IP original habría que configurar y validar explícitamente headers como `X-Forwarded-For`.


## Dependencias externas

Las dependencias de runtime están declaradas en `go.mod`:

| Dependencia | Uso |
| --- | --- |
| `github.com/go-chi/chi/v5` | Router HTTP y composición de rutas/middlewares. |
| `github.com/justinas/nosurf` | Generación y validación de tokens CSRF. |
| `github.com/alexedwards/scs/v2` | Gestión, carga y persistencia de sesiones. |

La aplicación usa además paquetes de la biblioteca estándar de Go, como `net/http`, `html/template`, `log` y `time`.

Para agregar o actualizar dependencias:

```bash
go get github.com/ejemplo/paquete
go mod tidy
```

`go.sum` guarda las sumas de comprobación de los módulos descargados. Debe versionarse junto con `go.mod` para que las instalaciones sean reproducibles.

Antes de aceptar cambios relacionados con dependencias conviene ejecutar:

```bash
go test ./...
go vet ./...
go mod tidy -diff
```

`go mod tidy -diff` permite detectar dependencias que sobran, faltan o están declaradas de forma incorrecta sin modificar los archivos.

## Cache de templates

`CreateTemplateCache` busca:

- archivos `*.page.tmpl`;
- archivos `*.layout.tmpl`.

Después combina cada página con los layouts y devuelve un mapa:

```go
map[string]*template.Template
```

En producción conviene crear el cache una sola vez al iniciar la aplicación y usarlo desde `AppConfig`. En desarrollo puede ser útil volver a cargar los archivos en cada request para ver cambios sin reiniciar el servidor.

## Principios que sigue esta arquitectura

- `main` arma la aplicación.
- Cada paquete tiene una responsabilidad concreta.
- Las dependencias se pasan explícitamente.
- Se evita compartir estado global.
- Los handlers no necesitan saber cómo se cargan los templates.
- El renderer no necesita conocer los detalles de cada endpoint.
- Los modelos comunes se ubican en `pkg/models`.

## Mejoras futuras posibles

Al crecer la aplicación se podrían hacer estos ajustes:

- eliminar `Handler.App` si los handlers no lo necesitan directamente;
- usar `http.Error` en lugar de terminar el proceso con `log.Fatal` durante una request;
- manejar los errores de `CreateTemplateCache` en vez de ignorarlos;
- agregar interfaces para facilitar mocks en los tests;
- agregar un repository real para el acceso a base de datos;
- separar servicios de negocio de los handlers cuando la lógica crezca;
- agregar formularios y handlers de escritura que usen `CSRFToken`;
- reemplazar `memstore` por un store compartido en producción.

## Ejecutar el proyecto

```bash
go run ./cmd/web
```

Verificar que todos los paquetes compilen:

```bash
go test ./...
```
