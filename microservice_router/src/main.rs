use actix_web::{web, App, HttpServer, HttpRequest, HttpResponse};
use actix_web::dev::Server;
use serde::Deserialize;
use std::collections::HashMap;
use reqwest::Client;
use log::info;
use env_logger;

#[derive(Deserialize, Clone)]
struct ServiceConfig {
    services: HashMap<String, String>,
}

async fn route_request(
    req: HttpRequest,
    body: web::Bytes,
    data: web::Data<ServiceConfig>,
) -> HttpResponse {
    let path = req.path();
    let client = Client::new();

    // Extract the first segment of the path as the service name
    let service_name = path.split('/').nth(1).unwrap_or("");

    if let Some(service_url) = data.services.get(service_name) {
        let url = format!("{}{}", service_url, req.uri().path_and_query().map(|x| x.as_str()).unwrap_or(""));
        let forwarded_req = client
            .request(req.method().clone(), &url)
            .headers(req.headers().clone().into())
            .body(body.to_vec())
            .send()
            .await;

        match forwarded_req {
            Ok(mut res) => {
                let mut client_response = HttpResponse::build(res.status());

                for (key, value) in res.headers().iter() {
                    client_response.insert_header((key.clone(), value.clone()));
                }

                client_response.body(res.bytes().await.unwrap_or_else(|_| web::Bytes::new()))
            }
            Err(_) => HttpResponse::InternalServerError().finish(),
        }
    } else {
        HttpResponse::NotFound().body(format!("Service '{}' not found", service_name))
    }
}

fn create_server(config: ServiceConfig) -> Server {
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(config.clone()))
            .default_service(web::route().to(route_request))
    })
    .bind("127.0.0.1:8080")
    .expect("Can not bind to port 8080")
    .run()
}

#[actix_web::main]
async fn main() {
    env_logger::init();
    
    // Define your microservice endpoints here
    let services = HashMap::from([
        (String::from("service1"), String::from("http://localhost:9001")),
        (String::from("service2"), String::from("http://localhost:9002")),
    ]);

    let config = ServiceConfig { services };

    info!("Starting server at http://127.0.0.1:8080");
    create_server(config).await.unwrap();
}
