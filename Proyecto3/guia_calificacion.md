# Guía de Calificación: Proyecto M.U.M.N.K8s (SO1)

---

## 1. Infraestructura Base y Zot (Fuera del Clúster)

### Criterio 1.4: Despliegue de Zot en una VM externa y uso de HTTPS.

| Acción                                  | Comando                               | Salida Esperada                                                                 |
|----------------------------------------|---------------------------------------|---------------------------------------------------------------------------------|
| Verificar Zot                          | ssh <usuario>@<ip-externa-zot>         | Acceso exitoso a la VM fuera de GKE.                                            |
| Verificar Proceso                      | docker ps <br> o <br> systemctl status zot | Un contenedor o servicio de Zot corriendo en el puerto 443/5000.               |
| Verificar subida de imágenes           | curl http://ip-zot:5000/v2/_catalog    | JSON con la lista de imágenes (rust-api, go-server, etc.).                      |

---

## 2. Orquestación y Red (GKE & Gateway API)

### Criterio 1.1: Implementación de Gateway API.

| Acción            | Comando                     | Salida Esperada                                                                 |
|------------------|-----------------------------|---------------------------------------------------------------------------------|
| Verificar Gateway| kubectl get gateway         | Un recurso de tipo Gateway con una IP externa asignada.                         |
| Verificar Rutas  | kubectl get httproute       | Rutas configuradas para /grpc-#carnet y opcionalmente /dapr-#carnet.            |

---

## 3. Lógica de Aplicación (Rust & Go)

### Criterios 1.2 y 1.3: API Rust y Servicios Go.
| Acción            | Comando                                   | Salida Esperada                                                                 |
|-------------------|-------------------------------------------|---------------------------------------------------------------------------------|
| Verificar HPA     | kubectl get hpa                           | El HPA de la API Rust con targets de CPU > 30%.                                 |
| Listar Todo       | kubectl get all                           | Listar Todo                                                                     |
| Logs de Ingesta   | kubectl logs deployment/api-rust          | Logs mostrando la recepción de JSON desde Locust.                               |
| Logs gRPC         | kubectl logs deployment/go-grpc-server    | Logs indicando que se recibió un mensaje vía gRPC desde el cliente Go.          |
---

## 4. Verificación de KubeVirt (Valkey & Grafana)

### Criterios 1.6 y 1.7: Punto crítico para evitar el fraude.

| Acción            | Comando                           | Salida Esperada                                                                 |
|------------------|-----------------------------------|---------------------------------------------------------------------------------|
| Listar VMs       | kubectl get vms                   | Al menos dos VirtualMachines: valkey-vm y grafana-vm.                           |
| Verificar Estado | kubectl get vmi                   | El estado debe ser Running.                                                     |
| Revisar Recursos | kubectl describe vmi <nombre-vm>  | Debería mostrar la imagen de disco (CloudInit o ContainerDisk).                 |

### Comandos de validación avanzada:

- Acceso a Consola:
  ```
  virtctl console <nombre-vmi>
  ```

- Persistencia Valkey:
  ```
  valkey-cli info persistence
  ```

---

## 5. Mensajería (RabbitMQ)

### Criterio 1.5: Broker y Consumidor.

| Acción             | Comando                                                        | Salida Esperada                                           |
|-------------------|----------------------------------------------------------------|-----------------------------------------------------------|
| Verificar Colas   | kubectl exec -it <pod-rabbitmq> -- rabbitmqctl list_queues     | Cola con mensajes en tiempo real, o mediante interfaz gráfica |
| Logs Consumidor   | kubectl logs deployment/rabbitmq-consumer                     | "Mensaje procesado y guardado en Valkey".                |

---

## 6. Pruebas de Carga (Locust)

### Criterio 2.1: Generación de tráfico.

| Acción            | Comando                  | Salida Esperada                           |
|------------------|--------------------------|-------------------------------------------|
| Interfaz Locust  | Abrir IP en navegador    | Gráficas de RPS y fallos en 0%.           |

---

## 7. Grafana

Acceder a Grafana por medio de NodePort, ip-forward o IP externa y mostrar el dashboard.  
De preferencia, que la base de datos no tenga datos para que el dashboard esté vacío al momento de ser mostrado.

---

## 8. Mensajería Alternativa (Dapr - Punteo Extra)

Esta sección valida la implementación del flujo adicional de mensajería utilizando el SDK de Dapr y su integración con RabbitMQ.

| Acción                  | Comando                         | Salida Esperada                                                                 |
|------------------------|---------------------------------|---------------------------------------------------------------------------------|
| Verificar Ruta Dapr    | kubectl get httproute           | Debe aparecer una ruta configurada para /dapr-#carnet vinculada al Gateway.     |
| Verificar Sidecars     | kubectl get pods                | Los pods deben mostrar 2/2 contenedores listos (App + Dapr sidecar).            |
| Componentes Dapr       | kubectl get components          | Un recurso pubsub.rabbitmq en estado "Ready".                                   |
| Verificar Suscripción  | kubectl get subscriptions       | Suscripción al tópico de reportes militares (si es declarativa).                |
| Logs de Dapr           | kubectl logs <pod-name> -c daprd| Logs indicando conexión exitosa al componente pubsub de RabbitMQ.              |

---

## ⚠️ Advertencia

**SI FALLA ALGUNA DE ESTAS INDICACIONES SE DARÁ POR HECHO QUE USTED NO REALIZÓ EL PROYECTO Y NO TIENE CONOCIMIENTO DEL MISMO, POR LO QUE AUTOMÁTICAMENTE TENDRÁ 0 PTS.**