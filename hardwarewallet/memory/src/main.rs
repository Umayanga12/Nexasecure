mod token;

use wrap::Filter;
use token::{Database, Token};
use std::convert::Infallible;  

async fn main(){
    let db = Database::new();

    let db = warp::any().map(move || db.clone());

    let create_token = warp::path!("tokens")
        .and(warp::post())
        .and(warp::body::json())
        .and(db.clone())
        .and_then(create_token_handler);

    let get_token = warp::path!("tokens" / String)
        .and(warp::get())
        .and(db.clone())
        .and_then(get_token_handler);

    let update_token = warp::path!("tokens" / String)
        .and(warp::put())
        .and(warp::body::json())
        .and(db.clone())
        .and_then(update_token_handler);

    let delete_token = warp::path!("tokens" / String)
        .and(warp::delete())
        .and(db.clone())
        .and_then(delete_token_handler);

    let routes = list_tokens
        .or(create_token)
        .or(get_token)
        .or(update_token)
        .or(delete_token);

    warp::serve(routes).run(([127.0.0.1], 3030)).await;
}

async fn create_token_handler(new_token: Token, db: Database) -> Result<impl warp::Reply, Infallible> {
    let mut tokens = db.tokens.write().await;
    let id = Uuid::new_v4().to_string();
    let token = Token {
        id: id.clone(),
        value: new_token.value,
    };
    tokens.insert(id.clone(), token.clone());
    Ok(warp::reply::json(&token))
}

async fn get_token_handler(id: String, db: Database) -> Result<impl warp::Reply, warp::Rejection> {
    let tokens = db.tokens.read().await;
    if let Some(token) = tokens.get(&id) {
        Ok(warp::reply::json(&token))
    } else {
        Err(warp::reject::not_found())
    }
}

async fn update_token_handler(id: String, updated_token: Token, db: Database) -> Result<impl warp::Reply, warp::Rejection> {
    let mut tokens = db.tokens.write().await;
    if tokens.contains_key(&id) {
        let token = Token {
            id: id.clone(),
            value: updated_token.value,
        };
        tokens.insert(id.clone(), token.clone());
        Ok(warp::reply::json(&token))
    } else {
        Err(warp::reject::not_found())
    }
}

async fn delete_token_handler(id: String, db: Database) -> Result<impl warp::Reply, warp::Rejection> {
    let mut tokens = db.tokens.write().await;
    if tokens.remove(&id).is_some() {
        Ok(warp::reply::with_status("Deleted", warp::http::StatusCode::NO_CONTENT))
    } else {
        Err(warp::reject::not_found())
    }
}