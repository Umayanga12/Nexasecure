use tokio::process::Command;
use std::process::Stdio;

async fn main(){
    let wallet = Command::new("cargo")
        .arg("run")
        .current_dir("../../wallet")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Wallet is not responding");

    let memorydb = Command::new("cargo")
        .arg("run")
        .current_dir("../../memory")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Error : Memory is not initiationg");


    let _ = tokio::join!(wallet, memorydb).unwrap();
}