use actix_web::{web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
use std::fs::{self, OpenOptions};
use std::io::{Write, Read};
use reqwest::Client;

#[derive(Serialize, Deserialize)]
struct WalletData {
    uuid: String,
    dip: String,
    token: String,
}

async fn create_file(data: web::Json<WalletData>) -> impl Responder {
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

async fn update_file(data: web::Json<WalletData>) -> impl Responder {
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

async fn read_file() -> impl Responder {
    let file_path = "data.json";
    let mut file = match OpenOptions::new().read(true).open(file_path) {
        Ok(file) => file,
        Err(_) => return HttpResponse::NotFound().body("File not found"),
    };

    let mut contents = String::new();
    if file.read_to_string(&mut contents).is_ok() {
        HttpResponse::Ok().body(contents)
    } else {
        HttpResponse::InternalServerError().body("Failed to read file")
    }
}

async fn external_api() -> impl Responder {
    let client = Client::new();
    let res = client.get("http://127.0.0.1:3030")
        .send()
        .await;

    match res {
        Ok(response) => {
            let body = response.text().await.unwrap();
            HttpResponse::Ok().body(body)
        },
        Err(_) => HttpResponse::InternalServerError().body("Error memory dose not responding"),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/create", web::post().to(create_file))
            .route("/update", web::put().to(update_file))
            .route("/delete", web::delete().to(delete_file))
            .route("/read", web::get().to(read_file))
            .route("/external-api", web::get().to(external_api))
    })
    .bind("127.0.0.1:3031")?
    .run()
    .await
}
