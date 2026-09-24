# Arquitectura de la aplicación

Este proyecto es una aplicación web escrita en Go. La aplicación está organizada en paquetes con responsabilidades separadas y utiliza inyección de dependencias para conectar sus componentes.

## Idea principal

`main` funciona como el punto de composición de la aplicación:

1. Crea la configuración general.
2. Crea el cache de templates.
3. Crea el renderer.
4. Crea los handlers e inyecta sus dependencias.
5. Registra las rutas.
6. Inicia el servidor HTTP.

El resto de los paquetes no crea dependencias globales por su cuenta. Recibe lo que necesita desde `main`.

```text
main
 └── AppConfig
      └── Renderer
           └── Handler
```

En código, el armado principal se ve así:

```go
app := &config.AppConfig{
    UseCache:      false,
    TemplateCache: tc,
}

renderer := render.NewRenderer(app)
handler := handlers.NewHandler(app, renderer)
```

## Estructura de paquetes

```text
cmd/web/main.go       Punto de entrada y composición de la aplicación
pkg/config             Configuración compartida
pkg/handlers           Handlers HTTP
pkg/models             Estructuras de datos
pkg/render             Renderizado de templates
templates              Templates HTML
```

### `pkg/config`

Define `AppConfig`, que contiene dependencias y opciones compartidas:

```go
type AppConfig struct {
    UseCache      bool
    TemplateCache map[string]*template.Template
    InfoLog       *log.Logger
}
```

El paquete define el tipo, pero la instancia se crea en `main`. Esto evita que la configuración sea una variable global.

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
    App    *config.AppConfig
    Render *render.Renderer
}
```

Los handlers reciben un `Renderer` y lo utilizan para responder las requests:

```go
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    h.Render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
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

- renombrar el campo `Render` a `Renderer` para que coincida con el tipo;
- eliminar `Handler.App` si los handlers no lo necesitan directamente;
- usar `http.Error` en lugar de terminar el proceso con `log.Fatal` durante una request;
- manejar los errores de `CreateTemplateCache` en vez de ignorarlos;
- agregar interfaces para facilitar mocks en los tests;
- agregar un repository real para el acceso a base de datos;
- separar servicios de negocio de los handlers cuando la lógica crezca.

## Ejecutar el proyecto

```bash
go run ./cmd/web
```

Verificar que todos los paquetes compilen:

```bash
go test ./...
```

