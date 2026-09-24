# Frontend server-rendered con Go, HTMX, Tailwind, Alpine.js y ECharts

Esta guía documenta una forma simple de construir páginas web server-rendered con Go. Go genera el HTML usando templates, HTMX agrega interacciones sin convertir la aplicación en una SPA, Tailwind se ocupa de los estilos, Alpine.js de las interacciones pequeñas del navegador y ECharts de los gráficos. TypeScript puede organizar y tipar el código del frontend sin convertir la aplicación en una SPA.

La idea general es:

```text
Navegador
   │
   ├── pide una página
   ▼
Go + html/template ──► HTML completo
   │
   └── responde a HTMX ─► fragmentos HTML
```

Esta guía describe la dirección recomendada del frontend. El repositorio todavía usa Bootstrap desde CDN en `templates/base.layout.tmpl` y no tiene una carpeta `frontend/` instalada. Por eso algunos ejemplos son decisiones propuestas y no código que ya esté activo.

## Índice

- [Responsabilidad de cada herramienta](#1-responsabilidad-de-cada-herramienta)
- [Estructura recomendada](#2-estructura-recomendada)
- [Templates de Go](#3-templates-de-go)
- [Servir archivos estáticos](#4-servir-archivos-estáticos)
- [Instalar Vite, Tailwind y TypeScript](#5-instalar-vite-tailwind-y-typescript)
- [Incluir el frontend compilado en un template](#6-incluir-el-frontend-compilado-en-un-template)
- [Primeros pasos con HTMX](#7-primeros-pasos-con-htmx)
- [Devolver un parcial desde Go](#8-devolver-un-parcial-desde-go)
- [Acciones comunes de HTMX](#9-acciones-comunes-de-htmx)
- [JavaScript o TypeScript junto con HTMX](#10-javascript-o-typescript-junto-con-htmx)
- [Validación](#11-validación-navegador-y-backend)
- [Seguridad](#12-csrf-autenticación-y-seguridad)
- [Códigos HTTP](#13-htmx-y-códigos-http)
- [¿JavaScript o TypeScript?](#14-javascript-o-typescript)
- [¿Cuándo usar Vue?](#15-cuándo-usar-vue)
- [Estado, Alpine.js y ECharts](#16-estado-stateless-alpinejs-y-echarts)
- [Flujo de trabajo recomendado](#17-flujo-de-trabajo-recomendado)
- [Checklist](#18-checklist-para-agregar-una-nueva-interacción)

## 1. Responsabilidad de cada herramienta

### Go

Go se ocupa de:

- rutas y handlers;
- validación confiable de los datos;
- acceso a la base de datos;
- autenticación y autorización;
- renderizado de templates;
- respuestas completas o fragmentos HTML.

### `html/template`

Genera HTML en el servidor:

```gotemplate
<h1>{{ .Title }}</h1>
<p>Hola, {{ .User.Name }}</p>
```

Los valores se escapan automáticamente como HTML, lo que ayuda a evitar XSS. Para contenido proveniente de usuarios no conviene convertir strings a `template.HTML` manualmente.

### HTMX

HTMX usa atributos HTML para hacer peticiones y reemplazar partes de la página:

```html
<button hx-get="/saludo" hx-target="#resultado">
    Saludar
</button>

<div id="resultado"></div>
```

El endpoint `/saludo` devuelve HTML, no JSON:

```go
func Saludo(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<p class="text-green-700">¡Hola desde Go!</p>`))
}
```

### Tailwind

Tailwind genera CSS a partir de las clases usadas en los templates:

```html
<h1 class="text-3xl font-bold text-blue-600">
    Página principal
</h1>
```

### Alpine.js y TypeScript

Alpine.js agrega comportamiento declarativo directamente en los templates mediante atributos como `x-data`, `x-show` y `@click`. Es útil para dropdowns, modales, tabs, filtros y otros estados pequeños del navegador.

TypeScript no reemplaza a Go ni a HTMX: se usa para organizar el código que sí necesita ejecutarse en el navegador. Las librerías JavaScript, como Alpine.js y ECharts, se pueden importar desde archivos `.ts`; no hace falta mantener una mezcla de archivos propios `.js` y `.ts`.

La validación importante siempre debe repetirse en Go.

## 2. Estructura recomendada

```text
.
├── go.mod
├── go.sum
├── cmd/
│   └── web/
│       ├── main.go
│       ├── middleware.go
│       └── routes.go
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── handlers.go
│   ├── models/
│   │   └── templateData.go
│   ├── render/
│   │   └── render.go
│   └── session/
│       └── sessions.go
├── templates/
│   ├── about.page.tmpl
│   ├── base.layout.tmpl
│   └── home.page.tmpl
├── frontend/                  # código fuente del frontend, si se usa Vite
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.ts
│       └── style.css
└── static/                    # salida generada por Vite, servida por Go
    ├── css/app.css
    └── js/app.js
```

`cmd/web` contiene el punto de entrada de la aplicación web, sus rutas y sus middlewares. Los paquetes reutilizables viven en `pkg`: la configuración y las sesiones se inicializan desde `main.go`, mientras que los handlers delegan el renderizado en `pkg/render`.

Los templates de página usan el sufijo `.page.tmpl` y los layouts el sufijo `.layout.tmpl`. `pkg/render` busca ambos grupos directamente dentro de `templates/` y arma el cache de templates al iniciar la aplicación.

`frontend/` y `static/` son opcionales en el estado actual. Se agregan cuando se decide usar Vite; `static/` contiene archivos generados y no debería ser la fuente de edición manual.

## 3. Templates de Go

### Datos para un template

Definimos structs para que el handler pase datos claros al template:

```go
type HomeData struct {
	Title string
	User  UserData
}

type UserData struct {
	Name  string
	Email string
}
```

El handler los usa así:

```go
func Home(w http.ResponseWriter, r *http.Request) {
	data := HomeData{
		Title: "Inicio",
		User: UserData{
			Name:  "Ana",
			Email: "ana@example.com",
		},
	}

	renderTemplate(w, "home.page.tmpl", data)
}
```

Una función de renderizado básica puede ser:

```go
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	parsedTemplate, err := template.ParseFiles("./templates/" + tmpl)
	if err != nil {
		http.Error(w, "no se pudo cargar el template", http.StatusInternalServerError)
		return
	}

	if err := parsedTemplate.Execute(w, data); err != nil {
		http.Error(w, "no se pudo renderizar el template", http.StatusInternalServerError)
	}
}
```

Es importante revisar los errores de `ParseFiles` y `Execute`; ignorarlos dificulta encontrar errores de sintaxis o de datos.

### Condicionales y loops

```gotemplate
{{ if .User.Name }}
    <p>Hola, {{ .User.Name }}</p>
{{ else }}
    <p>Visitante</p>
{{ end }}
```

```gotemplate
<ul>
{{ range .Items }}
    <li>{{ .Name }}</li>
{{ else }}
    <li>No hay elementos.</li>
{{ end }}
</ul>
```

### Layouts y parciales

Para evitar repetir el `head`, la navegación y los scripts, se pueden definir templates nombrados:

```gotemplate
{{ define "layout" }}
<!doctype html>
<html lang="es">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/css/app.css">
    <script src="/static/js/htmx.min.js" defer></script>
    <title>{{ .Title }}</title>
</head>
<body class="bg-slate-100 text-slate-900">
    {{ template "content" . }}
    <script src="/static/js/app.js" defer></script>
</body>
</html>
{{ end }}
```

En un proyecto pequeño se puede empezar con un archivo por página. Cuando haya varias páginas, conviene parsear un conjunto de archivos y ejecutar el layout con `ExecuteTemplate`.

## 4. Servir archivos estáticos

El repositorio actual todavía no tiene `static/` ni un bundle de Vite. Cuando se incorpore, Chi puede servir la salida generada por Vite desde `cmd/web/routes.go`:

```go
func routes(handler *handlers.Handler, app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf(app))
	mux.Use(SessionLoad(app))

	mux.Handle("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	))

	mux.Get("/", handler.Home)
	mux.Get("/about", handler.About)

	return mux
}
```

Una URL como `/static/css/app.css` buscará el archivo `static/css/app.css`. Los archivos dentro de `static/` son artefactos generados: se editan sus fuentes en `frontend/` y se ejecuta `npm run build`.

## 5. Instalar Vite, Tailwind y TypeScript

Desde la raíz del proyecto:

```bash
mkdir frontend
cd frontend
npm create vite@latest . -- --template vanilla-ts
npm install
npm install tailwindcss @tailwindcss/vite
```

Tailwind v4 usa el plugin `@tailwindcss/vite` y se importa desde CSS con `@import "tailwindcss"`.

### `frontend/src/style.css`

```css
@import "tailwindcss";

/* Los templates están un nivel por encima de frontend. */
@source "../../templates";
```

El `@source` es importante: le indica a Tailwind que también debe buscar clases dentro de los archivos `.tmpl`.

### `frontend/src/main.ts`

```ts
import "./style.css";

console.log("Frontend cargado");
```

Aunque el archivo sea TypeScript, se puede mantener muy pequeño. No hace falta usar TypeScript para los datos que ya fueron renderizados por Go.

### `frontend/vite.config.ts`

```ts
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import { resolve } from "node:path";

export default defineConfig({
  plugins: [tailwindcss()],
  build: {
    outDir: "../static",
    emptyOutDir: true,
    rollupOptions: {
      input: resolve(__dirname, "src/main.ts"),
      output: {
        entryFileNames: "js/app.js",
        assetFileNames: (assetInfo) => {
          if (assetInfo.name?.endsWith(".css")) {
            return "css/app.css";
          }

          return "assets/[name][extname]";
        },
      },
    },
  },
});
```

Si se usa `__dirname` y TypeScript muestra un error de tipos de Node, instalá:

```bash
npm install -D @types/node
```

### Instalar Alpine.js y ECharts

Desde `frontend` se pueden instalar ambas librerías como dependencias del bundle:

```bash
cd frontend
npm install alpinejs echarts
```

Alpine.js y ECharts son librerías JavaScript, pero se importan normalmente desde `main.ts`. El código propio puede mantenerse en TypeScript y Vite se encarga de empaquetar todo en `static/js/app.js`.

### Scripts de `package.json`

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "build:watch": "vite build --watch"
  }
}
```

Para generar los archivos que servirá Go:

```bash
cd frontend
npm run build
```

La salida será:

```text
static/
├── css/app.css
└── js/app.js
```

Durante el desarrollo, una opción simple es dejar Go ejecutándose en una terminal y Vite recompilando en otra:

```bash
# Terminal 1
go run ./cmd/web

# Terminal 2
cd frontend
npm run build:watch
```

## 6. Incluir el frontend compilado en un template

Actualmente `templates/base.layout.tmpl` carga Bootstrap desde CDN. Si se incorpora Vite, el layout debe reemplazar o complementar esa carga con los archivos generados:

```html
<!doctype html>
<html lang="es">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/css/app.css">
    <script src="/static/js/htmx.min.js" defer></script>
    <title>{{ .Title }}</title>
</head>
<body class="bg-slate-100 text-slate-900">
    <main class="mx-auto max-w-4xl p-8">
        <h1 class="text-3xl font-bold text-blue-600">{{ .Title }}</h1>
    </main>

    <script src="/static/js/app.js" defer></script>
</body>
</html>
```

HTMX se puede cargar desde CDN durante el aprendizaje, aunque para producción conviene descargar `htmx.min.js` y servirlo desde `static/js`.

## 7. Primeros pasos con HTMX

### `hx-get`

```html
<button
    class="rounded bg-blue-600 px-4 py-2 text-white"
    hx-get="/saludo"
    hx-target="#resultado"
    hx-swap="innerHTML"
>
    Cargar saludo
</button>

<div id="resultado" class="mt-4"></div>
```

- `hx-get`: URL que se llama.
- `hx-target`: elemento que recibe la respuesta.
- `hx-swap="innerHTML"`: reemplaza el contenido interno del target.

Otros swaps útiles:

```html
hx-swap="outerHTML"
hx-swap="beforebegin"
hx-swap="afterend"
hx-swap="delete"
```

### Indicador de carga

```html
<button hx-get="/datos" hx-target="#resultado" hx-indicator="#loading">
    Cargar
</button>

<span id="loading" class="htmx-indicator">Cargando...</span>
```

```css
.htmx-indicator {
    display: none;
}

.htmx-request .htmx-indicator,
.htmx-request.htmx-indicator {
    display: inline;
}
```

### Enviar formularios

```html
<form
    hx-post="/usuarios"
    hx-target="#resultado"
    hx-swap="innerHTML"
    class="space-y-4"
>
    <div>
        <label for="email">Email</label>
        <input
            id="email"
            name="email"
            type="email"
            required
            class="rounded border p-2"
        >
    </div>

    <button class="rounded bg-blue-600 px-4 py-2 text-white">
        Guardar
    </button>
</form>

<div id="resultado"></div>
```

El handler:

```go
func CrearUsuario(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`<p class="text-red-600">El email es obligatorio.</p>`))
		return
	}

	// Acá iría la validación de formato, autorización y persistencia.
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`<p class="text-green-700">Usuario creado.</p>`))
}
```

La validación `required` y `type="email"` mejora la experiencia, pero no reemplaza la validación del handler.

## 8. Devolver un parcial desde Go

HTMX funciona mejor cuando cada endpoint tiene claro qué fragmento devuelve.

`templates/partials/flash.tmpl`:

```gotemplate
{{ define "flash" }}
<div class="rounded border p-3 {{ .Class }}">
    {{ .Message }}
</div>
{{ end }}
```

El handler debe parsear el parcial y ejecutar el nombre definido:

```go
func Flash(w http.ResponseWriter, class, message string) {
	tmpl, err := template.ParseFiles("./templates/partials/flash.tmpl")
	if err != nil {
		http.Error(w, "template no disponible", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"Class":   class,
		"Message": message,
	}

	if err := tmpl.ExecuteTemplate(w, "flash", data); err != nil {
		http.Error(w, "no se pudo renderizar el mensaje", http.StatusInternalServerError)
	}
}
```

En una aplicación más grande conviene tener un renderer que parsee los templates una sola vez al iniciar el programa, en vez de leerlos del disco en cada request.

## 9. Acciones comunes de HTMX

### Buscar mientras se escribe

```html
<input
    name="q"
    type="search"
    placeholder="Buscar..."
    hx-get="/usuarios/buscar"
    hx-trigger="keyup changed delay:300ms"
    hx-target="#resultados"
>

<div id="resultados"></div>
```

### Borrar un elemento

```html
<button
    hx-delete="/usuarios/42"
    hx-target="closest li"
    hx-swap="outerHTML"
    hx-confirm="¿Eliminar este usuario?"
>
    Eliminar
</button>
```

El servidor debe comprobar que el usuario autenticado tiene permiso para borrar el recurso.

### Recargar una parte después de guardar

```html
<form hx-post="/usuarios" hx-on::after-request="this.reset()">
    <!-- campos -->
</form>
```

Para lógica más compleja, es preferible usar un archivo JavaScript en vez de colocar mucho código dentro de los atributos HTML.

## 10. JavaScript o TypeScript junto con HTMX

`static/js/app.js`:

```js
document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("[data-confirm]").forEach((element) => {
    element.addEventListener("click", (event) => {
      const message = element.dataset.confirm;

      if (message && !window.confirm(message)) {
        event.preventDefault();
      }
    });
  });
});

document.body.addEventListener("htmx:afterRequest", (event) => {
  if (event.detail.successful) {
    console.log("Petición HTMX completada");
  }
});
```

Usar `data-*` permite mantener la configuración en el template sin mezclar demasiado JavaScript con el HTML:

```html
<button data-confirm="¿Querés continuar?" hx-post="/accion">
    Continuar
</button>
```

HTMX reemplaza fragmentos del DOM, por lo que los listeners agregados manualmente a elementos reemplazados pueden perderse. Para esos casos se puede:

- usar delegación de eventos sobre `document` o `body`;
- escuchar eventos de HTMX como `htmx:afterSwap`;
- usar atributos HTMX cuando la interacción sea simple.

## 11. Validación: navegador y backend

La validación debe pensarse en capas:

```text
HTML/JS: feedback rápido para el usuario
        ↓
Go: validación real y reglas de negocio
        ↓
Base de datos: constraints e integridad
```

Ejemplo de validación en Go:

```go
func validarEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("el email es obligatorio")
	}

	if !strings.Contains(email, "@") {
		return errors.New("el email no es válido")
	}

	return nil
}
```

En una aplicación real conviene usar validaciones más completas y devolver errores específicos dentro del fragmento HTML correspondiente al campo.

## 12. CSRF, autenticación y seguridad

HTMX no elimina las reglas normales de seguridad.

- Validá siempre en Go.
- Verificá sesión y permisos en cada handler protegido.
- Usá tokens CSRF en formularios que cambian datos.
- No confíes en valores enviados por el navegador.
- Usá `http.MethodPost`, `http.MethodDelete`, etc., según la operación.
- Escapá los valores mediante `html/template`.
- No uses `template.HTML` para saltarte el escaping salvo que el contenido sea controlado y entiendas el riesgo.

Un token CSRF puede viajar como un campo oculto:

```gotemplate
<form hx-post="/perfil">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    <!-- campos -->
</form>
```

El handler debe comprobar ese token antes de modificar datos.

## 13. HTMX y códigos HTTP

Usá códigos HTTP significativos:

- `200 OK`: respuesta normal.
- `201 Created`: recurso creado.
- `400 Bad Request`: request mal formado.
- `401 Unauthorized`: falta autenticación.
- `403 Forbidden`: no hay permiso.
- `404 Not Found`: recurso inexistente.
- `422 Unprocessable Entity`: datos válidos como HTTP, pero inválidos para el formulario.
- `500 Internal Server Error`: error del servidor.

Para un formulario HTMX, un error de validación suele devolver un fragmento con estado `422` para que el usuario vea los errores. Un error inesperado debe registrarse en el servidor y no exponer detalles internos.

## 14. ¿JavaScript o TypeScript?

Para esta arquitectura, JavaScript simple suele ser suficiente cuando:

- Go renderiza los datos;
- HTMX hace la mayoría de las peticiones;
- las interacciones son pequeñas;
- no hay un estado complejo en el navegador.

TypeScript empieza a ser conveniente cuando:

- hay varios módulos de frontend;
- se consume bastante JSON con `fetch`;
- existe estado complejo del lado del cliente;
- hay muchos componentes reutilizables;
- el JavaScript creció lo suficiente como para que los errores de tipos sean frecuentes.

No hace falta usar TypeScript solo porque el proyecto usa Vite. También se puede usar Vite con JavaScript:

```bash
npm create vite@latest . -- --template vanilla
```

## 15. ¿Cuándo usar Vue?

Go templates + HTMX suele ser muy cómodo para:

- páginas públicas;
- CRUDs;
- formularios;
- paneles administrativos;
- contenido que depende del servidor.

Vue suele convenir cuando una parte de la aplicación necesita:

- mucho estado local;
- componentes muy interactivos;
- drag and drop complejo;
- edición en tiempo real;
- navegación y actualización de datos principalmente en el cliente.

También se puede usar una solución híbrida: Go templates para la estructura general y Vue solo para una sección compleja.

## 16. Estado, stateless, Alpine.js y ECharts

### ¿Go como responsable del HTML significa que la aplicación no es stateless?

No necesariamente. Son conceptos diferentes.

Una aplicación puede usar Go para renderizar HTML y seguir teniendo requests HTTP stateless. Esto significa que cada request contiene la información necesaria para procesarse y que el servidor no depende de variables globales de memoria para recordar una request anterior.

El estado puede vivir en distintos lugares:

```text
Request HTTP
    │
    ├── Cookie/sesión ──► identifica al usuario
    ├── Base de datos ──► usuarios, pedidos, configuración
    └── Go ─────────────► procesa la request y genera HTML
```

Por ejemplo, Go puede leer una cookie de sesión, consultar la base de datos y renderizar una página. La siguiente request vuelve a hacer lo mismo. La aplicación sigue usando HTTP de forma stateless, aunque la sesión y los datos de negocio representen estado persistente.

Lo que conviene evitar es guardar datos importantes en variables globales:

```go
// Evitar para datos de usuarios o de negocio.
var usuarios = map[string]string{}
```

Eso suele fallar con múltiples usuarios, reinicios del proceso o varias instancias de la aplicación. Para datos importantes, usar una base de datos o un almacenamiento compartido.

HTMX tampoco cambia esta regla. Una request HTMX sigue siendo una request HTTP normal, normalmente acompañada por cookies y encabezados como `HX-Request`.

### ECharts con Go templates y HTMX

ECharts se ejecuta en el navegador, mientras que Go puede entregar el HTML y los datos del gráfico. La separación recomendada es:

```text
Go       ──► HTML, fragmentos HTMX y datos JSON
HTMX     ──► requests y reemplazo de HTML
ECharts  ──► gráfico en el navegador
```

Si todavía no usás Vite, ECharts también puede cargarse desde un CDN:

```html
<script src="https://cdn.jsdelivr.net/npm/echarts@6/dist/echarts.min.js" defer></script>
```

Pero si ya usás Vite para Tailwind, conviene instalar ECharts con npm e importarlo desde TypeScript. Así controlás la versión, evitás depender de un CDN y mantenés todas las dependencias del frontend en el mismo bundle.

### Alpine.js con Vite

En `frontend/src/main.ts` se importa Alpine y se inicia una sola vez:

```ts
import Alpine from "alpinejs";
import "./style.css";

declare global {
    interface Window {
        Alpine: typeof Alpine;
    }
}

window.Alpine = Alpine;
Alpine.start();
```

`window.Alpine` es opcional, pero resulta útil para inspeccionar Alpine desde las herramientas del navegador. No hay que llamar a `Alpine.start()` más de una vez.

El template puede declarar una interacción sin escribir un archivo JavaScript separado:

```gotemplate
<div x-data="{ open: false }" class="relative">
    <button
        type="button"
        @click="open = !open"
        class="rounded bg-blue-600 px-4 py-2 text-white"
    >
        Opciones
    </button>

    <div x-show="open" x-cloak class="mt-2 rounded border bg-white p-4">
        Contenido de las opciones
    </div>
</div>
```

Para evitar un parpadeo antes de que Alpine se inicie, agregá este estilo:

```css
[x-cloak] {
    display: none !important;
}
```

Usá Alpine para estado local y pequeño. Si la interacción necesita datos del servidor, combiná Alpine con HTMX; si necesita un gráfico, delegá el dibujo a ECharts.

### ECharts instalado con Vite

ECharts también se puede importar directamente desde `main.ts`:

```ts
import Alpine from "alpinejs";
import * as echarts from "echarts";
import "./style.css";

window.Alpine = Alpine;
Alpine.start();

function initializeCharts(root: ParentNode = document) {
    root.querySelectorAll<HTMLElement>("[data-chart]").forEach((element) => {
        if (element.dataset.chartInitialized === "true") {
            return;
        }

        element.dataset.chartInitialized = "true";
        const chart = echarts.init(element);

        chart.setOption({
            title: { text: "Ventas por mes" },
            tooltip: {},
            xAxis: {
                type: "category",
                data: ["Enero", "Febrero", "Marzo", "Abril"],
            },
            yAxis: { type: "value" },
            series: [
                {
                    name: "Ventas",
                    type: "bar",
                    data: [1200, 1800, 1500, 2300],
                },
            ],
        });

        window.addEventListener("resize", () => chart.resize());
    });
}

document.addEventListener("DOMContentLoaded", () => {
    initializeCharts();
});

document.body.addEventListener("htmx:afterSwap", (event) => {
    initializeCharts(event.target as Element);
});
```

El template solo necesita un contenedor con tamaño definido:

```gotemplate
<div
    data-chart
    class="h-80 w-full"
></div>
```

El tamaño es importante porque ECharts necesita que el contenedor tenga ancho y alto antes de inicializarse. Si el contenedor llega mediante un `hx-swap`, `htmx:afterSwap` permite inicializarlo después de que HTMX lo inserte.

Para gráficos reales, Go puede entregar los datos como JSON dentro de un endpoint y el módulo TypeScript puede obtenerlos:

```go
type SalesPoint struct {
    Month string  `json:"month"`
    Value float64 `json:"value"`
}

func SalesData(w http.ResponseWriter, r *http.Request) {
    data := []SalesPoint{
        {Month: "Enero", Value: 1200},
        {Month: "Febrero", Value: 1800},
        {Month: "Marzo", Value: 1500},
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(data)
}
```

El código de ECharts puede leer la URL desde el template:

```gotemplate
<div
    data-chart-url="/api/ventas"
    class="h-80 w-full"
></div>
```

En ese caso, `initializeCharts` puede obtener `element.dataset.chartUrl`, llamar a `fetch` y construir las series con los valores recibidos. Para una aplicación pequeña, mantener Alpine, HTMX y ECharts en módulos separados dentro de `main.ts` suele ser suficiente.

### ¿Hace falta mezclar JavaScript y TypeScript?

No. Que Alpine.js y ECharts estén escritos en JavaScript no obliga a que tu aplicación tenga archivos propios `.js`. Una opción coherente es:

```text
frontend/src/
├── main.ts       entrada del bundle
├── charts.ts     inicialización de ECharts
├── alpine.ts     stores o componentes Alpine complejos
└── style.css
```

Todo el código propio queda en TypeScript y las dependencias se importan desde npm. Solo mantendría archivos `.js` separados si ya existe código JavaScript que no querés migrar o si una librería exige cargarse como script independiente.

La combinación recomendada para este proyecto es:

```text
Go       ──► HTML, fragmentos HTMX y datos JSON
HTMX     ──► requests y reemplazo de HTML
Alpine   ──► estado local e interacciones pequeñas
ECharts  ──► gráficos
TypeScript ─► código que coordina Alpine, ECharts y el DOM
```

## 17. Flujo de trabajo recomendado

Durante el aprendizaje:

```bash
# Terminal 1: servidor Go
go run ./cmd/web

# Terminal 2: recompilar CSS/JS al guardar
cd frontend
npm run build:watch
```

El ciclo normal es:

1. Crear o modificar un template `.tmpl`.
2. Agregar clases Tailwind.
3. Ejecutar o dejar activo `npm run build:watch`.
4. Crear un handler Go para la acción.
5. Si la acción devuelve una parte de la página, hacer que responda un parcial.
6. Agregar atributos HTMX al elemento que dispara la acción.
7. Usar JavaScript solo cuando HTMX y HTML no sean suficientes.
8. Validar de nuevo todo en Go.

## 18. Checklist para agregar una nueva interacción

- ¿La acción necesita datos del servidor?
- ¿Puede resolverla un `hx-get`, `hx-post` o `hx-delete`?
- ¿Qué elemento debe actualizarse (`hx-target`)?
- ¿El endpoint devuelve una página completa o un parcial?
- ¿Qué debe validar el navegador?
- ¿Qué debe validar obligatoriamente Go?
- ¿El usuario tiene permisos para la acción?
- ¿Hace falta CSRF?
- ¿Se necesita JavaScript o alcanza con HTMX?
- ¿El CSS usado está dentro de los archivos que Tailwind escanea?

La regla general es mantener el servidor como fuente de verdad, usar HTMX para intercambiar HTML y reservar JavaScript para interacciones que realmente necesitan lógica en el navegador.
