# Frontend server-rendered: templates, HTMX, Tailwind, Alpine.js y Vite

Esta guía propone una arquitectura progresiva. El servidor genera HTML y el navegador mejora la experiencia cuando hace falta; no es necesario empezar con una SPA.

## Responsabilidad de cada herramienta

- Go: routing, reglas de negocio, validación, autenticación, autorización y respuestas.
- `html/template` o Templ: generación de HTML server-rendered.
- HTMX: requests declarativas y reemplazo de fragmentos HTML.
- Alpine.js: estado local pequeño, como dropdowns, tabs, modales y toggles.
- Tailwind CSS: estilos utilitarios.
- Vite: bundle y desarrollo de CSS/JavaScript.

Usar la herramienta más simple que resuelva la interacción. Un formulario común no necesita HTMX solamente porque la aplicación lo tiene instalado.

## Templates

Con `html/template`:

```gotemplate
<h1>{{ .Title }}</h1>
<p>{{ .Description }}</p>
```

Los datos deben estar preparados por el handler o por una capa de presentación. No introducir lógica de negocio compleja en el template.

Templ es una alternativa en la que los archivos `.templ` se convierten en código Go generado. Los componentes siguen siendo funciones componibles y la arquitectura HTTP no cambia.

## HTMX

HTMX permite pedir HTML a un endpoint y reemplazar una parte del documento:

```html
<button hx-get="/resources/summary"
        hx-target="#summary"
        hx-swap="innerHTML">
  Actualizar
</button>

<section id="summary"></section>
```

El endpoint debe saber si responde una página completa o un fragmento. Un patrón habitual es que una misma operación tenga una representación completa y otra parcial, o que el handler elija la respuesta según `HX-Request`.

Para formularios HTMX:

- devolver el fragmento con errores y status `422` si la validación falla;
- devolver el fragmento actualizado o un redirect si tiene éxito;
- mantener CSRF y autorización normales;
- no asumir que `HX-Request` es una frontera de seguridad.

## Alpine.js

Alpine es apropiado para estado local:

```html
<div x-data="{ open: false }">
  <button type="button" @click="open = !open">Mostrar</button>
  <div x-show="open">Contenido</div>
</div>
```

Si el estado debe persistir o ser validado por el servidor, el servidor sigue siendo la fuente de verdad. Alpine no reemplaza autorización ni validación.

## Tailwind y Vite

Una estructura posible es:

```text
frontend/
├── package.json
├── vite.config.ts
└── src/
    ├── main.ts
    └── style.css
static/
└── dist/
```

En Tailwind moderno, el CSS puede comenzar con:

```css
@import "tailwindcss";
```

Tailwind debe poder detectar las clases escritas en HTML, `.tmpl` o `.templ`. Si los templates están fuera de las fuentes detectadas automáticamente, registrarlos explícitamente con `@source`.

Vite puede compilar los assets aunque el HTML lo genere Go. El flujo es independiente:

```text
templates/components → servidor Go → HTML
CSS/TypeScript       → Vite       → static/dist
```

En producción se compilan los assets durante el build de la imagen o del artefacto y Go sirve los archivos generados. No es obligatorio usar CDN.

## CDN o bundle local

CDN:

- rápido para probar;
- menos configuración inicial;
- dependencia de un tercero durante la carga;
- más trabajo para controlar versiones, CSP, disponibilidad e integridad.

Bundle local con Vite:

- versiones reproducibles;
- assets servidos por la propia aplicación o CDN controlado;
- integración con minificación, hashes y source maps;
- requiere Node y un paso de build.

Para producción suele ser preferible compilar y servir los assets propios. El CDN puede ser una decisión válida si se versiona, se evalúan sus riesgos y la política de seguridad lo permite.

## Desarrollo y producción

Desarrollo:

```text
templ generate o watch
Vite dev/watch
servidor Go
```

Producción:

```text
generar componentes
compilar CSS/JS
compilar binario Go
copiar assets compilados
servir con configuración segura
```

Mantener separados los artefactos fuente y los generados. Documentar qué comando produce cada archivo y si los archivos generados se versionan.

## Checklist de frontend

- [ ] La aplicación funciona con HTML y navegación básicos.
- [ ] HTMX solamente mejora interacciones que lo necesitan.
- [ ] Alpine se usa para estado local, no para reglas de seguridad.
- [ ] Las clases Tailwind son detectables en los templates.
- [ ] Los assets tienen un proceso reproducible de build.
- [ ] Los scripts y estilos están versionados o empaquetados de forma controlada.
- [ ] Los formularios siguen funcionando sin depender exclusivamente de JavaScript.
- [ ] Se revisan accesibilidad, foco, teclado y mensajes de error.
