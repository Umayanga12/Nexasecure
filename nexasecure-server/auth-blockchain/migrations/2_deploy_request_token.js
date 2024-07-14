const RequestToken = artifacts.require("RequestToken");
const { JsonRpcProvider } = require('ethers');

const provider = new JsonRpcProvider('http://127.0.0.1:8545');
module.exports = async function(deployer) {
  const name = "RequestToken";
  const symbol = "REQ";
  const accounts = await provider.listAccounts();
  const serverAccount = accounts[0].address;
  deployer.deploy(RequestToken,name,symbol,serverAccount);
};
