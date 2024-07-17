# NFT-Based Authentication System for Micro services- Demo

![diagrem 1 drawio](https://github.com/user-attachments/assets/137ef741-1bf6-4587-9084-46ba5779c735)

## Overview

This application demonstrates an innovative NFT-based authentication system designed for enhanced security and efficient user management. The system comprises four main components:

1. **Backend Authentication Server**: Manages user authentication and integrates with the blockchain servers.
2. **2FA Server**: Provides two-factor authentication to ensure secure access.
3. **Blockchain Server**: Connects to two private blockchains; one for authentication purposes and another for handling user requests.
4. **Hardware Wallet**: Contains an in-memory database to store request tokens and a hardware wallet to store authentication tokens.

**Note: This application is for demonstration purposes only.**

## System Features

- **VPN Setup**: Ensures encrypted communication between the user and the organization.
- **Proxy Server**: Configures a proxy server, disables forward proxies, and adds a reverse proxy as a load balancer.
- **SSL Encryption**: Secures communication between internal services and the database.

## Installation Instructions

### Rust Program

#### Prerequisites:
- Install Rust on your computer.

#### Steps:
1. Open the project using a code editor or IDE.
2. Run the command `cargo run` in the terminal.

### Python Program

#### Prerequisites:
- Install Python on your computer.

#### Steps:
1. Open the project using a code editor or IDE.
2. Run the command `pip install -r requirements.txt` in the terminal.
3. Run the command `python main.py` in the terminal.

### Go (Golang) Program

#### Prerequisites:
- Install Golang on your computer.

#### Steps:
1. Open the project using a code editor or IDE.
2. Run the commands `go mod init` and `go mod tidy` in the terminal.

### Node Project

#### Prerequisites:
- Install Node.js, Bun.js, or any JavaScript runtime on your computer.

#### Steps:
1. Open the project using a code editor or IDE.
2. Install dependencies.
3. Run the project.

### Blockchain Setup

#### Prerequisites:
- Install a locally runnable blockchain (Ganache is used in this project).

#### Steps:
1. Set up two blockchains to run on two different ports.
2. Deploy the smart contracts on the blockchains.

## Post-Setup Instructions

1. Start the blockchain and deploy the smart contracts.
2. Set server accounts and private keys for the Blockchain server and smart contract address (Node server).
3. Add an account to the hardware wallet.
4. Set up the PostgreSQL database.
5. Configure the database connection on the Authentication server.
6. Start the wallet and in-memory database.
7. Start the blockchain server.
8. Start the 2FA server.
9. Start the main authentication server.
10. Send API calls using any HTTP client.

## Conclusion

By following these steps, you can set up and run this demonstration of an NFT-based authentication system. This system showcases the potential of leveraging NFTs for secure and efficient authentication processes, with dynamic ownership and robust security practices.
