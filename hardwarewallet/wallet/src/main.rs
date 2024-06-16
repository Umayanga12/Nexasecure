use actix_web::{web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
use std::fs::{self, OpenOptions};
use std::io::Write;
use reqwest::Client;

#[derive(Serialize, Deserialize)]
struct walletdata {
    key: String,
    value: String,
}

async fn create_file(data: web::Json<MyData>) -> impl Responder {
    let file_path = "data.json";
    let json_data = serde_json::to_string(&data.into_inner()).unwrap();
    let mut file = OpenOptions::new()
        .write(true)
        .create(true)
        .open(file_path)
        .unwrap();
    file.write_all(json_data.as_bytes()).unwrap();
    HttpResponse::Ok().body("File created")
}

async fn update_file(data: web::Json<MyData>) -> impl Responder {
    let file_path = "data.json";
    let json_data = serde_json::to_string(&data.into_inner()).unwrap();
    let mut file = OpenOptions::new()
        .write(true)
        .truncate(true)
        .open(file_path)
        .unwrap();
    file.write_all(json_data.as_bytes()).unwrap();
    HttpResponse::Ok().body("File updated")
}

async fn delete_file() -> impl Responder {
    let file_path = "data.json";
    if fs::remove_file(file_path).is_ok() {
        HttpResponse::Ok().body("File deleted")
    } else {
        HttpResponse::NotFound().body("File not found")
    }
}

async fn external_api() -> impl Responder {
    let client = Client::new();
    let res = client.get("https://api.example.com/data")
        .send()
        .await;

    match res {
        Ok(response) => {
            let body = response.text().await.unwrap();
            HttpResponse::Ok().body(body)
        },
        Err(_) => HttpResponse::InternalServerError().body("Failed to communicate with the API"),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/create", web::post().to(create_file))
            .route("/update", web::put().to(update_file))
            .route("/delete", web::delete().to(delete_file))
            .route("/external-api", web::get().to(external_api))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
