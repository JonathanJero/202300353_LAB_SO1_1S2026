use actix_web::{post, web, App, HttpServer, HttpResponse, Responder};
use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct WarReport {
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: String,
}

#[post("/grpc-202300353")]
async fn handle_report(item: web::Json<WarReport>) -> impl Responder {
    let url = std::env::var("GO_INGEST_URL").unwrap_or_else(|_| "http://go-ingest.mumnk8s.svc.cluster.local:8081/ingest".to_string());
    
    let client = Client::new();
    match client.post(&url).json(&item.0).send().await {
        Ok(res) if res.status().is_success() => {
            HttpResponse::Ok().json(serde_json::json!({"status": "M.U.M.N.K8s API Rust - Forwarded"}))
        }
        Ok(res) => {
            HttpResponse::InternalServerError().json(serde_json::json!({"error": format!("Go Ingest error: {}", res.status())}))
        }
        Err(e) => {
            HttpResponse::InternalServerError().json(serde_json::json!({"error": format!("Forward error: {}", e)}))
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Rust API Gateway listening on port 8080...");
    HttpServer::new(|| {
        App::new().service(handle_report)
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
