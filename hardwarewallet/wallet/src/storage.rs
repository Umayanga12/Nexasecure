use std::collections::{HashMap, HashSet};
use std::fs::{File, create_dir_all};
use std::io::{self, BufReader, Write};
use serde::{Deserialize, Serialize};
use bcrypt::{hash, verify, DEFAULT_COST};

#[derive(Serialize, Deserialize, Default, Debug)]
pub struct Storage {
    accounts: HashMap<String, String>, // account_address -> hashed_private_key
    tokens: HashSet<String>, // Set of tokens
    file_path: String,
}

impl Storage {
    pub fn new(file_path: &str) -> Self {
        // Ensure the directory exists
        if let Some(parent) = std::path::Path::new(file_path).parent() {
            if let Err(e) = create_dir_all(parent) {
                eprintln!("Failed to create directories: {}", e);
            }
        }

        let accounts = match File::open(file_path) {
            Ok(file) => {
                let reader = BufReader::new(file);
                serde_json::from_reader(reader).unwrap_or_default()
            }
            Err(e) => {
                eprintln!("Failed to open file: {}. Starting with an empty storage.", e);
                HashMap::new()
            },
        };

        Storage {
            accounts,
            tokens: HashSet::new(), // Initialize tokens
            file_path: file_path.to_string(),
        }
    }

    pub fn save_to_file(&self) -> io::Result<()> {
        if let Some(parent) = std::path::Path::new(&self.file_path).parent() {
            create_dir_all(parent)?; // Ensure the directory exists
        }
        let json_data = serde_json::to_string(&self.accounts)?;
        let mut file = File::create(&self.file_path)?;
        file.write_all(json_data.as_bytes())?;
        Ok(())
    }

    pub fn add_account(&mut self, account_address: &str, private_key: &str) -> bool{
        if self.accounts.contains_key(account_address) {
            return false; // Account already exists
        }
        let hashed_key = hash(private_key, DEFAULT_COST).unwrap();
        self.accounts.insert(account_address.to_string(), hashed_key);
        if let Err(e) = self.save_to_file() {
            eprintln!("Failed to save to file: {}", e); // Log the error
            return false;
        }
        true
    }

    pub fn verify_account(&self, account_address: &str, private_key: &str) -> bool {
        match self.accounts.get(account_address) {
            Some(hashed_key) => verify(private_key, hashed_key).unwrap_or(false),
            None => false,
        }
    }

    pub fn get_accounts(&self) -> Vec<HashMap<String, String>> {
        self.accounts.iter().map(|(account_address, hashed_key)| {
            let mut map = HashMap::new();
            map.insert("user_address".to_string(), account_address.clone());
            map.insert("private_key".to_string(), hashed_key.clone());
            map
        }).collect()
    }

    pub fn add_token(&mut self, token_id: &str) -> bool {
        if self.tokens.contains(token_id) {
            return false; // Token already exists
        }
        self.tokens.insert(token_id.to_string());
        if let Err(e) = self.save_to_file() {
            eprintln!("Failed to save to file: {}", e); // Log the error
            return false;
        }
        true
    }

    pub fn remove_token(&mut self, token_id: &str) -> bool {
        let removed = self.tokens.remove(token_id);
        if removed {
            if let Err(e) = self.save_to_file() {
                eprintln!("Failed to save to file: {}", e); // Log the error
                return false;
            }
        }
        removed
    }

    pub fn get_tokens(&self) -> HashSet<String> {
        self.tokens.clone()
    }

    pub fn remove_account(&mut self, account_address: &str) -> bool {
        let removed = self.accounts.remove(account_address).is_some();
        if removed {
            if let Err(e) = self.save_to_file() {
                eprintln!("Failed to save to file: {}", e); // Log the error
                return false;
            }
        }
        removed
    }
}