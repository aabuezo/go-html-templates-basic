# Formularios y validación server-side

La validación del navegador mejora la experiencia, pero nunca reemplaza la validación del servidor. El cliente puede modificarse, omitirse o no existir.

## Flujo general

```text
GET
  → mostrar formulario vacío o con valores iniciales

POST
  → parsear request
  → construir input específico
  → validar en el servidor
  → si hay errores, renderizar nuevamente con valores y errores
  → si es válido, ejecutar el caso de uso
  → mostrar éxito o redirigir
```

El input del formulario debe ser específico del caso de uso. Un formulario genérico puede ayudar a transportar valores y errores, pero no debería contener reglas de negocio de todos los formularios.

## Separar responsabilidades

```go
type CreateResourceInput struct {
	Name  string
	Email string
}

type CreateResourceForm struct {
	Values map[string]string
	Errors map[string][]string
}
```

- El modelo/input representa datos recibidos.
- El validador decide si esos datos cumplen las reglas.
- El servicio decide si la operación tiene sentido en el dominio.
- El handler decide cómo traducir el resultado a HTTP.

## Errores

Conviene conservar todos los errores relevantes por campo cuando el usuario puede corregirlos de una vez. Un componente visual puede mostrar el primero, todos o una lista de requisitos; eso es una decisión de presentación, no una limitación del validador.

Por ejemplo, una contraseña puede devolver estados independientes para longitud, mayúsculas, minúsculas, dígitos y caracteres especiales. El frontend puede mostrarlos en tiempo real, pero el backend debe recalcularlos al recibir el formulario.

Los mensajes no deben revelar información sensible. En autenticación, por ejemplo, conviene evitar distinguir entre usuario inexistente y contraseña incorrecta.

## Preservar datos tras un error

Al re-renderizar un formulario inválido:

- conservar los valores no sensibles que sean útiles;
- no volver a mostrar contraseñas recibidas;
- conservar los errores por campo;
- conservar errores generales;
- mostrar el primer campo inválido de forma accesible;
- evitar que un refresh reenvíe el POST si el flujo puede usar PRG.

## Éxito: renderizar o redirigir

Usar redirect después de un POST cuando:

- la operación terminó y la siguiente página es otra;
- se quiere aplicar Post/Redirect/Get;
- no se debe repetir la operación al refrescar.

Renderizar la misma página puede ser mejor cuando:

- el usuario carga varios registros consecutivos;
- el resultado debe quedar visible junto al formulario;
- el formulario forma parte de una interacción HTMX;
- se necesita conservar contexto de la página.

## Seguridad

- Usar límites de tamaño para el body y para cada campo.
- Validar tipos, rangos, relaciones y permisos en el servidor.
- Normalizar solo cuando la regla del dominio lo permita.
- No confiar en campos ocultos para autorizar una operación.
- Proteger formularios que cambian estado contra CSRF.
- Escapar los valores al renderizar HTML.
- Guardar contraseñas con un algoritmo adaptativo diseñado para passwords, nunca en texto plano ni con un hash rápido genérico.
- Aplicar rate limiting donde el formulario pueda abusarse.

## Checklist

- [ ] Existe un input específico para el caso de uso.
- [ ] La validación server-side funciona sin JavaScript.
- [ ] Los errores se muestran por campo y de forma accesible.
- [ ] Los valores válidos se preservan al re-renderizar.
- [ ] Los secretos no se vuelven a renderizar.
- [ ] El endpoint devuelve el status HTTP apropiado.
- [ ] El POST exitoso no se repite accidentalmente al refrescar.
- [ ] Hay tests para valores válidos, vacíos, límites y combinaciones inválidas.
