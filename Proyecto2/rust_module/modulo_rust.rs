//! Módulo de Kernel en Rust para Puntos Extra SO1

use kernel::prelude::*;

// Definición de los metadatos del módulo
module! {
    type: HolaMundo,
    name: "modulo_rust",
    author: "Jonathan Jeronimo 202300353",
    description: "Modulo extra en Rust para SO1",
    license: "GPL",
}

// Estructura principal de nuestro módulo
struct HolaMundo;

// Implementación del trait Module de Linux
impl kernel::Module for HolaMundo {
    fn init(_name: &'static CStr, _module: &'static ThisModule) -> Result<Self> {
        // pr_info! es el equivalente a printk(KERN_INFO ...)
        pr_info!("Hola Mundo 202300353\n");
        Ok(HolaMundo)
    }
}

// El trait Drop se ejecuta automáticamente al hacer rmmod (limpieza)
impl Drop for HolaMundo {
    fn drop(&mut self) {
        pr_info!("Modulo de Rust 202300353 descargado exitosamente\n");
    }
}
