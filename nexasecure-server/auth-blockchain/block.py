import hashlib
import json
import time

class Block:
    def __init__(self,index,previous_hash,timestamp,auth_data,nonce=0) -> None:
        self.index = index
        self.previous_hash = previous_hash
        self.timestamp = timestamp
        self.auth_data = auth_data
        self.nonce = nonce

    def generate_hash(self):
        block_string = json.dumps(self.__dict__,sort_keys=True)
        return hashlib.sha256(block_string.encode()).hexdigest()
    
