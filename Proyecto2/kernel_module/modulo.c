#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/init.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>

// Metadatos del módulo
MODULE_LICENSE("GPL");
MODULE_AUTHOR("Jonathan Jeronimo 202300353");
MODULE_DESCRIPTION("Sonda de Kernel para Telemetria de RAM y Procesos - Proyecto 2 SO1");

// Definición del nombre del archivo en /proc según requerimientos
#define PROCFS_NAME "continfo_pr2_so1_202300353"

// Función principal que lee las estructuras del kernel y escribe en /proc
static int escribir_archivo(struct seq_file *m, void *v) {
    struct sysinfo i;
    struct task_struct *task;
    unsigned long total_ram, free_ram, used_ram;
    bool first_process = true;

    // 1. Obtención de métricas globales de memoria RAM
    si_meminfo(&i);
    // Conversión de páginas de memoria a Megabytes (MB)
    total_ram = (i.totalram * i.mem_unit) / (1024 * 1024);
    free_ram = (i.freeram * i.mem_unit) / (1024 * 1024);
    used_ram = total_ram - free_ram;

    // Depuración: Registro en el log del kernel de la lectura de RAM
    printk(KERN_DEBUG "[Sonda SO1] Lectura de RAM ejecutada. Total: %lu MB\n", total_ram);

    // Formateo del inicio del JSON con los datos de la RAM
    seq_printf(m, "{\n  \"ram\": {\n    \"total_mb\": %lu,\n    \"libre_mb\": %lu,\n    \"en_uso_mb\": %lu\n  },\n", total_ram, free_ram, used_ram);
    seq_printf(m, "  \"procesos\": [\n");

    // 2. Iteración sobre la lista de procesos activos usando la macro del kernel
    for_each_process(task) {
        unsigned long vsz = 0;
        unsigned long rss = 0;
        unsigned long mem_porcentaje = 0;
        unsigned long cpu_usage = 0;

        // Si el proceso tiene estructura de memoria (mm_struct) asignada, extraemos sus valores
        if (task->mm) {
            // Conversión de páginas a Kilobytes (KB) usando desplazamiento de bits (PAGE_SHIFT)
            vsz = task->mm->total_vm << (PAGE_SHIFT - 10);
            rss = get_mm_rss(task->mm) << (PAGE_SHIFT - 10);
            
            // Cálculo del porcentaje de memoria usada respecto al total
            if (total_ram > 0) {
                mem_porcentaje = (rss * 100) / (total_ram * 1024);
            }
        }
        
        // Cálculo de ticks de CPU (Tiempo en modo usuario + Tiempo en modo kernel)
        cpu_usage = task->utime + task->stime;

        // Control de comas para mantener la estructura JSON válida
        if (!first_process) {
            seq_printf(m, ",\n");
        }
        first_process = false;

        // Escritura de los datos del proceso en formato JSON
        seq_printf(m, "    {\n");
        seq_printf(m, "      \"pid\": %d,\n", task->pid);
        seq_printf(m, "      \"nombre\": \"%s\",\n", task->comm);
        seq_printf(m, "      \"vsz_kb\": %lu,\n", vsz);
        seq_printf(m, "      \"rss_kb\": %lu,\n", rss);
        seq_printf(m, "      \"mem_porcentaje\": %lu,\n", mem_porcentaje);
        seq_printf(m, "      \"cpu_ticks\": %lu\n", cpu_usage);
        seq_printf(m, "    }");
    }

    // Cierre de la estructura JSON
    seq_printf(m, "\n  ]\n}\n");
    return 0;
}

// Función que se ejecuta al abrir el archivo /proc
static int al_abrir(struct inode *inode, struct file *file) {
    return single_open(file, escribir_archivo, NULL);
}

// Definición de las operaciones permitidas sobre el archivo /proc
static const struct proc_ops operaciones = {
    .proc_open = al_abrir,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

// Función de inicialización: Se ejecuta al cargar el módulo con insmod
static int __init modulo_init(void) {
    // Creación del archivo en el sistema de archivos /proc
    proc_create(PROCFS_NAME, 0, NULL, &operaciones);
    // Uso de printk para notificar la carga exitosa (Requerimiento de depuración)
    printk(KERN_INFO "[Sonda SO1] Modulo cargado exitosamente. Archivo creado: /proc/%s\n", PROCFS_NAME);
    return 0;
}

// Función de limpieza: Se ejecuta al descargar el módulo con rmmod
static void __exit modulo_cleanup(void) {
    // Eliminación segura del archivo en /proc para evitar memory leaks o punteros colgantes
    remove_proc_entry(PROCFS_NAME, NULL);
    // Uso de printk para notificar la descarga exitosa (Requerimiento de depuración)
    printk(KERN_INFO "[Sonda SO1] Modulo removido exitosamente. Archivo eliminado: /proc/%s\n", PROCFS_NAME);
}

// Macros para registrar las funciones de inicio y salida en el kernel
module_init(modulo_init);
module_exit(modulo_cleanup);
