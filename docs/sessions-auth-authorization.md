# Sesiones, autenticación y autorización

Autenticación y autorización son problemas distintos:

- autenticación: ¿quién es la persona o servicio?
- autorización: ¿qué puede hacer sobre qué recurso?

## Sesiones

Una cookie de sesión normalmente contiene un identificador opaco. Los datos de sesión se guardan en el servidor o en un store externo:

```text
Cookie: session-id opaco
Servidor/store: user_id, expiración, flags y datos mínimos
```

No poner contraseñas, tokens sensibles ni grandes cantidades de datos en la cookie. Guardar solamente lo necesario para identificar el contexto de la request.

Para una sola instancia, memoria puede servir durante el desarrollo. Para varias instancias o reinicios frecuentes, usar un store compartido como Redis o una base de datos, según los requisitos de latencia, durabilidad, operación y costo.

## Flujo de login

1. Recibir y validar el formulario.
2. Buscar la cuenta sin revelar si existe.
3. Comparar la contraseña con un hash adaptativo.
4. Aplicar controles de intentos y alertas.
5. Regenerar el identificador de sesión después de autenticar.
6. Guardar el identificador mínimo del usuario.
7. Redirigir al destino permitido.

Nunca aceptar una URL de redirección arbitraria sin validarla contra destinos permitidos.

## Middleware de autenticación

Los endpoints protegidos pueden agruparse en un router o subrouter:

```go
protected := chi.NewRouter()
protected.Use(RequireAuthentication)
protected.Get("/resources", listResources)
protected.Post("/resources", createResource)
```

El middleware debería obtener el usuario desde la sesión, cargar el contexto necesario y redirigir o devolver `401` cuando no hay autenticación. Las respuestas HTML y las respuestas para una API pueden requerir tratamientos distintos.

## Autorización

No alcanza con comprobar que existe `user_id`. Cada acción debe verificar su permiso:

```text
resource:read
resource:create
resource:update
resource:approve
```

Un usuario puede tener `resource:read` sin tener los otros permisos. El chequeo debe ejecutarse en el servidor y, cuando corresponda, en más de una capa: middleware para permisos generales y servicio/repositorio para el alcance del recurso.

Ejemplo de alcance:

```text
Permiso global: resource:update
Restricción adicional: el recurso pertenece al tenant del usuario
```

El frontend puede ocultar botones, pero eso nunca reemplaza el chequeo del endpoint.

## Cookies y expiración

Revisar según el caso:

- `Secure` en producción;
- `HttpOnly` para impedir acceso desde JavaScript cuando corresponda;
- `SameSite` apropiado;
- expiración absoluta e inactividad;
- logout que invalide la sesión;
- revocación después de cambios de contraseña o privilegios;
- rotación de secretos y claves.

## Checklist

- [ ] Los endpoints protegidos tienen autenticación centralizada.
- [ ] Cada operación verifica autorización específica.
- [ ] Se controla el tenant, propietario o alcance del recurso.
- [ ] El ID de sesión se regenera después del login.
- [ ] La cookie tiene atributos seguros en producción.
- [ ] El store de sesiones es adecuado para la cantidad de instancias.
- [ ] El logout invalida la sesión.
- [ ] Hay límites para login, recuperación y códigos de verificación.
- [ ] Los logs no contienen passwords, cookies ni tokens completos.
