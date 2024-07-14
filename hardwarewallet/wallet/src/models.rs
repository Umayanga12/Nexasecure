use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Account {
    pub account_address: String,
    pub private_key: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OwnershipToken {
    pub id: String,
    pub owner: String,
    pub metadata: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AddTokenPayload {
    pub account: Account,
    pub token: OwnershipToken,
}


#[derive(Serialize, Deserialize)]
pub struct SignTokenPayload {
    pub account_address: String,
    pub private_key: String,
    pub token_id: String,
}

#[derive(Serialize, Deserialize)]
pub struct SignedTokenResponse {
    pub token_id: String,
    pub signature: String,
}

#[derive(Deserialize)]
pub struct RemoveTokenPayload {
    pub token_id: String,
}

#[derive(Deserialize)]
pub struct RemoveAccountPayload {
    pub account_address: String,
}