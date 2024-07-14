use actix_web::{web, HttpResponse, Responder};
use hmac::{Hmac, Mac};
use serde_json::json;
use sha2::Sha256;
use std::sync::{Mutex, Arc};
use base64::encode;
use crate::models::{Account, AddTokenPayload, OwnershipToken, SignTokenPayload, SignedTokenResponse,RemoveTokenPayload,RemoveAccountPayload};
use crate::storage::Storage;
use log::{error};
type TokenPool = Mutex<Vec<OwnershipToken>>;

pub async fn add_account(
    account_storage: web::Data<Arc<Mutex<Storage>>>,
    payload: web::Json<Account>
) -> impl Responder {
    let account = payload.into_inner();
    let mut storage = match account_storage.lock() {
        Ok(storage) => storage,
        Err(_) => {
            error!("Failed to acquire storage lock due to poisoning");
            return HttpResponse::InternalServerError().json(json!({ "error": "Failed to acquire storage lock due to poisoning" }));
        }
    };

    if storage.add_account(&account.account_address, &account.private_key) {
        HttpResponse::Ok().json(json!({ "message": "Account added successfully" }))
    } else {
        error!("Account already exists or failed to save to file");
        HttpResponse::BadRequest().json(json!({ "error": "Account already exists or failed to save to file" }))
    }
}


pub async fn add_token(
    payload: web::Json<AddTokenPayload>,
    token_pool: web::Data<TokenPool>,
    account_storage: web::Data<Arc<Mutex<Storage>>>,
) -> impl Responder {
    let AddTokenPayload { account, token } = payload.into_inner();

    let storage = account_storage.lock().unwrap();
    if !storage.verify_account(&account.account_address, &account.private_key) {
        return HttpResponse::BadRequest().json(json!({
            "error": "Invalid account or private key."
        }));
    }

    if account.account_address != token.owner {
        return HttpResponse::BadRequest().json(json!({
            "error": "Account address does not match NFT owner."
        }));
    }

    let mut token_storage = token_pool.lock().unwrap();
    token_storage.push(token);

    HttpResponse::Ok().json(json!({ "message": "NFT added successfully" }))
}

pub async fn get_accounts(account_storage: web::Data<Arc<Mutex<Storage>>>) -> impl Responder {
    let storage = match account_storage.lock() {
        Ok(storage) => storage,
        Err(_) => {
            eprintln!("Failed to acquire storage lock due to poisoning");
            return HttpResponse::InternalServerError().json(json!({ "error": "Failed to acquire storage lock due to poisoning" }));
        }
    };
    
    let accounts = storage.get_accounts();
    
    if accounts.is_empty() {
        HttpResponse::Ok().json(json!({ "message": "No accounts found" }))
    } else {
        HttpResponse::Ok().json(accounts)
    }
}

pub async fn sign_token(
    payload: web::Json<SignTokenPayload>,
    account_storage: web::Data<Arc<Mutex<Storage>>>,
    token_pool: web::Data<TokenPool>,
) -> impl Responder {
    let SignTokenPayload { account_address, private_key, token_id } = payload.into_inner();

    let storage = account_storage.lock().unwrap();
    if !storage.verify_account(&account_address, &private_key) {
        return HttpResponse::BadRequest().json(json!({
            "error": "Invalid account or private key."
        }));
    }

    let token_storage = token_pool.lock().unwrap();
    if let Some(token) = token_storage.iter().find(|n| n.id == token_id) {
        if token.owner != account_address {
            return HttpResponse::BadRequest().json(json!({
                "error": "Account is not the owner of the NFT."
            }));
        }

        // Sign the token
        let mut mac = Hmac::<Sha256>::new_from_slice(private_key.as_bytes()).unwrap();
        mac.update(token_id.to_string().as_bytes());
        let signature = encode(mac.finalize().into_bytes());

        let response = SignedTokenResponse {
            token_id,
            signature,
        };

        return HttpResponse::Ok().json(response);
    }

    HttpResponse::NotFound().json(json!({ "error": "NFT not found" }))
}

pub async fn get_tokens(
    token_pool: web::Data<Arc<Mutex<Storage>>>
) -> impl Responder {
    let storage = token_pool.lock();
    match storage {
        Ok(storage) => {
            let tokens = storage.get_tokens();
            HttpResponse::Ok().json(tokens)
        },
        Err(_) => {
            HttpResponse::InternalServerError().json(json!({ "error": "Failed to acquire storage lock due to poisoning" }))
        }
    }
}

pub async fn remove_token(
    token_pool: web::Data<Arc<Mutex<Storage>>>,
    payload: web::Json<RemoveTokenPayload>
) -> impl Responder {
    let remove_payload = payload.into_inner();

    let storage = token_pool.lock();
    match storage {
        Ok(mut storage) => {
            if storage.remove_token(&remove_payload.token_id) {
                HttpResponse::Ok().json(json!({ "message": "Token removed successfully" }))
            } else {
                HttpResponse::BadRequest().json(json!({ "error": "Token not found" }))
            }
        },
        Err(_) => {
            HttpResponse::InternalServerError().json(json!({ "error": "Failed to acquire storage lock due to poisoning" }))
        }
    }
}

pub async fn remove_account(
    account_storage: web::Data<Arc<Mutex<Storage>>>,
    payload: web::Json<RemoveAccountPayload>,
) -> impl Responder {
    let account_address = payload.into_inner().account_address;
    let mut storage = match account_storage.lock() {
        Ok(storage) => storage,
        Err(_) => {
            eprintln!("Failed to acquire storage lock due to poisoning");
            return HttpResponse::InternalServerError().json(json!({ "error": "Failed to acquire storage lock due to poisoning" }));
        }
    };

    if storage.remove_account(&account_address) {
        HttpResponse::Ok().json(json!({ "message": "Account removed successfully" }))
    } else {
        HttpResponse::BadRequest().json(json!({ "error": "Account not found" }))
    }
}