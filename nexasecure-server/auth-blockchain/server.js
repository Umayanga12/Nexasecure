require('dotenv').config();
const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const { ethers } = require('ethers');
const { JsonRpcProvider } = require('ethers');
const AuthTokenArtifact = require('./build/contracts/AuthToken.json');
const RequestTokenArtifact = require('./build/contracts/RequestToken.json');
const hardwareWallet = require('./hardwarewallet');

// Connect to both blockchains
const authProvider = new JsonRpcProvider(process.env.AUTH_BLOCKCHAIN_URI);
const requestProvider = new JsonRpcProvider(process.env.REQ_BLOCKCHAIN_URI);

// Set up wallets
const authServerWallet = new ethers.Wallet(process.env.AUTH_BLOCKCHAIN_SERVER_PRIVATE_KEY, authProvider);
const requestServerWallet = new ethers.Wallet(process.env.REQ_BLOCKCHAIN_SERVER_PRIVATE_KEY, requestProvider);

// Set up contracts
const AuthToken = new ethers.Contract(process.env.AUTH_BLOCKCHAIN_CONTRACT_ADDRESS ,AuthTokenArtifact.abi,authServerWallet);
const RequestToken = new ethers.Contract(process.env.REQ_BLOCKCHAIN_CONTRACT_ADDRESS, RequestTokenArtifact.abi, requestServerWallet);

app.use(bodyParser.json());

// Mint new auth token
async function mintAuthToken(userAddress) {
  const tx = await AuthToken.mintToken(userAddress);
  const receipt = await tx.wait();
  const logs = receipt.logs;
 // console.log(logs);
  for (let log of logs) {
   // console.log(log); 
    try {
      const parsedLog = AuthToken.interface.parseLog(log);
      if (parsedLog.name === "Transfer") {
        const tokenId = parsedLog.args[2].toString(); 
        console.log("Token ID is: ", tokenId);
        return tokenId;
       
      }
    } catch (error) {
      console.error("Error parsing log:", error);
    }
  }

  throw new Error("Failed to find mint event in receipt")
}

// Validate the auth token
async function verifyAuthToken(tokenId, userAddress) {
  const authTokenInstance = await AuthToken.deploy();
  await authTokenInstance.deployed();
  const isOwner = await authTokenInstance.verifyTokenOwner(tokenId, userAddress);
  return isOwner;
}

// Mint new request token
async function mintRequestToken(userAddress) {
  const tx = await RequestToken.mintToken(userAddress);
  const receipt = await tx.wait();
  const logs = receipt.logs;
  for (let log of logs) {
    try {
      const parsedLog = RequestToken.interface.parseLog(log);
      if (parsedLog.name === "Transfer") {
        const requestTokenId = parsedLog.args[2].toString();
        return requestTokenId;
      }
    } catch (error) {
      console.error("Error parsing log:", error);
    }
  }
}

// Validate the request token
async function verifyRequestToken(tokenId, userAddress) {
  const requestTokenInstance = await RequestToken.deploy();
  await requestTokenInstance.deployed();
  const isOwner = await requestTokenInstance.verifyTokenOwner(tokenId, userAddress);
  return isOwner;
}

// Change the ownership of the NFT
async function transferTokenToServer(instance, tokenId, userAddress) {
  await instance.transferToServer(tokenId, { from: userAddress });
}

// Terminate the request token
async function burnToken(instance, tokenId) {
  await instance.burnToken(tokenId, { from: requestServerWallet.address });
}

// Terminate the auth token
async function burnAuthToken(instance, tokenId) {
  await instance.burnToken(tokenId, { from: authServerWallet.address });
}

/* API ENDPOINTS */

// Sign-up endpoint - can mint a new token for the user
app.post('/sign', async (req, res) => {
  const { userAddress, signature } = req.body;
  const message = "Please sign this message to verify your identity.";

  try {
    // Verify the signature
    const recoveredAddress = ethers.utils.verifyMessage(message, signature);
    
    if (recoveredAddress.toLowerCase() === userAddress.toLowerCase()) {
      // User is valid, mint a new auth token
      const authTokenId = await mintAuthToken(userAddress);
      res.json({ authTokenId });
    } else {
      res.status(401).send('Invalid signature');
    }
  } catch (error) {
    res.status(500).send('Error verifying signature or minting token');
  }
});

