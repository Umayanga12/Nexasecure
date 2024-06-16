use serde::{Deserialize, Serialize};
user std::colloctions::HashMap;
use uuid::Uuid;
use toki::sync::RwLock;
use std::sync::Arc;

pub struct Token {
   pub user_id: Uuid,
   pub token: String,
}

pub struct Database{
    pub tokens: Arc<RwLock<HashMap<String,Token>>>
}


impl Database{
    pub fn new() -> self{
        Database{
            tokens: Arc::new(RwLock::new(HashMap::new())),
        }
    }
}