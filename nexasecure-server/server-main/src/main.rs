use tokio::process::Command;
use std::process::Stdio;

async fun main(){
    let authserver = Command::new("cargo")
        .arg("run")
        .current_dir("../../authserver")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Authserver is not responding");

    let facialrecognition = Command::new("cargo")
        .arg("run")
        .current_dir("../../facialrecognition")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Facialrecognition is not responding");

    let requestblockchain = Command::new("cargo")
        .arg("run")
        .current_dir("../../requestblockchain")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Requestblockchain is not responding");

    let authblockchain = Command::new("cargo")
        .arg("run")
        .current_dir("../../authblockchain")
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .expect("Authblockchain is not responding");

    let _ = tokio::join!(authserver, facialrecognition, requestblockchain, authblockchain).unwrap();
}