// // Login API endpoint
app.post('/login', async (req, res) => {
  const { userAddress, authTokenId } = req.body;
  const message = `Login request for auth token ${authTokenId}`;

  try {
    const signature = await hardwareWallet.signMessage(message, userAddress);

    // Validate the user's ownership of the token
    const recoveredAddress = ethers.utils.verifyMessage(message, signature);
    if (recoveredAddress.toLowerCase() === userAddress.toLowerCase()) {
      const isOwner = await verifyAuthToken(authTokenId, userAddress);
      if (isOwner) {
        // Transfer ownership to the server
        const authInstance = await AuthToken.deploy();
        await authInstance.deployed();
        await transferTokenToServer(authInstance, authTokenId, userAddress);
        
        // Mint a request token for the user
        const requestTokenId = await mintRequestToken(userAddress);

        // Burn the auth token
        await burnAuthToken(authInstance, authTokenId);

        res.json({ authTokenId, requestTokenId, message, signature, ownerTransferred: true });
      } else {
        res.status(400).json({ error: 'Ownership verification failed' });
      }
    } else {
      res.status(400).json({ error: 'Signature verification failed' });
    }
  } catch (error) {
    res.status(500).send('Error signing message with hardware wallet');
  }
});

// Validate request token endpoint
app.post('/verify-request', async (req, res) => {
  const { requestTokenId, userAddress, signature } = req.body;
  const message = `Request token ${requestTokenId}`;

  try {
    const recoveredAddress = ethers.utils.verifyMessage(message, signature);
    console.log("TOken validation called");
    if (recoveredAddress.toLowerCase() === userAddress.toLowerCase()) {
      const isOwner = await verifyRequestToken(requestTokenId, userAddress);
      res.json({ isOwner });
    } else {
      res.json({ isOwner: false });
    }
  } catch (error) {
    res.status(500).send('Error verifying request token');
  }
});

// Logout endpoint
app.post('/logout', async (req, res) => {
  const { userAddress, requestTokenId } = req.body;
  const message = `Logout request for request token ${requestTokenId}`;

  try {
    const signature = await hardwareWallet.signMessage(message, userAddress);

    // Validate the user's ownership of the request token
    const recoveredAddress = ethers.utils.verifyMessage(message, signature);
    if (recoveredAddress.toLowerCase() === userAddress.toLowerCase()) {
      const isOwner = await verifyRequestToken(requestTokenId, userAddress);
      if (isOwner) {
        // Transfer ownership of the request token to the server
        const requestInstance = await RequestToken.deploy();
        await requestInstance.deployed();
        await transferTokenToServer(requestInstance, requestTokenId, userAddress);

        // Mint a new auth token for the user
        const authTokenId = await mintAuthToken(userAddress);

        // Burn the request token
        await burnToken(requestInstance, requestTokenId);
        console.log("Logout endpoint called - TOken burned");
        res.json({ authTokenId, message, signature, ownerTransferred: true });
      } else {
        res.status(400).json({ error: 'Ownership verification failed' });
      }
    } else {
      res.status(400).json({ error: 'Signature verification failed' });
    }
  } catch (error) {
    res.status(500).send('Error signing message with hardware wallet');
  }
});

//Register endpoint -- for development and testing purposes only
app.post('/register', async (req, res) => {
  const { userAddress } = req.body;

  if (!userAddress) {
    res.status(400).json({ error: "Provide the user address" });
    return;
  }

  try {
    const newUserAuthToken = await mintAuthToken(userAddress);
    res.status(200).json({ newUserAuthToken });
    console.log("Registration endpoint called");
  } catch (error) {
    res.status(500).json({ error: 'Error minting authentication token', details: error.message });
  }
});

// Server will work on port 3020
app.listen(process.env.SERVER_PORT, () => {
  console.log('Auth server running on port ', process.env.SERVER_PORT);
});
