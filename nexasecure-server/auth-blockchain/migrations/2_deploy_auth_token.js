const AuthToken = artifacts.require("AuthToken");
const { JsonRpcProvider } = require('ethers');

const provider = new JsonRpcProvider('http://127.0.0.1:7545');
module.exports = async function (deployer) {
  const accounts = await provider.listAccounts();
  const serverAccount = accounts[0].address;
  const name = "AuthToken";
  const symbol = "ATKN";
  await deployer.deploy(AuthToken, serverAccount, name, symbol);
};

