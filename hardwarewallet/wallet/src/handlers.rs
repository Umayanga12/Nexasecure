use actix_web::{web, HttpResponse, Responder};
use serde_json::json;
use std::sync::{Mutex, Arc};

use crate::models::{Account, OwnershipToken, AddTokenPayload, TransferTokenPayload};
use crate::storage::Storage;

type TokenPool = Mutex<Vec<OwnershipToken>>;

pub async fn add_account(
    account_storage: web::Data<Arc<Mutex<Storage>>>,
    payload: web::Json<Account>
) -> impl Responder {
    let account = payload.into_inner();
    let mut storage = account_storage.lock().unwrap();
    if storage.add_account(&account.account_address, &account.private_key) {
        HttpResponse::Ok().json(json!({ "message": "Account added successfully" }))
    } else {
        HttpResponse::BadRequest().json(json!({ "error": "Account already exists" }))
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

pub async fn transfer_token(
    payload: web::Json<TransferTokenPayload>,
    token_pool: web::Data<TokenPool>,
    account_storage: web::Data<Arc<Mutex<Storage>>>,
) -> impl Responder {
    let TransferTokenPayload { account, transfer } = payload.into_inner();

    let storage = account_storage.lock().unwrap();
    if !storage.verify_account(&account.account_address, &account.private_key) {
        return HttpResponse::BadRequest().json(json!({
            "error": "Invalid account or private key."
        }));
    }

    let mut token_storage = token_pool.lock().unwrap();
    if let Some(token) = token_storage.iter_mut().find(|n| n.id == transfer.token_id) {
        if token.owner != account.account_address {
            return HttpResponse::BadRequest().json(json!({
                "error": "Account is not the owner of the NFT."
            }));
        }

        token.owner = transfer.new_owner.clone();

        return HttpResponse::Ok().json(json!({ "message": "NFT ownership transferred successfully" }));
    }

    HttpResponse::NotFound().json(json!({ "error": "NFT not found" }))
}

pub async fn get_accounts(account_storage: web::Data<Arc<Mutex<Storage>>>) -> impl Responder {
    let storage = account_storage.lock().unwrap();
    let accounts = storage.get_accounts();
    HttpResponse::Ok().json(accounts)
}
