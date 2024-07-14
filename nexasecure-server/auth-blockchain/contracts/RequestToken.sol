// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

contract RequestToken is ERC721 {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;
    address public serverAccount;

    constructor(address _serverAccount) ERC721("RequestToken", "RTKN") {
        serverAccount = _serverAccount;
    }

    function mintToken(address recipient) public returns (uint256) {
        require(msg.sender == serverAccount, "Only server can mint tokens");
        _tokenIds.increment();
        uint256 newTokenId = _tokenIds.current();
        _mint(recipient, newTokenId);
        return newTokenId;
    }

    function verifyTokenOwner(uint256 tokenId, address claimedOwner) public view returns (bool) {
        return ownerOf(tokenId) == claimedOwner;
    }

    function transferToServer(uint256 tokenId) public {
        require(ownerOf(tokenId) == msg.sender, "Only the owner can transfer the token");
        _transfer(msg.sender, serverAccount, tokenId);
    }

    function burnToken(uint256 tokenId) public {
        require(ownerOf(tokenId) == serverAccount, "Only the server can burn the token");
        _burn(tokenId);
    }
}
