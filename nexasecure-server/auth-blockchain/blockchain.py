import time
import hashlib
from block import Block

class Blockchain:
    def __init__(self):
        self.chain = []
        self.pending_auth_data = []
        self.diffculty = 2
        self.create_genisis_block()

    def create_genisis_block(self):
        genisis_block = Block(0,'0',time.time(),'0')
        genisis_block.hash = genisis_block.generate_hash()
        self.chain.append(genisis_block)

    def get_last_block_hash(self):
        return self.chain[-1]
    
    def proof_of_work(self,block):
        block.nonce = 0
        computed_hash = block.generate_hash()
        while not computed_hash.startswith('0'*self.diffculty):
            block.nonce += 1
            computed_hash = block.generate_hash()
        return computed_hash
    
    def proof_of_stake(self,block,validator):
        stake = validator['stake']
        return hashlib.sha256(f'{block.hash}{stake}'.encode()).hexdigest()
    
    
    def new_auth_request(self,username,user_id):
        self.pending_auth_data.append({
            'username':username,
            'uuid':user_id
        })
        return self.get_last_block().index + 1


    def add_block(self,block,proof):
        previous_hash = self.get_last_block_hash().hash
        if previous_hash != block.previous_hash:
            return False
        if not self.is_valid_proof(block,proof):
            return False
        block.hash = proof
        self.chain.append(block)
        return True
    
    def is_valid_proof(self,block,proof):
        return proof.startswith('0'*self.diffculty) and proof == block.generate_hash()
    

    def add_auth_data(self,auth_data):
        self.pending_auth_data.append(auth_data)
    
    def mine(self,miner_details):
        if not self.pending_auth_data:
            return False
        last_block = self.get_last_block()
        new_block = Block(index=last_block.index + 1,
                          previous_hash=last_block.hash,
                          transactions=self.pending_transactions,
                          timestamp=time.time())

        if miner_details['type'] == 'PoW':
            proof = self.proof_of_work(new_block)
        else:  # PoS
            proof = self.proof_of_stake(new_block, miner_details)

        self.add_block(new_block, proof)
        self.pending_transactions = []
        return new_block
