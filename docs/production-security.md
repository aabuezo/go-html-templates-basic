# Checklist de seguridad para producción

Esta lista orienta una revisión; no reemplaza un threat model, una auditoría ni asesoramiento especializado.

## Configuración y despliegue

- [ ] Los secretos vienen de un secret manager o variables protegidas.
- [ ] La configuración de producción falla al iniciar si falta un valor obligatorio.
- [ ] El proceso corre con privilegios mínimos.
- [ ] Las dependencias y la imagen base se actualizan y escanean.
- [ ] Hay backups cifrados y pruebas de restauración.
- [ ] Se documentan rollback, rotación de secretos y respuesta a incidentes.

## Transporte y headers

- [ ] Todo el tráfico externo usa HTTPS.
- [ ] Se configura HSTS cuando corresponde.
- [ ] Se revisan CSP, `frame-ancestors`, `X-Content-Type-Options` y `Referrer-Policy`.
- [ ] Los proxies confiables están configurados explícitamente.
- [ ] Se define un límite razonable de body, headers y tiempo de request.

## Sesiones y autenticación

- [ ] Las cookies usan `Secure`, `HttpOnly` y `SameSite` apropiados.
- [ ] El ID de sesión se regenera después del login y cambios de privilegios.
- [ ] Existe expiración por inactividad y absoluta según el riesgo.
- [ ] Logout y revocación invalidan la sesión.
- [ ] Las contraseñas se almacenan con un algoritmo adaptativo apropiado.
- [ ] El login, recuperación y verificación tienen rate limiting.
- [ ] Los mensajes de autenticación no permiten enumerar usuarios.
- [ ] MFA se considera para cuentas privilegiadas o de alto riesgo.

## Autorización y datos

- [ ] Cada endpoint protegido verifica autenticación.
- [ ] Cada acción verifica el permiso específico.
- [ ] Se comprueba el ownership o tenant del recurso.
- [ ] No se confía en IDs, roles ni flags enviados por el navegador.
- [ ] Se aplica mínimo privilegio a usuarios, servicios y base de datos.
- [ ] Se registran acciones sensibles y cambios de permisos.

## Requests y respuestas

- [ ] Los formularios que cambian estado tienen protección CSRF.
- [ ] Se validan tipos, tamaños, rangos y relaciones en el servidor.
- [ ] Las salidas HTML se escapan correctamente.
- [ ] Las consultas usan parámetros y no concatenan input.
- [ ] Las subidas de archivos tienen límites, validación y almacenamiento seguro.
- [ ] Los redirects externos se validan contra una allowlist.
- [ ] No se exponen stack traces, secretos ni detalles internos.

## Disponibilidad y observabilidad

- [ ] Hay límites por IP, cuenta o tenant donde corresponda.
- [ ] Las operaciones costosas tienen timeouts y límites.
- [ ] Hay logs estructurados y un request ID.
- [ ] Se monitorean errores, latencia, saturación, logins fallidos y rate limiting.
- [ ] Los logs no contienen passwords, cookies ni tokens completos.
- [ ] Hay alertas y un procedimiento de respuesta.

## Revisión final

- [ ] Se actualizó el threat model para la funcionalidad nueva.
- [ ] Se revisaron tests de autorización y límites.
- [ ] Se probó recuperación ante fallos.
- [ ] Se revisaron permisos del entorno de ejecución.
- [ ] Se verificó que documentación y configuración real coinciden.
