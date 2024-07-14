const axios = require('axios');

async function signMessage(aurgToken, userAddress) {
  try {
    const response = await axios.post('http://localhost:3030/sign', {
      aurgToken,
      address: userAddress
    });
    return response.data.signature;
  } catch (error) {
    console.error('Error signing message with hardware wallet:', error);
    throw error;
  }
}

module.exports = {
  signMessage
};
