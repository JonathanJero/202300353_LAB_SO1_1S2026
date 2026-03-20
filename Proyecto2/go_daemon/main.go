package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type RamInfo struct {
	TotalMB float64 `json:"total_mb"`
	LibreMB float64 `json:"libre_mb"`
	EnUsoMB float64 `json:"en_uso_mb"`
}

type Proceso struct {
	PID           int     `json:"pid"`
	Nombre        string  `json:"nombre"`
	VszKb         float64 `json:"vsz_kb"`
	RssKb         float64 `json:"rss_kb"`
	MemPorcentaje float64 `json:"mem_porcentaje"`
	CpuTicks      float64 `json:"cpu_ticks"`
}

type KernelData struct {
	Ram      RamInfo   `json:"ram"`
	Procesos []Proceso `json:"procesos"`
}

type Contenedor struct {
	ID      string
	Command string
	EsAlto  bool
}

var ctx = context.Background()
var totalEliminados int64 = 0

func inicializarInfraestructura() {
	fmt.Println("[*] 1. Verificando base de datos Valkey y Grafana...")
	exec.Command("sh", "-c", "cd ../db && sudo docker compose up -d").Run()
	
	fmt.Println("[*] 2. Verificando/Cargando Sonda de Kernel (Módulo C)...")
	// Solo hace insmod si lsmod no encuentra el modulo cargado
	exec.Command("sh", "-c", "lsmod | grep -q modulo || (cd ../kernel_module && sudo insmod modulo.ko)").Run()
	
	fmt.Println("[*] 3. Verificando/Programando Cronjob generador...")
	path_script, _ := os.Getwd()
	cron_cmd := fmt.Sprintf("crontab -l 2>/dev/null | grep -q 'generador.sh' || (crontab -l 2>/dev/null; echo \"*/2 * * * * %s/../cronjob/generador.sh\") | crontab -", path_script)
	exec.Command("sh", "-c", cron_cmd).Run()
	
	time.Sleep(3 * time.Second)
}

func limpiarInfraestructura() {
	fmt.Println("\n[!] Apagando daemon y limpiando sistema (Cronjob y Kernel)...")
	exec.Command("sh", "-c", "crontab -l | grep -v 'generador.sh' | crontab -").Run()
	// Solo hace rmmod si el modulo esta cargado
	exec.Command("sh", "-c", "lsmod | grep -q modulo && cd ../kernel_module && sudo rmmod modulo").Run()
}

func main() {
	fmt.Println("=== Iniciando Daemon Gestor ===")
	inicializarInfraestructura()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Error conectando a Valkey: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		limpiarInfraestructura()
		os.Exit(0)
	}()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	
	fmt.Println("[*] Infraestructura lista. Iniciando ciclo de monitoreo (cada 20s)...")
	
	for {
		<-ticker.C
		fmt.Printf("[%s] Ejecutando análisis de contenedores...\n", time.Now().Format("15:04:05"))
		ejecutarCicloAnalisis(rdb)
	}
}

func ejecutarCicloAnalisis(rdb *redis.Client) {
	data, err := os.ReadFile("/proc/continfo_pr2_so1_202300353")
	if err != nil {
		fmt.Printf("[!] Esperando a que el módulo genere datos en /proc...\n")
		return
	}
	
	var kernelData KernelData
	json.Unmarshal(data, &kernelData)

	rdb.Set(ctx, "sys_ram_total", kernelData.Ram.TotalMB, 0)
	rdb.Set(ctx, "sys_ram_libre", kernelData.Ram.LibreMB, 0)
	rdb.Set(ctx, "sys_ram_uso", kernelData.Ram.EnUsoMB, 0)
	rdb.Set(ctx, "sys_contenedores_eliminados", totalEliminados, 0)

	sort.Slice(kernelData.Procesos, func(i, j int) bool { return kernelData.Procesos[i].RssKb > kernelData.Procesos[j].RssKb })
	rdb.Del(ctx, "top5_ram")
	for i := 0; i < 5 && i < len(kernelData.Procesos); i++ {
		p := kernelData.Procesos[i]
		rdb.HSet(ctx, "top5_ram", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.RssKb)
	}

	sort.Slice(kernelData.Procesos, func(i, j int) bool { return kernelData.Procesos[i].CpuTicks > kernelData.Procesos[j].CpuTicks })
	rdb.Del(ctx, "top5_cpu")
	for i := 0; i < 5 && i < len(kernelData.Procesos); i++ {
		p := kernelData.Procesos[i]
		rdb.HSet(ctx, "top5_cpu", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.CpuTicks)
	}

	gestionarContenedores(rdb)
}

func gestionarContenedores(rdb *redis.Client) {
	out, _ := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Command}}").Output()
	lineas := strings.Split(string(out), "\n")
	var altos, bajos []Contenedor

	for _, linea := range lineas {
		if linea == "" || strings.Contains(linea, "grafana") || strings.Contains(linea, "valkey") {
			continue
		}
		partes := strings.Split(linea, "|")
		esAlto := strings.Contains(partes[1], "awk") || strings.Contains(partes[1], "bc")
		c := Contenedor{ID: partes[0], Command: partes[1], EsAlto: esAlto}
		
		if esAlto {
			altos = append(altos, c)
		} else {
			bajos = append(bajos, c)
		}
	}

	eliminarExceso(altos, 2, "Alto Consumo")
	eliminarExceso(bajos, 3, "Bajo Consumo")
}

func eliminarExceso(lista []Contenedor, limite int, tipo string) {
	if len(lista) > limite {
		exceso := len(lista) - limite
		for i := 0; i < exceso; i++ {
			exec.Command("docker", "rm", "-f", lista[i].ID).Run()
			totalEliminados++
			fmt.Printf("   -> [X] %s excedido. Contenedor eliminado: %s\n", tipo, lista[i].ID)
		}
	}
}
