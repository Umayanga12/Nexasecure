use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{self, BufReader, Write};
use serde::{Deserialize, Serialize};
use bcrypt::{hash, verify, DEFAULT_COST};

#[derive(Serialize, Deserialize, Default, Debug)]
pub struct Storage {
    accounts: HashMap<String, String>, // account_address -> hashed_private_key
    file_path: String,
}

impl Storage {
    pub fn new(file_path: &str) -> Self {
        let accounts = match File::open(file_path) {
            Ok(file) => {
                let reader = BufReader::new(file);
                serde_json::from_reader(reader).unwrap_or_default()
            }
            Err(_) => HashMap::new(),
        };

        Storage {
            accounts,
            file_path: file_path.to_string(),
        }
    }

    pub fn save_to_file(&self) -> io::Result<()> {
        let json_data = serde_json::to_string(&self.accounts)?;
        let mut file = File::create(&self.file_path)?;
        file.write_all(json_data.as_bytes())?;
        Ok(())
    }

    pub fn add_account(&mut self, account_address: &str, private_key: &str) -> bool {
        if self.accounts.contains_key(account_address) {
            return false; // Account already exists
        }
        let hashed_key = hash(private_key, DEFAULT_COST).unwrap();
        self.accounts.insert(account_address.to_string(), hashed_key);
        self.save_to_file().unwrap();
        true
    }

    pub fn verify_account(&self, account_address: &str, private_key: &str) -> bool {
        match self.accounts.get(account_address) {
            Some(hashed_key) => verify(private_key, hashed_key).unwrap_or(false),
            None => false,
        }
    }

    pub fn get_accounts(&self) -> HashMap<String, String> {
        self.accounts.clone()
    }
}
