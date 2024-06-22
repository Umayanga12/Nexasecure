// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract SimpleNFT is ERC721, Ownable {
    uint256 public nextTokenId;
    address public admin;

    constructor() ERC721("SimpleNFT", "SNFT") {
        admin = msg.sender;
    }

    function mint(address to) external onlyOwner {
        _safeMint(to, nextTokenId);
        nextTokenId++;
    }

    function transferFrom(address from, address to, uint256 tokenId) public override {
        require(msg.sender == admin, "only admin can transfer");
        _transfer(from, to, tokenId);
    }

    function changeAdmin(address newAdmin) external onlyOwner {
        admin = newAdmin;
    }
}
