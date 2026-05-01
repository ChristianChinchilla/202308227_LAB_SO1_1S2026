use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use tonic::Request;

// 1. Importamos el código gRPC generado automáticamente por build.rs
pub mod wartweets {
    tonic::include_proto!("wartweets");
}

use wartweets::war_report_service_client::WarReportServiceClient;
use wartweets::WarReportRequest;

// Estructura para recibir el JSON de Locust
#[derive(Serialize, Deserialize)]
struct WarReport {
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: String,
}

#[post("/report")]
async fn receive_report(report: web::Json<WarReport>) -> impl Responder {
    println!("Recibido de Locust: {}. Enviando a Go vía gRPC...", report.country);

    // 2. Conexión al servidor de Go
    // "api-go-service" es el nombre que usaremos en el Deployment de Kubernetes
    let mut client = match WarReportServiceClient::connect("http://api-go-service:50051").await {
        Ok(c) => c,
        Err(e) => {
            eprintln!("Error conectando a Go: {:?}", e);
            return HttpResponse::InternalServerError().body("No se pudo conectar con el servidor Go");
        }
    };

    // 3. Empaquetamos los datos para gRPC
    let request = Request::new(WarReportRequest {
        country: report.country.clone(),
        warplanes_in_air: report.warplanes_in_air,
        warships_in_water: report.warships_in_water,
        timestamp: report.timestamp.clone(),
    });

    // 4. Enviamos el reporte
    match client.send_report(request).await {
        Ok(response) => {
            let status = response.into_inner().status;
            println!("Go respondió: {}", status);
            HttpResponse::Ok().body(status)
        },
        Err(e) => {
            eprintln!("Error gRPC: {:?}", e);
            HttpResponse::InternalServerError().body("Error en la comunicación gRPC")
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("API Rust (Productor gRPC) iniciada en el puerto 8080");

    HttpServer::new(|| {
        App::new()
            .service(receive_report)
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
