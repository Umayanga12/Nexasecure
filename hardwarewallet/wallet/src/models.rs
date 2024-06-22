use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Account {
    pub account_address: String,
    pub private_key: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OwnershipToken {
    pub id: u64,
    pub owner: String,
    pub metadata: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TransferToken {
    pub token_id: u64,
    pub new_owner: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TransferTokenPayload {
    pub account: Account,
    pub transfer: TransferToken,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AddTokenPayload {
    pub account: Account,
    pub token: OwnershipToken,
}

