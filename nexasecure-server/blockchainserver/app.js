const express = require('express');
const bodyParser = require('body-parser');
const ethers  = require('ethers');
const { Pool } = require('pg');
const { JsonRpcProvider } = require('ethers');
require('dotenv').config();
const app = express();
const port = process.env.PORT;
app.use(bodyParser.json());
const pool = new Pool({
  user: process.env.DB_USER,
  host: process.env.DB_HOST,
  database: process.env.DATABASE,
  password: process.env.DB_PASSWORD,
  port: process.env.DB_PORT,
});

const provider = new JsonRpcProvider(process.env.REQ_BLOCKCHAIN_URL);

async function initializeAccounts() {
  const accounts = await provider.listAccounts();
  for (let i = 2; i < accounts.length; i++) {
    await pool.query(
      'INSERT INTO accounts (address, status) VALUES ($1, $2) ON CONFLICT (address) DO NOTHING',
      [accounts[i], 'available']
    );
  }
}

app.get('/accounts', async (req, res) => {
  const result = await pool.query('SELECT * FROM accounts');
  res.json(result.rows);
});

app.post('/allocate-account', async (req, res) => {
  const result = await pool.query(
    "UPDATE accounts SET status = 'used' WHERE address = (SELECT address FROM accounts WHERE status = 'available' LIMIT 1) RETURNING address"
  );
  if (result.rows.length > 0) {
    res.json({ allocatedAccount: result.rows[0].address });
  } else {
    res.status(400).json({ error: 'No available accounts' });
  }
});

app.post('/release-account', async (req, res) => {
  const { address } = req.body;
  const wallet = ethers.Wallet.fromMnemonic(ethers.Wallet.createRandom().mnemonic.phrase, `m/44'/60'/0'/0/0`);
  await pool.query(
    "UPDATE accounts SET status = 'burned' WHERE address = $1",
    [address]
  );
  await pool.query(
    'INSERT INTO accounts (address, status) VALUES ($1, $2)',
    [wallet.address, 'available']
  );
  res.json({ releasedAccount: address, newAccount: wallet.address });
});

app.listen(port, async () => {
  await initializeAccounts();
  console.log(`Server running on ${port}`);
});