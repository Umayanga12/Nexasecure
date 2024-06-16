use serde::{Serialize, Deserialize};
use std::sync::{Arc, Mutex};
use tokio::time::{self, Duration};
use std::fs::File;
use std::io::Write;

#[derive(Serialize, Deserialize, Debug)]
struct Request {
    id: u32,
    timestamp: u64,
    data: String,
}

struct InMemoryDB {
    requests: Vec<Request>,
    next_id: u32,
}

impl InMemoryDB {
    fn new() -> Self {
        InMemoryDB {
            requests: Vec::new(),
            next_id: 1,
        }
    }

    fn add_request(&mut self, data: String) {
        let req = Request {
            id: self.next_id,
            timestamp: chrono::Utc::now().timestamp() as u64,
            data,
        };
        self.requests.push(req);
        self.next_id += 1;
    }

    fn snapshot_and_clear(&mut self, filename: &str) -> std::io::Result<()> {
        // Take a snapshot
        let file = File::create(filename)?;
        let writer = std::io::BufWriter::new(file);
        serde_json::to_writer(writer, &self.requests)?;

        // Clear the database
        self.requests.clear();
        self.next_id = 1;

        Ok(())
    }
}

#[tokio::main]
async fn main() {
    let db = Arc::new(Mutex::new(InMemoryDB::new()));
    let snapshot_interval = Duration::from_secs(10);

    // Simulate incoming requests
    {
        let db = db.clone();
        tokio::spawn(async move {
            for i in 0..100 {
                let mut db = db.lock().unwrap();
                db.add_request(format!("Request data {}", i));
                time::sleep(Duration::from_millis(100)).await;
            }
        });
    }

    // Periodically take a snapshot and clear the database
    let db_clone = db.clone();
    tokio::spawn(async move {
        let mut interval = time::interval(snapshot_interval);
        loop {
            interval.tick().await;
            let filename = format!("snapshot-{}.json", chrono::Utc::now().timestamp());
            let mut db = db_clone.lock().unwrap();
            if let Err(e) = db.snapshot_and_clear(&filename) {
                println!("Error taking snapshot: {:?}", e);
            } else {
                println!("Snapshot taken and database cleared: {}", filename);
            }
        }
    }).await.unwrap();
}
