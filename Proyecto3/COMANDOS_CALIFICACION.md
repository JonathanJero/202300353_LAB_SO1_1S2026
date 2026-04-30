# Guía Exacta de Comandos para tu Calificación (Copy/Paste)

Esta guía contiene **exactamente** los comandos que debes copiar y pegar en tu terminal durante la calificación, adaptados específicamente a los nombres reales de tus Deployments y al *namespace* (`mumnk8s`) que usaste en tu proyecto. 

Con estos comandos no tendrás ningún error de "recurso no encontrado" como pasó antes.

---

## 1. Zot (Fuera del Clúster)

El auxiliar pedirá que le muestres la VM externa y que verifique el catálogo de imágenes.

```bash
# Conectarse a la máquina temporalmente y mostrar que Zot está corriendo en un solo paso
# (Este comando auto-genera tus llaves SSH si no las tienes y evita errores de Permission Denied)
gcloud compute ssh mumnk8s-zot-vm --zone=us-central1-f --command="sudo docker ps" --quiet

# El comando maestro: Mostrar que Zot tiene las imágenes (puedes correr esto desde cualquier terminal)
curl.exe http://35.238.80.248:5000/v2/_catalog
```
**Salida esperada:** Un JSON con `go-consumer`, `go-ingest`, `go-writer` y `rust-api`.

> 🌍 **LINK DE ZOT:** Puedes abrir en tu navegador: `http://35.238.80.248:5000/v2/_catalog`

---

## 2. Gateway API y Rutas

El auxiliar verificará que no usaste `Ingress` sino `Gateway API`.

```bash
# Verificar el Gateway aprovisionado por Google
kubectl get gateway -n mumnk8s

# Verificar la ruta HTTP con tu carnet
kubectl get httproute -n mumnk8s
```
**Salida esperada:** Un Gateway con IP asignada (`35.186.255.180`) y Programmed=True. Una ruta `mumnk8s-routes`.

> 🌍 **LINK DE GATEWAY:** La IP pública donde enviarás tráfico en Locust es `http://35.186.255.180` (endpoint `/grpc-202300353`).

---

## 3. Lógica de Aplicación (Rust & Go)

El auxiliar verificará tus pods, el autoescalado y los logs de comunicación.

```bash
# 1. Verificar el Autoescalado (HPA) apuntando a 30% de CPU
kubectl get hpa -n mumnk8s

# 2. Listar todo (Pods, Servicios, etc.)
kubectl get all -n mumnk8s

# 3. Mostrar que Rust recibe el tráfico de Locust (Haz esto mientras Locust manda carga)
kubectl logs deployment/rust-api -n mumnk8s --tail=10

# 4. Mostrar que Go recibe los mensajes binarios gRPC
kubectl logs deployment/go-writer -n mumnk8s --tail=10
```

---

## 4. Máquinas Virtuales (KubeVirt) - Anti-Fraude

Este es el paso crítico donde validan que usaste KubeVirt en lugar de Deployments normales.

```bash
# 1. Listar las Definiciones de VM
kubectl get vms -n mumnk8s

# 2. Listar las Instancias Corriendo (VMI)
kubectl get vmi -n mumnk8s

# 3. Mostrar el volumen o disco de CloudInit
kubectl describe vmi valkey-vm -n mumnk8s | Select-String "CloudInit"
```

*(Si el auxiliar pide usar `virtctl console`, tendrías que descargar `virtctl` primero. Por lo general con ver el estado `Running` en el comando `get vmi` y el volúmen de disco `CloudInit` es suficiente para validar KubeVirt).*

---

## 5. Mensajería (RabbitMQ y Consumidor)

El auxiliar querrá ver la cola "viva" almacenando mensajes.

```bash
# 1. Listar la cola dentro del pod de RabbitMQ (Hazlo mientras mandas carga)
kubectl exec pod/rabbitmq-0 -n mumnk8s -- rabbitmqctl list_queues

# 2. Mostrar los logs del consumidor guardando en base de datos
kubectl logs deployment/go-consumer -n mumnk8s --tail=10
```

---

## 6. Grafana (Dashboard)

El auxiliar verificará que los datos en vivo se visualicen en Grafana. Tienes listo tu archivo `dashboard_grafana.json` para importar en tu PC.

> 🌍 **LINK DE GRAFANA:** Abre tu navegador y entra a: `http://136.119.0.206:3000`
> - **Usuario/Contraseña inicial:** `admin` / `admin`
> - **Data Source a agregar:** Redis (URL: `redis://34.118.228.119:6379`)
> - **Importar Dashboard:** Presiona el botón "Import" y arrastra el archivo `dashboard_grafana.json` generado.

---

## 7. Pruebas de Carga (Locust)

Como ya descubrimos que Windows no te dejaba correr Python directamente, tu comando infalible para iniciar Locust con Docker es este:

```bash
# Ejecutar Locust
docker run -p 8089:8089 -v "${PWD}/locust:/mnt/locust" locustio/locust -f /mnt/locust/locustfile.py
```

> 🌍 **LINK DE LOCUST:** Abre tu navegador y entra a: `http://localhost:8089`
> - **Target Host:** `http://35.186.255.180`
> - **Usuarios/RPS sugeridos para prueba exitosa:** 100 usuarios, 10 RPS.

---
*Fin de la guía de comandos. Ten este archivo a la mano para solo copiar y pegar y que tu presentación fluya perfectamente.*
