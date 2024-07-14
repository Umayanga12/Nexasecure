use actix_web::{web, App, HttpServer};
use std::sync::{Arc, Mutex};
use env_logger::Env;

mod handlers;
mod models;
mod storage;

use handlers::{add_account, add_token, get_accounts, sign_token,get_tokens,remove_token,remove_account};
use models::OwnershipToken;
use storage::Storage;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize the logger
    env_logger::init_from_env(Env::default().default_filter_or("debug"));

    let token_pool = web::Data::new(Mutex::new(Vec::<OwnershipToken>::new()));
    let account_storage = web::Data::new(Arc::new(Mutex::new(Storage::default())));

    HttpServer::new(move || {
        App::new()
            .app_data(account_storage.clone())
            .app_data(token_pool.clone())
            .service(web::resource("/add_account").route(web::post().to(add_account)))
            .service(web::resource("/add_token").route(web::post().to(add_token)))
            .service(web::resource("/get_accounts").route(web::get().to(get_accounts)))
            .service(web::resource("/sign_token").route(web::post().to(sign_token)))
            .service(web::resource("/get_tokens").route(web::get().to(get_tokens)))
            .service(web::resource("/remove_token").route(web::post().to(remove_token)))
            .service(web::resource("/remove_account").route(web::delete().to(remove_account)))
    })
    .bind("127.0.0.1:3030")?
    .run()
    .await
}
